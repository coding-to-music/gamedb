package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/99designs/basicauth-go"
	"github.com/Jleagle/steam-go/steam"
	"github.com/dustin/go-humanize"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/schema"
	"github.com/gosimple/slug"
	"github.com/spf13/viper"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/session"
	"github.com/steam-authority/steam-authority/websockets"
)

func Serve() error {

	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	r.Use(middleware.DefaultCompress)

	if viper.GetString("ENV") == logger.Local {
		//r.Use(middleware.Logger)
	}

	r.Use(middleware.GetHead)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RedirectSlashes)
	r.Use(middleware.RequestID)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", HomeHandler)
	r.Mount("/admin", adminRouter())
	r.Get("/browserconfig.xml", RootFileHandler)
	r.Get("/changes", ChangesHandler)
	r.Get("/changes/{id}", ChangeHandler)
	r.Get("/chat", ChatHandler)
	r.Get("/chat/{id}", ChatHandler)
	r.Get("/commits", CommitsHandler)
	r.Get("/contact", ContactHandler)
	r.Post("/contact", PostContactHandler)
	r.Get("/coop", CoopHandler)
	r.Get("/discounts", DiscountsHandler)
	r.Get("/developers", StatsDevelopersHandler)
	r.Get("/donate", DonateHandler)
	r.Get("/esi/header", HeaderHandler)
	r.Get("/experience", ExperienceHandler)
	r.Get("/experience/{id}", ExperienceHandler)
	r.Get("/free-games", FreeGamesHandler)
	r.Get("/free-games/ajax", FreeGamesAjaxHandler)
	r.Get("/games", AppsHandler)
	r.Get("/games/{id}", AppHandler)
	r.Get("/games/{id}/{slug}", AppHandler)
	r.Get("/health-check", HealthCheckHandler)
	r.Get("/genres", StatsGenresHandler)
	r.Get("/info", InfoHandler)
	r.Get("/login", LoginHandler)
	r.Post("/login", LoginPostHandler)
	r.Get("/login/openid", LoginOpenIDHandler)
	r.Get("/login/callback", LoginCallbackHandler)
	r.Get("/logout", LogoutHandler)
	r.Get("/news", NewsHandler)
	r.Get("/packages", PackagesHandler)
	r.Get("/packages/{id}", PackageHandler)
	r.Get("/packages/{id}/{slug}", PackageHandler)
	r.Get("/players", RanksHandler)
	r.Post("/players", PlayerIDHandler)
	r.Get("/players/{id:[a-z]+}", RanksHandler)
	r.Get("/players/{id:[0-9]+}", PlayerHandler)
	r.Get("/players/{id:[0-9]+}/{slug}", PlayerHandler)
	r.Get("/price-changes", PriceChangesHandler)
	r.Get("/publishers", StatsPublishersHandler)
	r.Get("/queues", QueuesHandler)
	r.Get("/queues/queues.json", QueuesJSONHandler)
	r.Get("/robots.txt", RootFileHandler)
	r.Get("/settings", SettingsHandler)
	r.Post("/settings", SettingsPostHandler)
	r.Get("/sitemap.xml", SiteMapHandler)
	r.Get("/site.webmanifest", RootFileHandler)
	r.Get("/stats", StatsHandler)
	r.Get("/tags", StatsTagsHandler)
	r.Get("/websocket", websockets.Handler)

	// File server
	fileServer(r)

	// 404
	r.NotFound(Error404Handler)

	return http.ListenAndServe("0.0.0.0:"+viper.GetString("PORT"), r)
}

func adminRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(basicauth.New("Steam", map[string][]string{
		viper.GetString("ADMIN_USER"): {viper.GetString("ADMIN_PASS")},
	}))
	r.Get("/", AdminHandler)
	r.Get("/{option}", AdminHandler)
	r.Post("/{option}", AdminHandler)
	return r
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func fileServer(r chi.Router) {

	path := "/assets"

	if strings.ContainsAny(path, "{}*") {
		logger.Info("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(http.Dir(filepath.Join(viper.GetString("PATH"), "assets"))))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}

func returnTemplate(w http.ResponseWriter, r *http.Request, page string, pageData interface{}) (err error) {

	w.Header().Set("Content-Type", "text/html")

	// Load templates needed
	folder := viper.GetString("PATH")
	t, err := template.New("t").Funcs(getTemplateFuncMap()).ParseFiles(
		folder+"/templates/_header.html",
		folder+"/templates/_header_esi.html",
		folder+"/templates/_footer.html",
		folder+"/templates/_stats_header.html",
		folder+"/templates/_deals_header.html",
		folder+"/templates/_pagination.html",
		folder+"/templates/_flashes.html",
		folder+"/templates/"+page+".html",
	)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 404, err.Error())
		return err
	}

	// Write a respone
	buf := &bytes.Buffer{}
	err = t.ExecuteTemplate(buf, page, pageData)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, "Something has gone wrong, the error has been logged!")
		return err
	} else {
		// No error, send the content, HTTP 200 response status implied
		buf.WriteTo(w)
	}

	return nil
}

func returnErrorTemplate(w http.ResponseWriter, r *http.Request, code int, message string) {

	w.WriteHeader(code)

	tmpl := errorTemplate{}
	tmpl.Fill(w, r, "Error")
	tmpl.Code = code
	tmpl.Message = message

	returnTemplate(w, r, "error", tmpl)
}

type errorTemplate struct {
	GlobalTemplate
	Code    int
	Message string
}

func getTemplateFuncMap() map[string]interface{} {
	return template.FuncMap{
		"join":    func(a []string) string { return strings.Join(a, ", ") },
		"title":   func(a string) string { return strings.Title(a) },
		"comma":   func(a int) string { return humanize.Comma(int64(a)) },
		"comma64": func(a int64) string { return humanize.Comma(a) },
		"commaf":  func(a float64) string { return humanize.Commaf(a) },
		"slug":    func(a string) string { return slug.Make(a) },
		"apps": func(a []int, appsMap map[int]mysql.App) template.HTML {
			var apps []string
			for _, v := range a {
				apps = append(apps, "<a href=\"/games/"+strconv.Itoa(v)+"\">"+appsMap[v].GetName()+"</a>")
			}
			return template.HTML("Apps: " + strings.Join(apps, ", "))
		},
		"packages": func(a []int, packagesMap map[int]mysql.Package) template.HTML {
			var packages []string
			for _, v := range a {
				packages = append(packages, "<a href=\"/packages/"+strconv.Itoa(v)+"\">"+packagesMap[v].GetName()+"</a>")
			}
			return template.HTML("Packages: " + strings.Join(packages, ", "))
		},
		"tags": func(a []mysql.Tag) template.HTML {

			sort.Slice(a, func(i, j int) bool {
				return a[i].Name < a[j].Name
			})

			var tags []string
			for _, v := range a {
				tags = append(tags, "<a href=\"/games?tags="+strconv.Itoa(v.ID)+"\">"+v.GetName()+"</a>")
			}
			return template.HTML(strings.Join(tags, ", "))
		},
		"genres": func(a []steam.AppDetailsGenre) template.HTML {

			sort.Slice(a, func(i, j int) bool {
				return a[i].Description < a[j].Description
			})

			var genres []string
			for _, v := range a {
				genres = append(genres, "<a href=\"/games?genres="+strconv.Itoa(v.ID)+"\">"+v.Description+"</a>")
			}
			return template.HTML(strings.Join(genres, ", "))
		},
		"unix":       func(t time.Time) int64 { return t.Unix() },
		"startsWith": func(a string, b string) bool { return strings.HasPrefix(a, b) },
		"contains":   func(a string, b string) bool { return strings.Contains(a, b) },
		"max":        func(a int, b int) float64 { return math.Max(float64(a), float64(b)) },
	}
}

// GlobalTemplate is added to every other template
type GlobalTemplate struct {
	Title  string // page title
	Avatar string
	Path   string // URL
	Env    string

	// User
	UserName  string // Username
	UserID    int
	UserLevel int

	// Session
	FlashesGood []interface{}
	FlashesBad  []interface{}
	Session     map[interface{}]interface{}

	//
	request *http.Request // Internal
}

func (t *GlobalTemplate) Fill(w http.ResponseWriter, r *http.Request, title string) {

	var err error

	t.request = r
	t.Title = title
	t.Env = viper.GetString("ENV")
	t.Path = r.URL.Path

	// User ID
	id, err := session.Read(r, session.PlayerID)
	logger.Error(err)

	t.UserID, err = strconv.Atoi(id)
	if err != nil {
		t.UserID = 0
	}

	// User name
	t.UserName, err = session.Read(r, session.PlayerName)
	logger.Error(err)

	// Level
	level, err := session.Read(r, session.PlayerLevel)
	logger.Error(err)

	t.UserLevel, err = strconv.Atoi(level)
	if err != nil {
		t.UserLevel = 0
	}

	// Flashes
	t.FlashesGood, err = session.GetGoodFlashes(w, r)
	logger.Error(err)

	t.FlashesBad, err = session.GetBadFlashes(w, r)
	logger.Error(err)

	// All session data
	t.Session, err = session.ReadAll(r)
	logger.Error(err)
}

func (t GlobalTemplate) IsLoggedIn() (bool) {
	return t.UserID > 0
}

func (t GlobalTemplate) IsLocal() (bool) {
	return t.Env == "local"
}

func (t GlobalTemplate) IsVarnished() (bool) {
	return t.request.Header.Get("through-varnish") == "true"
}

func (t GlobalTemplate) IsProduction() (bool) {
	return t.Env == "production"
}

func (t GlobalTemplate) IsAdmin() (bool) {
	return t.request.Header.Get("Authorization") != ""
}

func (t GlobalTemplate) GetUserJSON() (string) {

	stringMap := map[string]interface{}{
		"userID":     t.UserID,
		"userLevel":  t.UserLevel,
		"userName":   t.UserName,
		"isLoggedIn": t.IsLoggedIn(),
		"isLocal":    t.IsLocal(),
		"showAds":    t.ShowAd(),
	}

	bytesx, err := json.Marshal(stringMap)
	logger.Error(err)
	return string(bytesx)
}

func (t GlobalTemplate) ShowAd() (bool) {

	noAds := []string{
		"/admin",
		"/donate",
		"/settings",
	}

	for _, v := range noAds {
		if strings.HasPrefix(t.request.URL.Path, v) {
			return false
		}
	}

	return true
}

type Pagination struct {
	path  string
	page  int
	limit int
	total int
}

func (t Pagination) GetPages() (ret []int) {

	ret = append(ret, 1)
	for i := t.GetPage() - 2; i < t.GetPage()+3; i++ {
		if i >= 1 && i <= t.GetLast() {
			ret = append(ret, i)
		}
	}
	ret = append(ret, t.GetLast())

	ret = helpers.Unique(ret)

	sort.Slice(ret, func(i, j int) bool {
		return ret[i] < ret[j]
	})

	return ret
}

func (t Pagination) GetNext() (float64) {
	return math.Min(float64(t.GetLast()), float64(t.GetPage()+1))
}

func (t Pagination) GetPrev() (float64) {
	return math.Max(1, float64(t.GetPage()-1))
}

func (t Pagination) GetPage() (int) {
	return int(math.Max(1, float64(t.page)))
}

func (t Pagination) GetLast() (int) {
	last := math.Ceil(float64(t.total) / float64(t.limit))
	return int(math.Max(1, last))
}

func (t Pagination) GetPath() string {
	return t.path
}

func (t Pagination) GetLinks() string {
	// todo, return link html here instead of making in template, that way we can add a variable to the path and replace with the page number
	return ""
}

type SimplePagination struct {
	path  string
	next  string
	prev  string
	limit int
}

type DataTablesAjaxResponse struct {
	Draw            string     `json:"-"`
	RecordsTotal    string     `json:"recordsTotal"`
	RecordsFiltered string     `json:"recordsFiltered"`
	Data            [][]string `json:"data"`
}

func (t DataTablesAjaxResponse) AddRow(row []string) {
	t.Data = append(t.Data, row)
}

type DataTablesQuery struct {
	Draw   string                         `json:"draw"`
	Order  []map[string]map[string]string `json:"order"`
	Start  string                         `json:"start"`
	Length string                         `json:"length"`
	Search map[string]string              `json:"search"`
}

func (q *DataTablesQuery) FillFromURL(url *url.URL) (err error) {

	var decoder = schema.NewDecoder()

	err = decoder.Decode(q, url.Query())
	if err != nil {
		logger.Info(err.Error())
		return err
	}

	fmt.Println(q)

	return nil
}
