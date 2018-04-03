package web

import (
	"bytes"
	"encoding/json"
	"html/template"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/99designs/basicauth-go"
	"github.com/dustin/go-humanize"
	"github.com/go-chi/chi"
	"github.com/gosimple/slug"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/session"
	"github.com/steam-authority/steam-authority/steam"
	"github.com/steam-authority/steam-authority/websockets"
)

func Serve() {

	r := chi.NewRouter()

	r.Mount("/admin", adminRouter())

	r.Get("/apps", AppsHandler)
	r.Get("/apps/{id}", AppHandler)
	r.Get("/apps/{id}/{slug}", AppHandler)

	r.Get("/changes", ChangesHandler)
	r.Get("/changes/{id}", ChangeHandler)

	r.Get("/chat", ChatHandler)
	r.Get("/chat/{id}", ChatHandler)

	r.Get("/contact", ContactHandler)
	r.Post("/contact", PostContactHandler)

	r.Get("/deals", DealsHandler)
	r.Get("/deals/{id}", DealsHandler)

	r.Get("/experience", ExperienceHandler)
	r.Get("/experience/{id}", ExperienceHandler)

	r.Get("/login", LoginHandler)
	r.Get("/logout", LogoutHandler)
	r.Get("/login-callback", LoginCallbackHandler)

	r.Get("/packages", PackagesHandler)
	r.Get("/packages/{id}", PackageHandler)

	r.Post("/players", PlayerIDHandler)
	r.Get("/players", RanksHandler)
	r.Get("/players/{id:[a-z]+}", RanksHandler)
	r.Get("/players/{id:[0-9]+}", PlayerHandler)
	r.Get("/players/{id:[0-9]+}/{slug}", PlayerHandler)

	r.Get("/queues", QueuesHandler)
	r.Get("/queues/queues.json", QueuesJSONHandler)

	r.Get("/settings", SettingsHandler)
	r.Post("/settings", SaveSettingsHandler)

	// Files
	r.Get("/browserconfig.xml", RootFileHandler)
	r.Get("/site.webmanifest", RootFileHandler)

	// Other
	r.Get("/", HomeHandler)
	r.Get("/commits", CommitsHandler)
	r.Get("/coop", CoopHandler)
	r.Get("/developers", StatsDevelopersHandler)
	r.Get("/donate", DonateHandler)
	r.Get("/genres", StatsGenresHandler)
	r.Get("/info", InfoHandler)
	r.Get("/news", NewsHandler)
	r.Get("/publishers", StatsPublishersHandler)
	r.Get("/stats", StatsHandler)
	r.Get("/tags", StatsTagsHandler)
	r.Get("/websocket", websockets.Handler)

	// 404
	r.NotFound(Error404Handler)

	// File server
	fileServer(r)

	http.ListenAndServe(":8085", r)
}

func adminRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(basicauth.New("Steam", map[string][]string{
		os.Getenv("STEAM_ADMIN_USER"): {os.Getenv("STEAM_ADMIN_PASS")},
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

	fs := http.StripPrefix(path, http.FileServer(http.Dir(filepath.Join(os.Getenv("STEAM_PATH"), "assets"))))

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

	// Load templates needed
	folder := os.Getenv("STEAM_PATH")
	t, err := template.New("t").Funcs(getTemplateFuncMap()).ParseFiles(
		folder+"/templates/_header.html",
		folder+"/templates/_footer.html",
		folder+"/templates/_stats_header.html",
		folder+"/templates/_pagination.html",
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
		return
	} else {
		// No error, send the content, HTTP 200 response status implied
		buf.WriteTo(w)
		return
	}

	return nil
}

func returnErrorTemplate(w http.ResponseWriter, r *http.Request, code int, message string) {

	w.WriteHeader(code)

	tmpl := errorTemplate{}
	tmpl.Fill(r, "Error")
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
		"join":   func(a []string) string { return strings.Join(a, ", ") },
		"title":  func(a string) string { return strings.Title(a) },
		"comma":  func(a int) string { return humanize.Comma(int64(a)) },
		"commaf": func(a float64) string { return humanize.Commaf(a) },
		"slug":   func(a string) string { return slug.Make(a) },
		"apps": func(a []int, appsMap map[int]mysql.App) template.HTML {
			var apps []string
			for _, v := range a {
				apps = append(apps, "<a href=\"/apps/"+strconv.Itoa(v)+"\">"+appsMap[v].GetName()+"</a>")
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
			var tags []string
			for _, v := range a {
				tags = append(tags, "<a href=\"/apps?tags="+strconv.Itoa(v.ID)+"\">"+v.GetName()+"</a>")
			}
			return template.HTML(strings.Join(tags, ", "))
		},
		"genres": func(a []steam.AppDetailsGenre) template.HTML {
			var genres []string
			for _, v := range a {
				genres = append(genres, "<a href=\"/apps?genres="+strconv.Itoa(v.ID)+"\">"+v.Description+"</a>")
			}
			return template.HTML(strings.Join(genres, ", "))
		},
		"unix":       func(t time.Time) int64 { return t.Unix() },
		"startsWith": func(a string, b string) bool { return strings.HasPrefix(a, b) },
		"contains":   func(a string, b string) bool { return strings.Contains(a, b) },
	}
}

// GlobalTemplate is added to every other template
type GlobalTemplate struct {
	Env     string
	ID      int
	Name    string // Username
	Title   string // page title
	Avatar  string
	Level   int
	Games   []int
	Path    string // URL
	IsAdmin bool
	request *http.Request // Internal
}

func (t *GlobalTemplate) Fill(r *http.Request, title string) {

	t.Title = title
	t.request = r

	// From ENV
	t.Env = os.Getenv("ENV")

	// From session
	id, _ := session.Read(r, session.ID)
	level, _ := session.Read(r, session.Level)

	t.ID, _ = strconv.Atoi(id)
	t.Name, _ = session.Read(r, session.Name)
	t.Avatar, _ = session.Read(r, session.Avatar)
	t.Avatar, _ = session.Read(r, session.Avatar)
	t.Level, _ = strconv.Atoi(level)

	gamesString, _ := session.Read(r, session.Games)
	if gamesString != "" {
		err := json.Unmarshal([]byte(gamesString), &t.Games)
		if err != nil {
			logger.Error(err)
			if strings.Contains(err.Error(), "cannot unmarshal") {
				logger.Info(gamesString)
			}
		}
	}

	// From request
	t.Path = r.URL.Path
	t.IsAdmin = r.Header.Get("Authorization") != ""
}

func (t GlobalTemplate) LoggedIn() (bool) {
	return t.ID > 0
}

func (t GlobalTemplate) IsLocal() (bool) {
	return t.Env == "local"
}

func (t GlobalTemplate) IsProduction() (bool) {
	return t.Env == "production"
}

func (t GlobalTemplate) ShowAd() (bool) {

	noAds := []string{
		"/admin",
		"/donate",
	}

	for _, v := range noAds {
		if strings.HasPrefix(t.request.URL.Path, v) {
			return false
		}
	}

	return true
}

type Pagination struct {
	page int
	last int
	path string
}

func (t Pagination) GetPages() (ret []int) {

	ret = append(ret, 1)
	for i := t.GetPage() - 2; i < t.GetPage()+3; i++ {
		if i >= 1 && i <= t.last {
			ret = append(ret, i)
		}
	}
	ret = append(ret, t.last)

	ret = helpers.Unique(ret)

	sort.Slice(ret, func(i, j int) bool {
		return ret[i] < ret[j]
	})

	return ret
}

func (t Pagination) GetNext() (float64) {
	return math.Min(float64(t.last), float64(t.GetPage()+1))
}

func (t Pagination) GetPrev() (float64) {
	return math.Max(1, float64(t.GetPage()-1))
}

func (t Pagination) GetPage() (int) {
	return int(math.Max(1, float64(t.page)))
}

func (t Pagination) GetLast() (int) {
	return t.last
}
