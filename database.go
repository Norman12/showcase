package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"

	"go.uber.org/zap"

	"golang.org/x/crypto/bcrypt"
)

const (
	BUCKET_COMMON   string = "common"
	BUCKET_PROJECTS string = "projects"
	BUCKET_CONTENT  string = "content"
	BUCKET_ROUTES   string = "routes"
	BUCKET_THEMES   string = "themes"
)

type DB interface {
	// Get
	GetUser(ctx context.Context) (User, error)
	GetRoutes(ctx context.Context) (map[string]Route, error)
	GetContents(ctx context.Context) ([]Content, error)
	GetContent(ctx context.Context, slug string) (Content, error)
	GetProjects(ctx context.Context) ([]Project, error)
	GetProject(ctx context.Context, slug string) (Project, error)
	GetMenu(ctx context.Context) (Menu, error)
	GetConfiguration(ctx context.Context) (Configuration, error)
	GetCredentials(ctx context.Context) (Credentials, error)

	//Create
	CreateContent(content *Content) error
	CreateProject(project *Project) error

	// Set
	PutUser(user *User) error
	PutContent(content *Content) error
	PutProject(project *Project) error
	PutMenu(menu *Menu) error
	PutCredentials(credentials *Credentials) error
	PutConfiguration(configutation *Configuration) error
	PutRoute(route *Route) error

	//Delete
	DeleteContent(slug string) error
	DeleteProject(slug string) error

	// Config
	Setup(context.Context) (Configuration, error)
	Mock(ctx context.Context) error
}

type cachedDatabase struct {
	cache  Cache
	bolt   *bolt.DB
	logger *zap.Logger
}

func NewCachedDatabase(bolt *bolt.DB, cache Cache, logger *zap.Logger) DB {
	return &cachedDatabase{
		cache:  cache,
		bolt:   bolt,
		logger: logger,
	}
}

func (db *cachedDatabase) GetUser(ctx context.Context) (User, error) {
	user, f := db.cache.Get("user")
	if f {
		return user.(User), nil
	}

	var u User
	err := db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_COMMON))
		v := b.Get([]byte("user"))

		err := json.Unmarshal(v, &u)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		db.logger.Error("cannot get user", zap.Error(err))
		return User{}, ErrDatabase
	}

	db.cache.Set("user", u)

	return u, nil
}

func (db *cachedDatabase) GetRoutes(ctx context.Context) (map[string]Route, error) {
	routes, f := db.cache.Get("routes")
	if f {
		return routes.(map[string]Route), nil
	}

	var r map[string]Route = make(map[string]Route)
	err := db.bolt.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(BUCKET_ROUTES)).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var o Route
			err := json.Unmarshal(v, &o)
			if err != nil {
				return err
			}
			r[o.Slug] = o
		}

		return nil
	})

	if err != nil {
		db.logger.Error("cannot get routes", zap.Error(err))
		return r, ErrDatabase
	}

	db.cache.Set("routes", r)

	return r, nil
}

func (db *cachedDatabase) GetProjects(ctx context.Context) ([]Project, error) {
	projects, f := db.cache.Get("projects")
	if f {
		return projects.([]Project), nil
	}

	var ps []Project
	err := db.bolt.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(BUCKET_PROJECTS)).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var p Project
			err := json.Unmarshal(v, &p)
			if err != nil {
				return err
			}
			ps = append(ps, p)
		}

		return nil
	})

	if err != nil {
		db.logger.Error("cannot get projects", zap.Error(err))
		return []Project{}, ErrDatabase
	}

	if len(ps) > 0 {
		db.cache.Set("projects", ps)
	}

	return ps, nil
}

func (db *cachedDatabase) GetProject(ctx context.Context, Slug string) (Project, error) {
	project, f := db.cache.Get(fmt.Sprintf("project-%s", Slug))
	if f {
		return project.(Project), nil
	}

	var p Project
	err := db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_PROJECTS))
		v := b.Get([]byte(Slug))

		err := json.Unmarshal(v, &p)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		db.logger.Error("cannot get project", zap.Error(err))
		return Project{}, ErrDatabase
	}

	db.cache.Set(fmt.Sprintf("project-%s", Slug), p)

	return p, nil
}

func (db *cachedDatabase) GetMenu(ctx context.Context) (Menu, error) {
	menu, f := db.cache.Get("menu")
	if f {
		return menu.(Menu), nil
	}

	var m Menu
	err := db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_COMMON))
		v := b.Get([]byte("menu"))

		err := json.Unmarshal(v, &m)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		db.logger.Error("cannot get menu", zap.Error(err))
		return Menu{}, ErrDatabase
	}

	db.cache.Set("menu", m)

	return m, nil
}

func (db *cachedDatabase) GetContents(ctx context.Context) ([]Content, error) {
	contents, f := db.cache.Get("contents")
	if f {
		return contents.([]Content), nil
	}

	var cs []Content
	err := db.bolt.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(BUCKET_CONTENT)).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var co Content
			err := json.Unmarshal(v, &co)
			if err != nil {
				return err
			}
			cs = append(cs, co)
		}

		return nil
	})

	if err != nil {
		db.logger.Error("cannot get contents", zap.Error(err))
		return []Content{}, ErrDatabase
	}

	if len(cs) > 0 {
		db.cache.Set("contents", cs)
	}

	return cs, nil
}

func (db *cachedDatabase) GetContent(ctx context.Context, Slug string) (Content, error) {
	content, f := db.cache.Get(fmt.Sprintf("content-%s", Slug))
	if f {
		return content.(Content), nil
	}

	var c Content
	err := db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_CONTENT))
		v := b.Get([]byte(Slug))

		err := json.Unmarshal(v, &c)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		db.logger.Error("cannot get content", zap.Error(err))
		return Content{}, ErrDatabase
	}

	db.cache.Set(fmt.Sprintf("content-%s", Slug), c)

	return c, nil
}

func (db *cachedDatabase) GetConfiguration(ctx context.Context) (Configuration, error) {
	configuration, f := db.cache.Get("configuration")
	if f {
		return configuration.(Configuration), nil
	}

	var c Configuration
	err := db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_COMMON))
		v := b.Get([]byte("configuration"))

		err := json.Unmarshal(v, &c)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		db.logger.Error("cannot get configuration", zap.Error(err))
		return Configuration{}, ErrDatabase
	}

	db.cache.Set("configuration", c)

	return c, nil
}

func (db *cachedDatabase) GetCredentials(ctx context.Context) (Credentials, error) {
	credentials, f := db.cache.Get("credentials")
	if f {
		return credentials.(Credentials), nil
	}

	var c Credentials
	err := db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_COMMON))
		v := b.Get([]byte("credentials"))

		err := json.Unmarshal(v, &c)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		db.logger.Error("cannot get credentials", zap.Error(err))
		return Credentials{}, ErrDatabase
	}

	db.cache.Set("credentials", c)

	return c, nil
}

func (db *cachedDatabase) CreateContent(content *Content) error {
	err := db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_COMMON))
		v := b.Get([]byte(content.Slug))

		if v != nil {
			return ErrContentExists
		}

		return nil
	})

	if err != nil {
		return err
	}

	return db.PutContent(content)
}

func (db *cachedDatabase) CreateProject(project *Project) error {
	err := db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_PROJECTS))
		v := b.Get([]byte(project.Slug))

		if v != nil {
			return ErrProjectExists
		}

		return nil
	})

	if err != nil {
		return err
	}

	return db.PutProject(project)
}

func (db *cachedDatabase) PutUser(user *User) error {
	err := db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_COMMON))
		return save(b, []byte("user"), user)
	})

	if err != nil {
		return err
	}

	db.cache.Set("user", *user)

	return nil
}

func (db *cachedDatabase) PutContent(content *Content) error {
	if content.Slug == "" {
		return ErrNoSlug
	}

	err := db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_CONTENT))
		return save(b, []byte(content.Slug), content)
	})

	if err != nil {
		return err
	}

	m, err := db.GetMenu(context.TODO())
	if err != nil {
		return err
	}

	for i, r := range m {
		if r.Slug == content.Slug {
			m[i] = Route{Title: content.Title, Slug: content.Slug}
			break
		}
	}

	err = db.PutMenu(&m)
	if err != nil {
		return err
	}

	db.cache.Delete("routes")
	db.cache.Delete("contents")
	db.cache.Set(fmt.Sprintf("content-%s", content.Slug), *content)

	return nil
}

func (db *cachedDatabase) PutProject(project *Project) error {
	if project.Slug == "" {
		return ErrNoSlug
	}

	err := db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_PROJECTS))
		return save(b, []byte(project.Slug), project)
	})

	if err != nil {
		return err
	}

	m, err := db.GetMenu(context.TODO())
	if err != nil {
		return err
	}

	for i, r := range m {
		if r.Slug == project.Slug {
			m[i] = Route{Title: project.Title, Slug: project.Slug}
			break
		}
	}

	err = db.PutMenu(&m)
	if err != nil {
		return err
	}

	db.cache.Delete("routes")
	db.cache.Delete("projects")
	db.cache.Set(fmt.Sprintf("project-%s", project.Slug), *project)

	return nil
}

func (db *cachedDatabase) PutMenu(menu *Menu) error {
	err := db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_COMMON))
		return save(b, []byte("menu"), menu)
	})

	if err != nil {
		return err
	}

	db.cache.Set("menu", *menu)

	return nil
}

func (db *cachedDatabase) PutCredentials(credentials *Credentials) error {
	err := db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_COMMON))
		return save(b, []byte("credentials"), credentials)
	})

	if err != nil {
		return err
	}

	db.cache.Set("credentials", *credentials)

	return nil
}

func (db *cachedDatabase) PutConfiguration(configutation *Configuration) error {
	err := db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_COMMON))
		return save(b, []byte("configuration"), configutation)
	})

	if err != nil {
		return err
	}

	db.cache.Set("configuration", *configutation)

	return nil
}

func (db *cachedDatabase) PutRoute(route *Route) error {
	err := db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_ROUTES))
		return save(b, []byte(route.Slug), route)
	})

	if err != nil {
		return err
	}

	db.cache.Delete("routes")

	return nil
}

func (db *cachedDatabase) DeleteProject(slug string) error {
	err := db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_PROJECTS))
		return b.Delete([]byte(slug))
	})

	if err != nil {
		return err
	}

	db.cache.Delete(fmt.Sprintf("project-%s", slug))
	db.cache.Delete("projects")
	db.cache.Delete("routes")

	return nil
}

func (db *cachedDatabase) DeleteContent(slug string) error {
	err := db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_CONTENT))
		return b.Delete([]byte(slug))
	})

	if err != nil {
		return err
	}

	db.cache.Delete(fmt.Sprintf("content-%s", slug))
	db.cache.Delete("contents")
	db.cache.Delete("routes")

	return nil
}

func (db *cachedDatabase) Setup(ctx context.Context) (Configuration, error) {
	db.bolt.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(BUCKET_COMMON))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte(BUCKET_PROJECTS))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte(BUCKET_CONTENT))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte(BUCKET_THEMES))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte(BUCKET_ROUTES))
		if err != nil {
			return err
		}

		return nil
	})

	var c Configuration
	err := db.bolt.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_THEMES))

		var (
			scanner = NewThemeScanner()
			themes  = scanner.LoadThemes()
		)

		if len(themes) == 0 {
			return ErrNoThemes
		}

		var err error
		for k, v := range themes {
			err = save(b, []byte(k), v)
			if err != nil {
				return err
			}
		}

		b = tx.Bucket([]byte(BUCKET_COMMON))
		v := b.Get([]byte("configuration"))

		if v != nil {
			err := json.Unmarshal(v, &c)
			if err != nil {
				return err
			}
			return nil
		}

		var (
			t  string
			ok bool
		)
		if _, ok = themes["peaches"]; ok {
			t = "peaches"
		} else {
			k, _ := firstKey(themes)
			t = string(k.(string))
		}

		u := uuid.New().String()

		if u == "" {
			return errors.New("cannot generate uuid for jwt")
		}

		c = Configuration{
			CurrentThemePath: t,
			CurrentTheme:     themes[t],
			SetupCompleted:   false,
			JwtSecret:        u,
			Meta: Meta{
				Title: "",
				Site:  "",
				Tags: map[string]string{
					"description": "",
					"keywords":    "",
					"author":      "",
					"viewport":    "width=device-width, initial-scale=1.0",
				},
				OGTags: map[string]string{
					"title": "",
					"type":  "website",
					"url":   "",
					"image": "",
				},
			},
		}

		err = save(b, []byte("configuration"), c)
		if err != nil {
			return err
		}

		user := User{
			Name:        "User",
			Title:       "Title",
			About:       "",
			Image:       Media{},
			Logo:        Media{},
			Joined:      time.Now(),
			References:  Map{},
			Networks:    Map{},
			Experiences: Map{},
			Interests:   []Interest{},
			Contact: Contact{
				Country: "",
				City:    "",
				Street:  "",
				Email:   "",
				Phone:   "",
			},
		}

		err = save(b, []byte("user"), user)
		if err != nil {
			return err
		}

		credentials := Credentials{}

		err = save(b, []byte("credentials"), credentials)
		if err != nil {
			return err
		}

		b = tx.Bucket([]byte(BUCKET_ROUTES))

		r1 := Route{
			Title: "Home",
			Slug:  "home",
		}

		save(b, []byte(r1.Slug), r1)

		r2 := Route{
			Title: "Contact",
			Slug:  "contact",
		}

		save(b, []byte(r2.Slug), r2)

		r3 := Route{
			Title: "Not Found",
			Slug:  "notfound",
		}

		save(b, []byte(r3.Slug), r3)

		b = tx.Bucket([]byte(BUCKET_COMMON))

		menu := Menu{
			0: r1,
			1: r2,
		}

		err = save(b, []byte("menu"), menu)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		db.logger.Error("cannot create configuration", zap.Error(err))
		return Configuration{}, ErrDatabase
	}

	return c, nil
}

func (db *cachedDatabase) Mock(ctx context.Context) error {
	return db.bolt.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_COMMON))

		user := User{
			Name:  "Marcin Praski",
			Title: "Developer",
			About: "Some bio",
			Image: Media{},
			Logo: Media{
				Type: MediaImage,
				Name: "Logo",
				Path: "media/images/l.jpg",
			},
			Joined: time.Now(),
			References: Map{
				"UCL Student": "",
			},
			Networks: Map{
				"Facebook": "https://www.facebook.com/marcin.praski.5",
			},
			Experiences: Map{
				"Some hackathon": "",
			},
			Interests: []Interest{
				"Golang", "Java", "PostgreSQL",
			},
			Contact: Contact{
				Country: "UK",
				City:    "London",
				Street:  "109 Camden Road",
				Email:   "me@marcinpraski.com",
				Phone:   "123123123",
			},
		}

		err := save(b, []byte("user"), user)
		if err != nil {
			return err
		}

		credentials := Credentials{
			Email: "marcin.praski@live.com",
		}

		hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		credentials.Hash = string(hash)

		err = save(b, []byte("credentials"), credentials)
		if err != nil {
			return err
		}

		b2 := tx.Bucket([]byte(BUCKET_PROJECTS))

		project1 := Project{
			Slug:     "project-1",
			Title:    "Showcase",
			Subtitle: "Portfolio generator",
			About:    "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
			Image: Media{
				Type: MediaImage,
				Name: "Some image",
				Path: "media/images/m.jpg",
			},
			Logo:      Media{},
			Published: time.Now(),
			Images: []Media{
				Media{
					Type:    MediaImage,
					Name:    "Some image",
					Caption: "Some caption",
					Path:    "media/images/m.jpg",
				},
				Media{
					Type:    MediaImage,
					Name:    "Some image 2",
					Caption: "Some caption 2",
					Path:    "media/images/m2.jpg",
				},
				Media{
					Type:    MediaVideo,
					Mime:    "video/mp4",
					Name:    "Some video 3",
					Caption: "Some caption 3",
					Path:    "media/videos/fuck.mp4",
				},
			},
			Tags: []Tag{
				"Portfolio", "Static", "Generator",
			},
			Technologies: []Technology{
				"Golang", "Html", "Boltdb",
			},
			References: Map{},
			Client: Client{
				Name:  "Some Client",
				About: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
				Image: Media{
					Type: MediaImage,
					Name: "Logo",
					Path: "media/images/l.jpg",
				},
			},
			Imported: Imported{},
		}

		save(b2, []byte(project1.Slug), project1)

		project2 := Project{
			Slug:      "project-2",
			Title:     "Showcase 2",
			Subtitle:  "Portfolio generator",
			About:     "Lorem ipsum",
			Image:     Media{},
			Logo:      Media{},
			Published: time.Now(),
			Images:    []Media{},
			Tags: []Tag{
				"portfolio", "static", "generator",
			},
			Technologies: []Technology{
				"golang", "html", "boltdb",
			},
			References: Map{},
			Client: Client{
				Name:  "Some Client",
				About: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
			},
			Imported: Imported{},
		}

		save(b2, []byte(project2.Slug), project2)

		b3 := tx.Bucket([]byte(BUCKET_CONTENT))

		content1 := Content{
			Slug:      "content-1",
			Title:     "Some content",
			Subtitle:  "No important",
			Published: time.Now(),
			Paragraphs: []Paragraph{
				Paragraph{
					Slug:    "paragraph-12529349235",
					Title:   "Some title",
					Content: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
					Media: Media{
						Type:    MediaImage,
						Name:    "Some image 2",
						Caption: "Some caption 2",
						Path:    "media/images/m.jpg",
					},
				},
			},
			Tags: []Tag{
				"portfolio", "static", "generator",
			},
			Technologies: []Technology{
				"golang", "html", "boltdb",
			},
			References: Map{},
		}

		save(b3, []byte(content1.Slug), content1)

		b4 := tx.Bucket([]byte(BUCKET_ROUTES))

		r1 := Route{
			Title: "Home",
			Slug:  "home",
		}

		save(b4, []byte(r1.Slug), r1)

		r2 := Route{
			Title: "Contact",
			Slug:  "contact",
		}

		save(b4, []byte(r2.Slug), r2)

		r3 := Route{
			Title: "Not Found",
			Slug:  "notfound",
		}

		save(b4, []byte(r3.Slug), r3)

		r4 := Route{
			Title: "Some content",
			Slug:  "content-1",
		}

		save(b4, []byte(r4.Slug), r4)

		b5 := tx.Bucket([]byte(BUCKET_COMMON))

		menu := Menu{
			0: r1,
			1: r4,
			2: r2,
		}

		err = save(b5, []byte("menu"), menu)
		if err != nil {
			return err
		}

		return nil
	})
}

func save(b *bolt.Bucket, k []byte, v interface{}) error {
	buf, err := json.Marshal(v)
	if err != nil {
		return err
	}

	err = b.Put(k, buf)
	if err != nil {
		return err
	}

	return nil
}

func firstKey(m interface{}) (interface{}, error) {
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Map {
		return nil, errors.New("Not a map")
	}

	keys := v.MapKeys()
	if len(keys) == 0 {
		return nil, errors.New("Empty map")
	}

	return keys[0], nil
}
