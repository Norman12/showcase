package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/kataras/go-fs"

	"go.uber.org/zap"

	"golang.org/x/crypto/bcrypt"
)

const (
	ErrorFile        string = "http/application/error.html"
	NotAvailableFile string = "http/application/not-available.html"
	SetupFile        string = "http/application/setup.html"
	SuccessFile      string = "http/application/success.html"
	FaviconFile      string = "favicon.ico"

	ConfigurationTimeoutInterval time.Duration = 5 * time.Second
)

type SwappableServeMux struct {
	sync.RWMutex
	m *http.ServeMux
}

func NewSwappableServeMux(r *http.ServeMux) (s *SwappableServeMux) {
	return &SwappableServeMux{
		m: r,
	}
}

func (s *SwappableServeMux) Swap(r *http.ServeMux) {
	s.Lock()
	s.m = r
	s.Unlock()
}

func (s *SwappableServeMux) Handle(p string, h http.Handler) {
	s.RLock()
	t := s.m
	s.RUnlock()

	t.Handle(p, h)
}

func (s *SwappableServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.RLock()
	t := s.m
	s.RUnlock()

	t.ServeHTTP(w, r)
}

type NotFoundHandler struct {
	s *Server
}

func (n *NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var p Page = n.s.c.GetNotFoundPage()

	b := n.s.bp.Get()
	defer n.s.bp.Put(b)

	err := n.s.r.Render(b, &p, r.URL.Path)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		serveError(w, r)

		n.s.l.Error("cannot render template", zap.Error(err))

		return
	}

	w.WriteHeader(http.StatusOK)
	b.WriteTo(w)
}

type RouteHandler struct {
	Method  string
	Handler func(*Server) func(http.ResponseWriter, *http.Request)
}

var (
	routes = map[string]RouteHandler{
		"/": RouteHandler{
			Method:  "GET",
			Handler: homeHandler,
		},
		"/sitemap.xml": RouteHandler{
			Method:  "GET",
			Handler: sitemapHandler,
		},
		"/robots.txt": RouteHandler{
			Method:  "GET",
			Handler: robotsHandler,
		},
		"/favicon.ico": RouteHandler{
			Method:  "GET",
			Handler: faviconHandler,
		},
		"/success": RouteHandler{
			Method:  "GET",
			Handler: successHandler,
		},
		"/project/{slug:[0-9a-zA\\-]+}": RouteHandler{
			Method:  "GET",
			Handler: projectHandler,
		},
		"/page/{slug:[0-9a-zA\\-]+}": RouteHandler{
			Method:  "GET",
			Handler: pageHandler,
		},
	}

	setupRoutes = map[string]RouteHandler{
		"/setup": RouteHandler{
			Method:  "GET",
			Handler: setupHandler,
		},
	}

	adminRoutes_Public = map[string]RouteHandler{
		"/admin/login": RouteHandler{
			Method:  "POST",
			Handler: loginHandler,
		},
	}

	adminRoutes_Protected = map[string]RouteHandler{
		"/admin/projects": RouteHandler{
			Method:  "GET",
			Handler: getProjectsHandler,
		},
		"/admin/contents": RouteHandler{
			Method:  "GET",
			Handler: getContentsHandler,
		},

		"/admin/user": RouteHandler{
			Method:  "GET",
			Handler: getUserHandler,
		},
		"/admin/theme": RouteHandler{
			Method:  "GET",
			Handler: getThemeHandler,
		},
		"/admin/meta": RouteHandler{
			Method:  "GET",
			Handler: getMetaHandler,
		},
		"/admin/credentials": RouteHandler{
			Method:  "GET",
			Handler: getCredentialsHandler,
		},

		"/admin/user/update": RouteHandler{
			Method:  "PUT",
			Handler: updateUserHandler,
		},
		"/admin/theme/update": RouteHandler{
			Method:  "PUT",
			Handler: updateThemeHandler,
		},
		"/admin/meta/update": RouteHandler{
			Method:  "PUT",
			Handler: updateMetaHandler,
		},
		"/admin/credentials/update": RouteHandler{
			Method:  "PUT",
			Handler: updateCredentialsHandler,
		},

		"/admin/project/{slug}": RouteHandler{
			Method:  "GET",
			Handler: getProjectHandler,
		},
		"/admin/project/create": RouteHandler{
			Method:  "POST",
			Handler: createProjectHandler,
		},
		"/admin/project/{slug}/update": RouteHandler{
			Method:  "PUT",
			Handler: updateProjectHandler,
		},
		"/admin/project/{slug}/delete": RouteHandler{
			Method:  "DELETE",
			Handler: deleteProjectHandler,
		},

		"/admin/content/{slug}": RouteHandler{
			Method:  "GET",
			Handler: getContentHandler,
		},
		"/admin/content/create": RouteHandler{
			Method:  "POST",
			Handler: createContentHandler,
		},
		"/admin/content/{slug}/update": RouteHandler{
			Method:  "PUT",
			Handler: updateContentHandler,
		},
		"/admin/content/{slug}/delete": RouteHandler{
			Method:  "DELETE",
			Handler: deleteContentHandler,
		},

		"/admin/menu": RouteHandler{
			Method:  "GET",
			Handler: getMenuHandler,
		},
		"/admin/menu/add": RouteHandler{
			Method:  "PUT",
			Handler: addToMenuHandler,
		},
		"/admin/menu/remove": RouteHandler{
			Method:  "PUT",
			Handler: deleteFromMenuHandler,
		},

		"/admin/site": RouteHandler{
			Method:  "GET",
			Handler: siteHandler,
		},

		"/admin/logout": RouteHandler{
			Method:  "POST",
			Handler: logoutHandler,
		},
	}
)

type Server struct {
	db  DB
	ca  Cache
	co  *Configurator
	c   *Composer
	r   *Renderer
	m   *MediaManager
	l   *zap.Logger
	bp  *BufferPool
	gzp *fs.GzipPool
}

func NewServer(co *Configurator, db DB, ca Cache, c *Composer, r *Renderer, m *MediaManager, l *zap.Logger) *Server {
	return &Server{
		db:  db,
		ca:  ca,
		co:  co,
		c:   c,
		r:   r,
		m:   m,
		l:   l,
		bp:  NewBufferPool(32, 1024),
		gzp: fs.NewGzipPool(6),
	}
}

func (s *Server) NewRouter() http.Handler {
	r := mux.NewRouter()

	var mfs http.Handler
	{
		mfs = http.FileServer(http.Dir("media"))
		mfs = NewBrowserCacheMiddleware(s.ca)(mfs)
		mfs = NewFileGzipMiddleware(s.gzp)(mfs)
	}

	var tfs http.Handler
	{
		tfs = http.FileServer(http.Dir("themes"))
		tfs = NewBrowserCacheMiddleware(s.ca)(tfs)
		tfs = NewFileGzipMiddleware(s.gzp)(tfs)
	}

	var dfs http.Handler
	{
		dfs = http.FileServer(http.Dir("data"))
		dfs = NewBrowserCacheMiddleware(s.ca)(dfs)
		dfs = NewFileGzipMiddleware(s.gzp)(dfs)
	}

	var afs http.Handler
	{
		afs = http.FileServer(http.Dir("admin/dist/assets"))
		afs = NewBrowserCacheMiddleware(s.ca)(afs)
		afs = NewFileGzipMiddleware(s.gzp)(afs)
	}

	r.PathPrefix("/media").Handler(http.StripPrefix("/media/", mfs)).Methods("GET")
	r.PathPrefix("/themes").Handler(http.StripPrefix("/themes/", tfs)).Methods("GET")
	r.PathPrefix("/data").Handler(http.StripPrefix("/data/", dfs)).Methods("GET")
	r.PathPrefix("/assets").Handler(http.StripPrefix("/assets/", afs)).Methods("GET")

	for p, f := range routes {
		var h HandleFunc
		{
			h = f.Handler(s)
			h = NewLoggingMiddleware(s.l)(h)
			h = NewGzipMiddleware(s.gzp)(h)
		}
		r.HandleFunc(p, h).Methods(f.Method)
	}

	r.NotFoundHandler = &NotFoundHandler{
		s: s,
	}

	return r
}

func (s *Server) NewAdminRouter() http.Handler {
	var (
		r  = mux.NewRouter()
		sf = s.provideSigningFunc()
	)

	// Serve Angular2 App
	var dfs http.Handler
	{
		dfs = http.FileServer(http.Dir("admin/dist"))
		dfs = NewFileGzipMiddleware(s.gzp)(dfs)
	}

	r.PathPrefix("/admin/dashboard").Handler(http.StripPrefix("/admin/dashboard", dfs)).Methods("GET")

	for p, f := range adminRoutes_Protected {
		var h HandleFunc
		{
			h = f.Handler(s)
			h = NewAuthorisationMiddleware(s.ca, sf)(h)
			h = NewLoggingMiddleware(s.l)(h)
			h = NewCorsMiddleware()(h)
			h = NewJsonMiddleware()(h)
			h = NewGzipMiddleware(s.gzp)(h)
		}
		r.HandleFunc(p, h).Methods(f.Method, "OPTIONS")
	}

	for p, f := range adminRoutes_Public {
		var h HandleFunc
		{
			h = f.Handler(s)
			h = NewLoggingMiddleware(s.l)(h)
			h = NewCorsMiddleware()(h)
			h = NewJsonMiddleware()(h)
			h = NewGzipMiddleware(s.gzp)(h)
		}
		r.HandleFunc(p, h).Methods(f.Method, "OPTIONS")
	}

	r.NotFoundHandler = &NotFoundHandler{
		s: s,
	}

	return r
}
func (s *Server) NewSetupRouter() (http.Handler, <-chan bool, chan<- bool) {
	var (
		r = mux.NewRouter()
		d = make(chan bool, 2)
		n = make(chan bool)
	)

	for p, f := range setupRoutes {
		var h HandleFunc
		{
			h = f.Handler(s)
			h = NewLoggingMiddleware(s.l)(h)
			h = NewGzipMiddleware(s.gzp)(h)
		}
		r.HandleFunc(p, h).Methods(f.Method)
	}

	var finish HandleFunc
	{
		finish = func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()

			var (
				ctx = context.TODO()
				m   = map[string]string{
					"site":     r.FormValue("site"),
					"title":    r.FormValue("title"),
					"email":    r.FormValue("email"),
					"password": r.FormValue("password"),
				}
			)

			if !IsMapValid(m) {
				redirectToSetup(w, r, ErrSetupEmpty)
				return
			}

			_, err := url.Parse(m["site"])
			if err != nil {
				redirectToSetup(w, r, err)
				return
			}

			if !strings.HasSuffix(m["site"], "/") {
				m["site"] = m["site"] + "/"
			}

			hash, err := bcrypt.GenerateFromPassword([]byte(m["password"]), bcrypt.DefaultCost)
			if err != nil {
				redirectToSetup(w, r, err)
				return
			}

			c, err := s.db.GetCredentials(ctx)
			if err != nil {
				redirectToSetup(w, r, err)
				return
			}

			c.Email = m["email"]
			c.Hash = string(hash)

			err = s.db.PutCredentials(&c)
			if err != nil {
				redirectToSetup(w, r, err)
				return
			}

			co, err := s.db.GetConfiguration(ctx)
			if err != nil {
				redirectToSetup(w, r, err)
				return
			}

			co.Meta.Site = m["site"]
			co.Meta.Title = m["title"]

			co.SetupCompleted = true

			err = s.db.PutConfiguration(&co)
			if err != nil {
				redirectToSetup(w, r, err)
				return
			}

			var (
				e      error
				ch     = make(chan bool)
				c0, c1 = s.co.Configure(co)
			)
			go func(err error) {
				for {
					select {
					case <-time.After(ConfigurationTimeoutInterval):
						err = ErrConfigurationTimedOut
						ch <- true
					case <-c0:
						ch <- true
					case e := <-c1:
						err = e
						ch <- true
					case <-ch:
						return
					}
				}
			}(e)

			<-ch

			if e != nil {
				redirectToSetup(w, r, e)
				return
			}

			d <- true

			<-n

			redirectToSuccess(w, r)
		}
		finish = NewLoggingMiddleware(s.l)(finish)
		finish = NewGzipMiddleware(s.gzp)(finish)
	}
	r.HandleFunc("/setup/finish", finish).Methods("POST")

	r.NotFoundHandler = &NotFoundHandler{
		s: s,
	}

	return r, d, n
}

// Public routes
func homeHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		s.pushResources(w, s.co.GetConfiguration())

		var p Page = s.c.GetHomePage()

		b := s.bp.Get()
		defer s.bp.Put(b)

		err := s.r.Render(b, &p, r.URL.Path)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			serveError(w, r)

			s.l.Error("cannot render template", zap.Error(err))

			return
		}

		w.WriteHeader(http.StatusOK)
		b.WriteTo(w)
	}
}

func projectHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			vars = mux.Vars(r)
			slug = vars["slug"]
		)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		s.pushResources(w, s.co.GetConfiguration())

		var p Page = s.c.GetProject(slug)

		b := s.bp.Get()
		defer s.bp.Put(b)

		err := s.r.Render(b, &p, r.URL.Path)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			serveError(w, r)

			s.l.Error("cannot render template", zap.Error(err))

			return
		}

		w.WriteHeader(http.StatusOK)
		b.WriteTo(w)
	}
}

func pageHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			vars = mux.Vars(r)
			slug = vars["slug"]
		)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		s.pushResources(w, s.co.GetConfiguration())

		var p Page
		{
			switch slug {
			case "contact":
				p = s.c.GetContactPage()
			default:
				p = s.c.GetPage(slug)
			}
		}

		b := s.bp.Get()
		defer s.bp.Put(b)

		err := s.r.Render(b, &p, r.URL.Path)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			serveError(w, r)

			s.l.Error("cannot render template", zap.Error(err))

			return
		}

		w.WriteHeader(http.StatusOK)
		b.WriteTo(w)
	}
}

func sitemapHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, SitemapFile)
	}
}

func robotsHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		u, err := url.Parse(s.co.GetConfiguration().Meta.Site)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			serveError(w, r)

			s.l.Error("cannot write robots.txt", zap.Error(err))

			return
		}

		fmt.Fprintf(w, "User-agent: *\nAllow: /\nSitemap: %s", u.String()+"sitemap.xml")
	}
}

func faviconHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(ThemesPath, s.co.GetConfiguration().CurrentThemePath, FaviconFile))
	}
}

func successHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, SuccessFile)
	}
}

//Admin routes
func getProjectsHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var ctx = context.TODO()

		ps, err := s.db.GetProjects(ctx)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		var ps_ = make([]Project_, 0, len(ps))
		for _, p := range ps {
			var i string
			{
				if p.Image.Path != "" {
					i = "/" + p.Image.Path
				} else {
					i = ""
				}
			}

			ps_ = append(ps_, Project_{
				Slug: p.Slug,

				Title:    p.Title,
				Subtitle: p.Subtitle,
				Image:    i,
			})
		}

		writeResponse(w, ps_, nil)
	}
}

func getContentsHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var ctx = context.TODO()

		cs, err := s.db.GetContents(ctx)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		var cs_ = make([]Content_, 0, len(cs))
		for _, c := range cs {
			cs_ = append(cs_, Content_{
				Slug: c.Slug,

				Title:    c.Title,
				Subtitle: c.Subtitle,
			})
		}

		writeResponse(w, cs_, nil)
	}
}

func getUserHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var ctx = context.TODO()

		u, err := s.db.GetUser(ctx)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		req := UpdateUserRequest{
			Name:  u.Name,
			Title: u.Title,
			About: u.About,
			Image: []Media_{
				Media_{
					Name:     u.Image.Name,
					Caption:  u.Image.Caption,
					Resource: u.Image.Path,
				},
			},
			Logo: []Media_{
				Media_{
					Name:     u.Logo.Name,
					Caption:  u.Logo.Caption,
					Resource: u.Logo.Path,
				},
			},
			References:  u.References,
			Networks:    u.Networks,
			Experiences: u.Experiences,
			Interests:   u.Interests,
			Contact:     u.Contact,
		}

		writeResponse(w, req, nil)
	}
}

func getThemeHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			scanner = NewThemeScanner()
			themes  = scanner.LoadThemes()
			themes_ = make([]Theme_, 0, len(themes))
		)

		if len(themes) == 0 {
			writeResponse(w, nil, ErrNoThemes)
			return
		}

		for p, t := range themes {
			themes_ = append(themes_, Theme_{
				Name:   t.Name,
				Author: t.Author,
				Path:   p,
				Image:  s.co.GetConfiguration().Meta.Site + filepath.Join(ThemesPath, p, t.Image),
			})
		}

		req := struct {
			Selected string   `json:"selected"`
			Themes   []Theme_ `json:"themes"`
		}{
			Selected: s.co.GetConfiguration().CurrentThemePath,
			Themes:   themes_,
		}

		writeResponse(w, req, nil)
	}
}

func getMetaHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c := s.co.GetConfiguration()

		req := UpdateMetaRequest{
			Site:   c.Meta.Site,
			Title:  c.Meta.Title,
			Tags:   c.Meta.Tags,
			OGTags: c.Meta.OGTags,
		}

		writeResponse(w, req, nil)
	}
}

func getCredentialsHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var ctx = context.TODO()

		c, err := s.db.GetCredentials(ctx)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		req := UpdateCredentialsRequest{
			Email: c.Email,
		}

		writeResponse(w, req, nil)
	}
}

func updateUserHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var ctx = context.TODO()

		var req UpdateUserRequest
		if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
			writeResponse(w, nil, e)
			return
		}

		u, err := s.db.GetUser(ctx)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		if req.Image[0].Removed {
			s.m.Delete(&u.Image)
		} else {
			if len(req.Image[0].File.Data) > 0 {
				m, err := s.m.Save(&req.Image[0].File)
				if err == nil {
					u.Image = *m
				}
			}

			u.Image.Name = req.Image[0].Name
			u.Image.Caption = req.Image[0].Caption
		}

		if req.Logo[0].Removed {
			s.m.Delete(&u.Logo)
		} else {
			if len(req.Logo[0].File.Data) > 0 {
				m, err := s.m.Save(&req.Logo[0].File)
				if err == nil {
					u.Logo = *m
				}
			}

			u.Logo.Name = req.Logo[0].Name
			u.Logo.Caption = req.Logo[0].Caption
		}

		u.Name = req.Name
		u.Title = req.Title
		u.About = req.About
		u.References = req.References
		u.Networks = req.Networks
		u.Experiences = req.Experiences
		u.Interests = req.Interests
		u.Contact = req.Contact

		err = s.db.PutUser(&u)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		writeResponse(w, true, nil)
	}
}

func updateThemeHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var ctx = context.TODO()

		var req UpdateThemeRequest
		if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
			writeResponse(w, nil, e)
			return
		}

		c, err := s.db.GetConfiguration(ctx)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		var (
			scanner = NewThemeScanner()
			themes  = scanner.LoadThemes()
		)

		if len(themes) == 0 {
			writeResponse(w, nil, ErrNoThemes)
			return
		}

		if _, ok := themes[req.Path]; !ok {
			writeResponse(w, nil, ErrInvalidTheme)
			return
		}

		c.CurrentTheme = themes[req.Path]
		c.CurrentThemePath = req.Path

		err = s.db.PutConfiguration(&c)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		var (
			e      error
			ch     = make(chan bool)
			c0, c1 = s.co.Configure(c)
		)
		go func(err error) {
			for {
				select {
				case <-time.After(ConfigurationTimeoutInterval):
					err = ErrConfigurationTimedOut
					ch <- true
				case <-c0:
					ch <- true
				case e := <-c1:
					err = e
					ch <- true
				case <-ch:
					return
				}
			}
		}(e)

		<-ch

		if e != nil {
			writeResponse(w, nil, e)
			return
		}

		writeResponse(w, true, nil)
	}
}

func updateMetaHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var ctx = context.TODO()

		var req UpdateMetaRequest
		if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
			writeResponse(w, nil, e)
			return
		}

		c, err := s.db.GetConfiguration(ctx)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		_, err = url.Parse(req.Site)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		if !strings.HasSuffix(req.Site, "/") {
			req.Site = req.Site + "/"
		}

		c.Meta.Site = req.Site
		c.Meta.Title = req.Title
		c.Meta.Tags = req.Tags
		c.Meta.OGTags = req.OGTags

		err = s.db.PutConfiguration(&c)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		var (
			e      error
			ch     = make(chan bool)
			c0, c1 = s.co.Configure(c)
		)
		go func(err error) {
			for {
				select {
				case <-time.After(ConfigurationTimeoutInterval):
					err = ErrConfigurationTimedOut
					ch <- true
				case <-c0:
					ch <- true
				case e := <-c1:
					err = e
					ch <- true
				case <-ch:
					return
				}
			}
		}(e)

		<-ch

		if e != nil {
			writeResponse(w, nil, e)
			return
		}

		writeResponse(w, true, nil)
	}
}

func updateCredentialsHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var ctx = context.TODO()

		var req UpdateCredentialsRequest
		if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
			writeResponse(w, nil, e)
			return
		}

		c, err := s.db.GetCredentials(ctx)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		c.Email = req.Email
		c.Hash = string(hash)

		err = s.db.PutCredentials(&c)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		writeResponse(w, true, nil)
	}
}

func getProjectHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx  = context.TODO()
			vars = mux.Vars(r)
			slug = vars["slug"]
		)

		p, err := s.db.GetProject(ctx, slug)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		req := UpdateProjectRequest{
			Slug: p.Slug,

			Title:        p.Title,
			Subtitle:     p.Subtitle,
			About:        p.About,
			Tags:         p.Tags,
			Technologies: p.Technologies,
			References:   p.References,
			Image: []Media_{
				Media_{
					Name:     p.Image.Name,
					Caption:  p.Image.Caption,
					Resource: "image",
				},
			},
			Logo: []Media_{
				Media_{
					Name:     p.Logo.Name,
					Caption:  p.Logo.Caption,
					Resource: "logo",
				},
			},
			Client: struct {
				Name  string   `json:"name"`
				About string   `json:"about"`
				Image []Media_ `json:"image"`
			}{
				Name:  p.Client.Name,
				About: p.Client.About,
				Image: []Media_{
					Media_{
						Name:     p.Client.Image.Name,
						Caption:  p.Client.Image.Caption,
						Resource: p.Client.Image.Path,
					},
				},
			},
		}

		var media = make([]Media_, 0, len(p.Images))
		for _, m := range p.Images {
			media = append(media, Media_{
				Name:     m.Name,
				Caption:  m.Caption,
				Resource: m.Path,
				Uploaded: true,
			})
		}

		req.Media = media

		writeResponse(w, req, nil)
	}
}

func createProjectHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateProjectRequest
		if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
			writeResponse(w, nil, e)
			return
		}

		p := Project{
			Slug: GenerateSlug(req.Title),

			Title:    req.Title,
			Subtitle: req.Subtitle,
			About:    req.About,

			Published: time.Now(),

			Tags:         req.Tags,
			Technologies: req.Technologies,
			References:   req.References,

			Client: Client{
				Name:  req.Client.Name,
				About: req.Client.About,
			},

			Imported: Imported{},
		}

		err := s.db.CreateProject(&p)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		if req.Image[0].Removed {
			p.Image = Media{}
		} else {
			if len(req.Image[0].File.Data) > 0 {
				m, err := s.m.Save(&req.Image[0].File)
				if err == nil {
					m.Name = req.Image[0].Name
					m.Caption = req.Image[0].Caption

					p.Image = *m
				}
			}
		}

		if req.Logo[0].Removed {
			p.Logo = Media{}
		} else {
			if len(req.Logo[0].File.Data) > 0 {
				m, err := s.m.Save(&req.Logo[0].File)
				if err == nil {
					m.Name = req.Logo[0].Name
					m.Caption = req.Logo[0].Caption

					p.Logo = *m
				}
			}
		}

		if req.Client.Image[0].Removed {
			p.Client.Image = Media{}
		} else {
			if len(req.Client.Image[0].File.Data) > 0 {
				me, err := s.m.Save(&req.Client.Image[0].File)
				if err == nil {
					me.Name = req.Client.Image[0].Name
					me.Caption = req.Client.Image[0].Caption

					p.Client.Image = *me
				}
			}
		}

		media := make([]Media, 0)
		for _, m := range req.Media {
			if len(m.File.Data) > 0 {
				me, err := s.m.Save(&m.File)
				if err == nil {
					me.Name = m.Name
					me.Caption = m.Caption

					media = append(media, *me)
				}
			}
		}

		p.Images = media

		err = s.db.PutProject(&p)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		writeResponse(w, true, nil)
	}
}

func updateProjectHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx  = context.TODO()
			vars = mux.Vars(r)
			slug = vars["slug"]
		)

		var req UpdateProjectRequest
		if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
			writeResponse(w, nil, e)
			return
		}

		p, err := s.db.GetProject(ctx, slug)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		p.Title = req.Title
		p.Subtitle = req.Subtitle
		p.About = req.About

		p.Tags = req.Tags
		p.Technologies = req.Technologies
		p.References = req.References

		p.Client.Name = req.Client.Name
		p.Client.About = req.Client.About

		p.Images = s.diffMedia(p.Images, req.Media)

		if req.Image[0].Removed {
			s.m.Delete(&p.Image)
		} else {
			if len(req.Image[0].File.Data) > 0 {
				s.m.Delete(&p.Image)

				me, err := s.m.Save(&req.Image[0].File)
				if err == nil {
					p.Image = *me
				}
			}

			p.Image.Name = req.Image[0].Name
			p.Image.Caption = req.Image[0].Caption
		}

		if req.Logo[0].Removed {
			s.m.Delete(&p.Logo)
		} else {
			if len(req.Logo[0].File.Data) > 0 {
				s.m.Delete(&p.Logo)

				me, err := s.m.Save(&req.Logo[0].File)
				if err == nil {
					p.Logo = *me
				}
			}

			p.Logo.Name = req.Logo[0].Name
			p.Logo.Caption = req.Logo[0].Caption
		}

		if req.Client.Image[0].Removed {
			s.m.Delete(&p.Client.Image)
		} else {
			if len(req.Client.Image[0].File.Data) > 0 {
				s.m.Delete(&p.Client.Image)

				me, err := s.m.Save(&req.Client.Image[0].File)
				if err == nil {
					p.Client.Image = *me
				}
			}

			p.Client.Image.Name = req.Client.Image[0].Name
			p.Client.Image.Caption = req.Client.Image[0].Caption
		}

		err = s.db.PutProject(&p)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		writeResponse(w, true, nil)
	}
}

func deleteProjectHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			vars = mux.Vars(r)
			slug = vars["slug"]
		)

		err := s.db.DeleteProject(slug)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		writeResponse(w, true, nil)
	}
}

func getContentHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx  = context.TODO()
			vars = mux.Vars(r)
			slug = vars["slug"]
		)

		c, err := s.db.GetContent(ctx, slug)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		var ps = make([]Paragraph_, 0, len(c.Paragraphs))
		for _, p := range c.Paragraphs {
			ps = append(ps, Paragraph_{
				Resource: p.Slug,
				Title:    p.Title,
				Content:  p.Content,
				Media: []Media_{
					Media_{
						Name:     p.Media.Name,
						Caption:  p.Media.Caption,
						Resource: p.Media.Path,
					},
				},
			})
		}

		req := UpdateContentRequest{
			Slug: c.Slug,

			Title:        c.Title,
			Subtitle:     c.Subtitle,
			Paragraphs:   ps,
			Tags:         c.Tags,
			Technologies: c.Technologies,
			References:   c.References,
		}

		writeResponse(w, req, nil)
	}
}

func createContentHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateContentRequest
		if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
			writeResponse(w, nil, e)
			return
		}

		c := Content{
			Slug: GenerateSlug(req.Title),

			Title:    req.Title,
			Subtitle: req.Subtitle,

			Published: time.Now(),

			Tags:         req.Tags,
			Technologies: req.Technologies,
			References:   req.References,
		}

		err := s.db.CreateContent(&c)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		paragraphs := make([]Paragraph, 0, len(req.Paragraphs))
		for _, v := range req.Paragraphs {
			p := Paragraph{
				Slug:    v.Resource,
				Title:   v.Title,
				Content: v.Content,
			}
			if v.Media[0].Removed {
				p.Media = Media{}
			} else {
				if len(v.Media[0].File.Data) > 0 {
					me, err := s.m.Save(&v.Media[0].File)
					if err == nil {
						me.Name = v.Media[0].Name
						me.Caption = v.Media[0].Caption

						p.Media = *me
					}
				}
			}

			paragraphs = append(paragraphs, p)
		}

		c.Paragraphs = paragraphs

		err = s.db.PutContent(&c)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		err = s.db.PutRoute(&Route{
			Title: c.Title,
			Slug:  c.Slug,
		})
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		writeResponse(w, true, nil)
	}
}

func updateContentHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx  = context.TODO()
			vars = mux.Vars(r)
			slug = vars["slug"]
		)

		var req UpdateContentRequest
		if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
			writeResponse(w, nil, e)
			return
		}

		c, err := s.db.GetContent(ctx, slug)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		c.Title = req.Title
		c.Subtitle = req.Subtitle

		c.Tags = req.Tags
		c.Technologies = req.Technologies
		c.References = req.References

		c.Paragraphs = s.diffParagraphs(c.Paragraphs, req.Paragraphs)

		err = s.db.PutContent(&c)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		writeResponse(w, true, nil)
	}
}

func deleteContentHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			vars = mux.Vars(r)
			slug = vars["slug"]
		)

		err := s.db.DeleteContent(slug)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		writeResponse(w, true, nil)
	}
}

func getMenuHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var ctx = context.TODO()

		rt, err := s.db.GetRoutes(ctx)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		m, err := s.db.GetMenu(ctx)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		var k []int = make([]int, 0, len(m))
		for v := range m {
			k = append(k, v)
		}
		sort.Ints(k)

		var (
			rt_ = make([]string, 0, len(rt)-len(m))
			m_  = make([]string, 0, len(m))
		)

		for _, r := range rt {
			var found bool
			for _, v := range m {
				if v.Slug == r.Slug {
					found = true
					break
				}
			}
			if !found {
				rt_ = append(rt_, r.Slug)
			}
		}
		for _, i := range k {
			m_ = append(m_, m[i].Slug)
		}

		req := Menu_{
			Routes: rt_,
			Added:  m_,
		}

		writeResponse(w, req, nil)
	}
}

func addToMenuHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var ctx = context.TODO()

		var req AddToMenuRequest
		if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
			writeResponse(w, nil, e)
			return
		}

		rt, err := s.db.GetRoutes(ctx)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		m, err := s.db.GetMenu(ctx)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		if _, ok := rt[req.Slug]; !ok {
			writeResponse(w, nil, err)
			return
		}

		var keys []int = make([]int, 0, len(m))
		for k, v := range m {
			if v.Slug == req.Slug {
				writeResponse(w, nil, err)
				return
			}
			keys = append(keys, k)
		}

		sort.Ints(keys)

		m[keys[(len(keys)-1)]+1] = rt[req.Slug]

		err = s.db.PutMenu(&m)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		writeResponse(w, true, nil)
	}
}

func deleteFromMenuHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var ctx = context.TODO()

		var req RemoveFromMenuRequest
		if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
			writeResponse(w, nil, e)
			return
		}

		m, err := s.db.GetMenu(ctx)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		var (
			i    int   = -1
			keys []int = make([]int, 0, len(m))
		)

		for k, v := range m {
			if v.Slug == req.Slug {
				i = k
			}
			keys = append(keys, k)
		}

		if i == -1 {
			writeResponse(w, nil, err)
			return
		}

		sort.Ints(keys)

		for k := 0; k < len(keys)-1; k++ {
			if k == i {
				m[k] = m[k+1]
				i = k + 1
			}
		}

		delete(m, i)

		err = s.db.PutMenu(&m)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		writeResponse(w, true, nil)
	}
}

func siteHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var ctx = context.TODO()

		c, err := s.db.GetConfiguration(ctx)
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		writeResponse(w, c.Meta.Site, nil)
	}
}

func loginHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
			writeResponse(w, nil, e)
			return
		}

		c, err := s.db.GetCredentials(context.TODO())
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		if c.Email != req.Email {
			http.Error(w, "incorrect email", http.StatusUnauthorized)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(c.Hash), []byte(req.Password))
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		u := uuid.New().String()

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"token": u,
		})

		tokenString, err := token.SignedString([]byte(s.co.GetConfiguration().JwtSecret))
		if err != nil {
			writeResponse(w, nil, err)
			return
		}

		s.ca.Set("token", u)

		writeResponse(w, tokenString, nil)
	}
}

func logoutHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s.ca.Delete("token")

		writeResponse(w, true, nil)
	}
}

func setupHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, SetupFile)
	}
}

func unavailableHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, NotAvailableFile)
	}
}

// Auxiliary
func (s *Server) pushResources(w http.ResponseWriter, c Configuration) {
	r := append(c.CurrentTheme.Css, c.CurrentTheme.Js...)

	if pusher, ok := w.(http.Pusher); ok {
		for _, f := range r {
			pusher.Push("/"+filepath.Join(ThemesPath, s.co.GetConfiguration().CurrentThemePath, f), nil)
		}
	}
}

func (s *Server) provideSigningFunc() func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(s.co.GetConfiguration().JwtSecret), nil
	}
}

func (s *Server) diffMedia(ms []Media, mms []Media_) []Media {
	var (
		oldMedia = make([]Media, 0, 0)
		newMedia = make([]Media, 0, 0)
	)
	for _, m := range ms {
		var (
			found bool
			req   Media_
		)
		for _, mm := range mms {
			if mm.Resource == m.Path {
				found = true
				req = mm
				break
			}
		}
		if found {
			m.Name = req.Name
			m.Caption = req.Caption

			oldMedia = append(oldMedia, m)
		}
	}
	for _, mm := range mms {
		var found bool
		for _, m := range ms {
			if m.Path == mm.Resource {
				found = true
				break
			}
		}
		if !found {
			if len(mm.File.Data) > 0 {
				me, err := s.m.Save(&mm.File)
				if err == nil {
					me.Name = mm.Name
					me.Caption = mm.Caption

					newMedia = append(newMedia, *me)
				}
			}
		}
	}

	return append(oldMedia, newMedia...)
}

func (s *Server) diffParagraphs(ps []Paragraph, pps []Paragraph_) []Paragraph {
	var (
		oldParagraphs = make([]Paragraph, 0, 0)
		newParagraphs = make([]Paragraph, 0, 0)
	)
	for _, p := range ps {
		var (
			found bool
			req   Paragraph_
		)
		for _, pp := range pps {
			if pp.Resource == p.Slug {
				found = true
				req = pp
				break
			}
		}
		if found {
			p.Title = req.Title
			p.Content = req.Content

			if len(req.Media[0].File.Data) > 0 {
				s.m.Delete(&p.Media)
				me, err := s.m.Save(&req.Media[0].File)
				if err == nil {
					p.Media = *me
				}
			}

			p.Media.Name = req.Media[0].Name
			p.Media.Caption = req.Media[0].Caption

			oldParagraphs = append(oldParagraphs, p)
		} else {
			s.m.Delete(&p.Media)
		}
	}
	for _, pp := range pps {
		var found bool
		for _, p := range ps {
			if p.Slug == pp.Resource {
				found = true
				break
			}
		}
		if !found {
			np := Paragraph{
				Slug:    pp.Resource,
				Title:   pp.Title,
				Content: pp.Content,
			}

			if pp.Media[0].Removed {
				np.Media = Media{}
			} else {
				if len(pp.Media[0].File.Data) > 0 {
					me, err := s.m.Save(&pp.Media[0].File)
					if err == nil {
						me.Name = pp.Media[0].Name
						me.Caption = pp.Media[0].Caption

						np.Media = *me
					}
				}
			}

			newParagraphs = append(newParagraphs, np)
		}
	}

	return append(oldParagraphs, newParagraphs...)
}

func serveError(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, ErrorFile)
}

func writeResponse(w http.ResponseWriter, v interface{}, e error) {
	r := GenericResponse{
		Content: v,
	}

	if e != nil {
		r.Error = e.Error()
	}

	buf, err := json.Marshal(r)
	if err != nil {
		writeResponse(w, nil, err)
		return
	}

	w.Write(buf)
}

func redirectToSuccess(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/success", 301)
}

func redirectToSetup(w http.ResponseWriter, r *http.Request, e error) {
	var s string
	{
		if e != nil {
			s = fmt.Sprintf("/setup?error=%s", e.Error())
		} else {
			s = "/setup"
		}
	}
	http.Redirect(w, r, s, 301)
}
