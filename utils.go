package main

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	XMLns           string        = "http://www.sitemaps.org/schemas/sitemap/0.9"
	SitemapFreq     string        = "weekly"
	SitemapFile     string        = "http/sitemap.xml"
	SitemapInterval time.Duration = time.Hour
)

var (
	chars = []string{"]", "^", "\\\\", "[", ".", "(", ")", "!", "@", "#", "$", "%", "*", "_", "+", "=", "{", "}", ";", ":", "'", "\"", "<", ">", "?", "/", ",", "`", "~"}
	c     = strings.Join(chars, "")
	re    = regexp.MustCompile("[" + c + "]+")
)

// Functions
func GenerateSlug(s string) string {
	var r string
	{
		r = strings.TrimSpace(s)
		r = strings.ToLower(r)
		r = re.ReplaceAllString(r, "")
		r = strings.Replace(r, " ", "-", -1)
	}
	return r
}

func GenerateRandomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%X", b)
}

func IsMapValid(m map[string]string) bool {
	for _, v := range m {
		if v == "" {
			return false
		}
	}

	return true
}

// Types
type Configurator struct {
	c  []Configurable
	cf Configuration
}

func NewConfigurator(args ...Configurable) *Configurator {
	return &Configurator{
		c: args,
	}
}

func (c *Configurator) Append(args ...Configurable) {
	c.c = append(c.c, args...)
}

func (c *Configurator) Configure(cf Configuration) (<-chan struct{}, <-chan error) {
	c.cf = cf

	var (
		wg sync.WaitGroup
		c0 chan struct{} = make(chan struct{})
		c1 chan error    = make(chan error, 2)
	)

	wg.Add(len(c.c))

	for _, v := range c.c {
		go func(c Configurable, f Configuration) {
			defer wg.Done()

			if e := c.Configure(cf); e != nil {
				c1 <- e
			}
		}(v, cf)
	}

	go func() {
		wg.Wait()

		c0 <- struct{}{}
		close(c0)
	}()

	return c0, c1
}

func (c *Configurator) GetConfiguration() Configuration {
	return c.cf
}

type Finalizer struct {
	f []Finalizable
}

func NewFinalizer(args ...Finalizable) *Finalizer {
	return &Finalizer{
		f: args,
	}
}

func (f *Finalizer) Append(args ...Finalizable) {
	f.f = append(f.f, args...)
}

func (f *Finalizer) Finalize() {
	for _, v := range f.f {
		v.Finalize()
	}
}

type BufferPool struct {
	c chan *bytes.Buffer
	a int
}

func NewBufferPool(size int, alloc int) (bp *BufferPool) {
	return &BufferPool{
		c: make(chan *bytes.Buffer, size),
		a: alloc,
	}
}

func (bp *BufferPool) Get() (b *bytes.Buffer) {
	select {
	case b = <-bp.c:
	default:
		b = bytes.NewBuffer(make([]byte, 0, bp.a))
	}
	return
}

func (bp *BufferPool) Put(b *bytes.Buffer) {
	b.Reset()

	if cap(b.Bytes()) > bp.a {
		b = bytes.NewBuffer(make([]byte, 0, bp.a))
	}

	select {
	case bp.c <- b:
	default: // Discard the buffer if the pool is full.
	}
}

type SitemapBuilder struct {
	db       DB
	logger   *zap.Logger
	interval time.Duration
	stop     chan bool
	url      string
}

func NewSitemapBuiler(db DB, logger *zap.Logger, interval time.Duration) *SitemapBuilder {
	return &SitemapBuilder{
		db:       db,
		logger:   logger,
		interval: interval,
		stop:     make(chan bool),
	}
}

func (s *SitemapBuilder) Configure(configuration Configuration) error {
	s.url = configuration.Meta.Site

	return nil
}

func (s *SitemapBuilder) Run() {
	s.Build()

	ticker := time.NewTicker(s.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.Build()
			case <-s.stop:
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *SitemapBuilder) Build() {
	var ctx = context.TODO()

	u, err := url.Parse(s.url)
	if err != nil {
		s.logger.Error("cannot parse url to update sitemap", zap.Error(err))
		return
	}

	r, err := s.db.GetRoutes(context.TODO())
	if err != nil {
		s.logger.Error("cannot get routes to update sitemap", zap.Error(err))
	}

	p, err := s.db.GetProjects(ctx)
	if err != nil {
		s.logger.Error("cannot get projects to update sitemap", zap.Error(err))
	}

	var urls []Url = make([]Url, 0, len(r)+len(p))

	for _, v := range r {
		var a string
		if v.Slug == "home" {
			a = u.String()
		} else {
			a = u.String() + path.Join("page", v.Slug)
		}

		urls = append(urls, Url{
			Loc:        a,
			LastMod:    time.Now().Format("2006-01-02"),
			ChangeFreq: SitemapFreq,
		})
	}

	for _, v := range p {
		urls = append(urls, Url{
			Loc:        u.String() + path.Join("project", v.Slug),
			LastMod:    v.Published.Format("2006-01-02"),
			ChangeFreq: SitemapFreq,
		})
	}

	set := UrlSet{
		XMLns: XMLns,
		Urls:  urls,
	}

	output, err := xml.MarshalIndent(set, "  ", "    ")
	if err != nil {
		s.logger.Error("cannot save output to update sitemap", zap.Error(err))
		return
	}

	ioutil.WriteFile(SitemapFile, output, 0644)
}

func (s *SitemapBuilder) Finalize() {
	s.stop <- true
}

type ThemeScanner struct{}

func NewThemeScanner() *ThemeScanner {
	return &ThemeScanner{}
}

func (s *ThemeScanner) LoadThemes() map[string]Theme {
	ts := make(map[string]Theme)

	dirs, _ := ioutil.ReadDir(ThemesPath)
	for _, d := range dirs {
		data, err := ioutil.ReadFile(filepath.Join(ThemesPath, d.Name(), ThemeConfig))
		if err != nil {
			continue
		}

		var t Theme

		err = json.Unmarshal(data, &t)
		if err != nil {
			continue
		}

		ts[d.Name()] = t
	}

	return ts
}
