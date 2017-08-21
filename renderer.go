package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	ThemesPath     string = "themes"
	ThemesFilePath string = "themes/%s/%s"
	ThemeConfig    string = "theme.json"
	DefaultTheme   string = "peaches"
)

var (
	Layouts = map[PageType]string{
		PageRegular:  "page.html",
		PageHome:     "home.html",
		PageProject:  "project.html",
		PageContact:  "contact.html",
		PageNotFound: "notfound.html",
	}

	Includes = []string{
		"head.html",
		"header.html",
		"footer.html",
		"menu.html",
	}
)

type Renderer struct {
	configuration Configuration
	templates     map[string]*template.Template
	shortUrl      string
}

func NewRenderer() *Renderer {
	return &Renderer{}
}

func (r *Renderer) Configure(configuration Configuration) error {
	r.configuration = configuration

	var s string
	{
		if strings.HasSuffix(r.configuration.Meta.Site, "/") {
			s = strings.TrimSuffix(r.configuration.Meta.Site, "/")
		} else {
			s = r.configuration.Meta.Site
		}
	}

	r.shortUrl = s

	err := r.LoadTheme()
	if err != nil {
		return err
	}

	return nil
}

func (r *Renderer) LoadTheme() error {
	templates, err := loadTheme(r.configuration.CurrentThemePath, r)

	if err != nil {
		templates, err = loadTheme(DefaultTheme, r)
		if err != nil {
			return err
		}
	}

	r.templates = templates

	return nil
}

func loadTheme(path string, r *Renderer) (map[string]*template.Template, error) {
	if !filesExist(path) {
		return nil, ErrInvalidTheme
	}

	m := make(map[string]*template.Template)

	inc := make([]string, 0, len(Includes))
	for _, include := range Includes {
		inc = append(inc, fmt.Sprintf(ThemesFilePath, path, include))
	}

	f := getFuncMap(r)

	for _, file := range Layouts {
		files := append(inc, fmt.Sprintf(ThemesFilePath, path, file))

		t, err := template.New(file).Funcs(f).ParseFiles(files...)
		if err != nil {
			return m, err
		}

		m[file] = t
	}

	return m, nil
}

func filesExist(path string) bool {
	for _, v := range Layouts {
		if _, err := os.Stat(fmt.Sprintf(ThemesFilePath, path, v)); os.IsNotExist(err) {
			return false
		}
	}
	for _, v := range Includes {
		if _, err := os.Stat(fmt.Sprintf(ThemesFilePath, path, v)); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func (r *Renderer) Render(b *bytes.Buffer, p *Page, ro string) error {
	if r.templates == nil {
		return ErrNoTheme
	}

	var (
		o = sortMenuKeys(p.Menu)
		t = strings.Split(Layouts[p.Type], ".")
		d = map[string]interface{}{
			"Title": p.Title,
			"Menu":  p.Menu,
			"Order": o,
			"Meta":  p.Meta,
			"User":  p.User,
			"Css":   r.configuration.CurrentTheme.Css,
			"Js":    r.configuration.CurrentTheme.Js,
			"active": func(s string) bool {
				var l string
				{
					if s == "home" {
						l = "/"
					} else {
						l = filepath.Join("/", "page", s)
					}

				}
				return ro == l
			},
		}
	)

	switch p.Type {
	case PageHome:
		d["Projects"] = p.Content.([]Project)
	case PageRegular:
		d["Content"] = p.Content.(Content)
	case PageProject:
		d["Project"] = p.Content.(Project)
	}

	return r.templates[Layouts[p.Type]].ExecuteTemplate(b, t[0], d)
}

func getFuncMap(r *Renderer) template.FuncMap {
	return template.FuncMap{
		"html": func(s string) template.HTML {
			return template.HTML(s)
		},
		"now": time.Now,
		"resource": func(s string) string {
			return filepath.Join("/", ThemesPath, r.configuration.CurrentThemePath, s)
		},
		"full": func(s string) string {
			return r.configuration.Meta.Site + s
		},
		"project": func(s string) string {
			return filepath.Join("/", "project", s)
		},
		"route": func(s string) string {
			if s == "home" {
				return "/"
			}

			return filepath.Join("/", "page", s)
		},
		"timehour": func(t time.Time) string {
			return t.Format("2006-01-02 15:04")
		},
		"timedate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
		"media": func(m Media) template.HTML {
			var r string
			{
				switch m.Type {
				case MediaImage:
					r = fmt.Sprintf(`<img src="/%s" alt="%s" />`, m.Path, m.Name)
				case MediaVideo:
					r = fmt.Sprintf(`<video controls>
									   <source src="/%s" type="%s">
									   Your browser does not support the video tag.
									 </video>`, m.Path, m.Mime)
				}
			}

			return template.HTML(r)
		},
		"social": func(k, v string) template.HTML {
			var r string
			{
				switch k {
				case "facebook":
					r = "fa-facebook-square"
				case "twitter":
					r = "fa-twitter-square"
				case "behance":
					r = "fa-behance-square"
				case "linkedin":
					r = "fa-linkedin-square"
				case "soundcloud":
					r = "fa-soundcloud"
				case "youtube":
					r = "fa-youtube-square"
				case "deviantart":
					r = "fa-deviantart"
				case "gplus":
					r = "fa-google-plus-square"
				case "medium":
					r = "fa-medium"
				case "quora":
					r = "fa-quora"
				case "stack-overflow":
					r = "fa-stack-overflow"
				case "tumblr":
					r = "fa-tumblr-square"
				case "github":
					r = "fa-github-square"
				default:
					r = "fa-link"
				}
			}

			r = fmt.Sprintf(`<div class="social">
			                    <div class="social__icon">
			                        <i class="fa fa-lg %s" aria-hidden="true"></i>
			                    </div>
			                    <div class="social__link">
			                        <a href="%s">%s</a>
			                    </div>
			                </div>`, r, v, v)

			return template.HTML(r)
		},
		"dark": func(p ProjectStyle) bool {
			return p == StyleDark
		},
	}
}

func sortMenuKeys(m Menu) []int {
	var k []int = make([]int, 0, len(m))
	for v := range m {
		k = append(k, v)
	}
	sort.Ints(k)
	return k
}
