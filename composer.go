package main

import (
	"bytes"
	"context"
	"sync"

	"go.uber.org/zap"
)

type Composer struct {
	db            DB
	configuration Configuration
	logger        *zap.Logger
}

func NewComposer(db DB, logger *zap.Logger) *Composer {
	return &Composer{
		db:     db,
		logger: logger,
	}
}

func (c *Composer) Configure(configuration Configuration) error {
	c.configuration = configuration
	return nil
}

func (c *Composer) getMeta() Meta {
	return c.configuration.Meta
}

func (c *Composer) getPageElements() (User, Menu, Meta, map[string]Route) {
	var (
		wg sync.WaitGroup
		c0 chan User             = make(chan User)
		c1 chan Menu             = make(chan Menu)
		c2 chan Meta             = make(chan Meta)
		c3 chan map[string]Route = make(chan map[string]Route)
	)

	wg.Add(4)

	go func() {
		defer wg.Done()

		u, err := c.db.GetUser(context.TODO())
		if err != nil {
			c0 <- User{}
			return
		}

		c0 <- u
	}()

	go func() {
		defer wg.Done()

		m, err := c.db.GetMenu(context.TODO())
		if err != nil {
			c1 <- Menu{}
			return
		}

		c1 <- m
	}()

	go func() {
		defer wg.Done()

		c2 <- c.getMeta()
	}()

	go func() {
		defer wg.Done()

		r, err := c.db.GetRoutes(context.TODO())
		if err != nil {
			c3 <- map[string]Route{}
			return
		}

		c3 <- r
	}()

	go func() {
		wg.Wait()

		close(c0)
		close(c1)
		close(c2)
		close(c3)
	}()

	return <-c0, <-c1, <-c2, <-c3
}

func (c *Composer) getUser() User {
	u, err := c.db.GetUser(context.TODO())
	if err != nil {
		return User{}
	}

	return u
}

func (c *Composer) GetHomePage() Page {
	var (
		u, m, t, r = c.getPageElements()
		ctx        = context.TODO()
	)

	ps, err := c.db.GetProjects(ctx)
	if err != nil {
		return Page{}
	}

	return Page{
		Title: r["home"].Title,

		Type: PageHome,

		User: u,
		Meta: t,
		Menu: m,

		Content: ps,
	}
}

func (c *Composer) GetContactPage() Page {
	var (
		u, m, t, r = c.getPageElements()
		ctx        = context.TODO()
	)

	u, err := c.db.GetUser(ctx)
	if err != nil {
		return Page{}
	}

	buildPageMeta(&t, r["contact"].Title)

	return Page{
		Title: r["contact"].Title,

		Type: PageContact,

		User: u,
		Meta: t,
		Menu: m,

		Content: u,
	}
}

func (c *Composer) GetNotFoundPage() Page {
	var (
		t = c.getMeta()
	)

	buildPageMeta(&t, "Not found")

	return Page{
		Title: "Not found",

		Type: PageNotFound,

		Meta: t,

		Content: nil,
	}
}

func (c *Composer) GetProject(slug string) Page {
	var (
		u, m, t, _ = c.getPageElements()
		ctx        = context.TODO()
	)

	project, err := c.db.GetProject(ctx, slug)
	if err != nil {
		return Page{
			Title: "Not found",

			Type: PageNotFound,

			Meta: t,

			Content: nil,
		}
	}

	buildProjectMeta(&t, &project)

	return Page{
		Title: project.Title,

		Type: PageProject,

		User: u,
		Meta: t,
		Menu: m,

		Content: project,
	}
}

func (c *Composer) GetPage(slug string) Page {
	var (
		u, m, t, r = c.getPageElements()
		ctx        = context.TODO()
	)

	var (
		e  Route
		ok bool
	)

	if e, ok = r[slug]; !ok {
		return Page{
			Title: "Not found",

			Type: PageNotFound,

			Meta: t,

			Content: nil,
		}
	}

	content, err := c.db.GetContent(ctx, e.Slug)
	if err != nil {
		return Page{}
	}

	buildContentMeta(&t, &content)

	return Page{
		Title: e.Title,

		Type: PageRegular,

		User: u,
		Meta: t,
		Menu: m,

		Content: content,
	}
}

func buildPageMeta(m *Meta, t string) {
	title := t + " | " + m.Title

	m.Title = title
	m.OGTags["title"] = title
}

func buildProjectMeta(m *Meta, p *Project) {
	var t bytes.Buffer

	if len(p.Tags) > 0 {
		for _, v := range p.Tags {
			t.WriteString(string(v))
			t.WriteString(", ")
		}
	}

	if len(p.Technologies) > 0 {
		for _, v := range p.Technologies {
			t.WriteString(string(v))
			t.WriteString(", ")
		}
	}

	s := t.String()
	if len(s) > 2 {
		s = s[:(len(s) - 2)]
	}

	title := p.Title + " | " + m.Title

	m.Title = title

	m.Tags["description"] = p.Subtitle
	m.Tags["keywords"] = s

	m.OGTags["title"] = title
	m.OGTags["type"] = "article"
	m.OGTags["url"] = ""
	m.OGTags["image"] = ""
}

func buildContentMeta(m *Meta, c *Content) {
	var t bytes.Buffer

	if len(c.Tags) > 0 {
		for _, v := range c.Tags {
			t.WriteString(string(v))
			t.WriteString(", ")
		}
	}

	if len(c.Technologies) > 0 {
		for _, v := range c.Technologies {
			t.WriteString(string(v))
			t.WriteString(", ")
		}
	}

	s := t.String()
	if len(s) > 2 {
		s = s[:(len(s) - 2)]
	}

	title := c.Title + " | " + m.Title

	m.Title = title

	m.Tags["description"] = c.Subtitle
	m.Tags["keywords"] = s

	m.OGTags["title"] = title
	m.OGTags["type"] = "article"
	m.OGTags["url"] = ""
	m.OGTags["image"] = ""
}
