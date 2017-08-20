package main

import (
	"encoding/xml"
	"time"
)

// Interfaces
type Configurable interface {
	Configure(Configuration) error
}

type Finalizable interface {
	Finalize()
}

// Types
type User struct {
	Name, Title, About                string
	Image, Logo                       Media
	Joined                            time.Time
	References, Networks, Experiences Map
	Interests                         []Interest
	Contact                           Contact
}

type Project struct {
	Slug string

	Title, Subtitle, About string
	Image, Logo            Media
	Published              time.Time
	Images                 []Media
	Tags                   []Tag
	Technologies           []Technology
	References             Map
	Client                 Client
	Imported               Imported
}

type Content struct {
	Slug string

	Title, Subtitle string
	Published       time.Time
	Paragraphs      []Paragraph
	Tags            []Tag
	Technologies    []Technology
	References      Map
}

type Menu map[int]Route

type Route struct {
	Slug, Title string
}

type Paragraph struct {
	Slug string

	Title, Content string
	Media          Media
}

type Page struct {
	Title string

	Type PageType

	User User
	Meta Meta
	Menu Menu

	Content interface{}
}

type Client struct {
	Name, About string
	Image       Media
}

type Media struct {
	Type                      MediaType
	Name, Caption, Path, Mime string
}

type Contact struct {
	Country string `json:"country"`
	City    string `json:"city"`
	Street  string `json:"street"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
}

type Imported struct {
	ExternalID      int
	ExternalService string

	Date time.Time
}

type Credentials struct {
	Email, Hash string
}

type ProjectStatistics struct {
	Views, Likes uint64
}

type Configuration struct {
	SetupCompleted bool
	JwtSecret      string

	CurrentThemePath string
	CurrentTheme     Theme
	Meta             Meta
}

type Theme struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Image       string   `json:"image"`
	Css         []string `json:"css"`
	Js          []string `json:"js"`
}

type Meta struct {
	Title, Site  string
	Tags, OGTags map[string]string
}

type File struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Data []byte `json:"data"`
}

type Name string

type Link string

type Map map[string]string

type Technology string

type Interest string

type Tag string

type PageType uint8

const (
	PageRegular  PageType = iota
	PageProject  PageType = iota
	PageHome     PageType = iota
	PageContact  PageType = iota
	PageNotFound PageType = iota
)

type MediaType uint8

const (
	MediaOther MediaType = iota
	MediaImage MediaType = iota
	MediaVideo MediaType = iota
)

type UrlSet struct {
	XMLName xml.Name `xml:"urlset"`
	XMLns   string   `xml:"xmlns,attr"`
	Urls    []Url
}

type Url struct {
	XMLName    xml.Name `xml:"url"`
	Loc        string   `xml:"loc"`
	LastMod    string   `xml:"lastmod"`
	ChangeFreq string   `xml:"changefreq"`
}

// Requests
type UpdateUserRequest struct {
	Name        string     `json:"name"`
	Title       string     `json:"title"`
	About       string     `json:"about"`
	Image       []Media_   `json:"image"`
	Logo        []Media_   `json:"logo"`
	References  Map        `json:"references"`
	Networks    Map        `json:"networks"`
	Experiences Map        `json:"experiences"`
	Interests   []Interest `json:"interests"`
	Contact     Contact    `json:"contact"`
}

type UpdateThemeRequest struct {
	Path string `json:"path"`
}

type UpdateMetaRequest struct {
	Title  string            `json:"title"`
	Site   string            `json:"site"`
	Tags   map[string]string `json:"tags"`
	OGTags map[string]string `json:"og_tags"`
}

type UpdateCredentialsRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	PasswordRepeat string `json:"password_repeat"`
}

type CreateProjectRequest struct {
	Title        string       `json:"title"`
	Subtitle     string       `json:"subtitle"`
	About        string       `json:"about"`
	Image        []Media_     `json:"image"`
	Logo         []Media_     `json:"logo"`
	Media        []Media_     `json:"media"`
	Tags         []Tag        `json:"tags"`
	Technologies []Technology `json:"technologies"`
	References   Map          `json:"references"`
	Client       struct {
		Name  string   `json:"name"`
		About string   `json:"about"`
		Image []Media_ `json:"image"`
	} `json:"client"`
}

type UpdateProjectRequest struct {
	Slug string `json:"slug"`

	Title        string       `json:"title"`
	Subtitle     string       `json:"subtitle"`
	About        string       `json:"about"`
	Image        []Media_     `json:"image"`
	Logo         []Media_     `json:"logo"`
	Media        []Media_     `json:"media"`
	Tags         []Tag        `json:"tags"`
	References   Map          `json:"references"`
	Technologies []Technology `json:"technologies"`

	Client struct {
		Name  string   `json:"name"`
		About string   `json:"about"`
		Image []Media_ `json:"image"`
	} `json:"client"`
}

type DeleteProjectRequest struct {
	Slug string `json:"slug"`
}

type CreateContentRequest struct {
	Title        string       `json:"title"`
	Subtitle     string       `json:"subtitle"`
	Paragraphs   []Paragraph_ `json:"paragraphs"`
	Tags         []Tag        `json:"tags"`
	References   Map          `json:"references"`
	Technologies []Technology `json:"technologies"`
}

type UpdateContentRequest struct {
	Slug string `json:"slug"`

	Title        string       `json:"title"`
	Subtitle     string       `json:"subtitle"`
	Paragraphs   []Paragraph_ `json:"paragraphs"`
	Tags         []Tag        `json:"tags"`
	References   Map          `json:"references"`
	Technologies []Technology `json:"technologies"`
}

type DeleteContentRequest struct {
	Slug string `json:"slug"`
}

type AddToMenuRequest struct {
	Slug string `json:"slug"`
}

type RemoveFromMenuRequest struct {
	Slug string `json:"slug"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

//Response
type GenericResponse struct {
	Error   string      `json:"error,omitempty"`
	Content interface{} `json:"content,omitempty"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// API models
type Media_ struct {
	Resource string `json:"resource"`

	Name    string `json:"name"`
	Caption string `json:"caption"`

	Uploaded bool `json:"uploaded"`
	Removed  bool `json:"removed"`

	File File `json:"file"`
}

type Project_ struct {
	Slug string `json:"slug"`

	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Image    string `json:"image"`
}

type Content_ struct {
	Slug string `json:"slug"`

	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}

type Paragraph_ struct {
	Resource string `json:"resource"`

	Title   string   `json:"title"`
	Content string   `json:"content"`
	Media   []Media_ `json:"media"`
}

type Theme_ struct {
	Name   string `json:"name"`
	Author string `json:"author"`
	Image  string `json:"image"`
	Path   string `json:"path"`
}

type Menu_ struct {
	Added  []string `json:"added"`
	Routes []string `json:"routes"`
}
