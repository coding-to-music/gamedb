package web

import (
	"bytes"
	"encoding/json"
	"html/template"
	"math"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/steam-go/steam"
	"github.com/derekstavis/go-qs"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/logging"
	"github.com/gamedb/website/session"
	"github.com/gamedb/website/websockets"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Called from main
func Init() {

	session.Init()

	InitChat()
	InitCommits()
}

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logging.InfoG(r.Method + " " + r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func middlewareTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		r.Header.Set("start-time", strconv.FormatInt(time.Now().UnixNano(), 10))

		next.ServeHTTP(w, r)
	})
}

func Serve() error {

	r := chi.NewRouter()

	r.Use(middlewareTime)
	r.Use(middleware.RealIP)
	r.Use(middleware.DefaultCompress)
	r.Use(middleware.GetHead)
	r.Use(middleware.RedirectSlashes)

	if viper.GetString("ENV") == logging.EnvProd {
		r.Use(middlewareLog)
	}

	// Pages
	r.Get("/", homeHandler)
	r.Mount("/admin", adminRouter())
	r.Mount("/changes", changesRouter())
	r.Mount("/chat", chatRouter())
	r.Get("/commits", commitsHandler)
	r.Mount("/contact", contactRouter())
	r.Get("/coop", coopHandler)
	r.Get("/discounts", discountsHandler)
	r.Get("/developers", statsDevelopersHandler)
	r.Get("/donate", donateHandler)
	r.Get("/esi/header", headerHandler)
	r.Mount("/experience", experienceRouter())
	r.Mount("/free-games", freeGamesRouter())
	r.Mount("/games", gamesRouter())
	r.Get("/genres", statsGenresHandler)
	r.Get("/health-check", healthCheckHandler)
	r.Get("/info", infoHandler)
	r.Mount("/login", loginRouter())
	r.Get("/logout", logoutHandler)
	r.Get("/news", newsHandler)
	r.Get("/news/ajax", newsAjaxHandler)
	r.Mount("/packages", packagesRouter())
	r.Mount("/players", playersRouter())
	r.Mount("/price-changes", priceChangeRouter())
	r.Get("/publishers", statsPublishersHandler)
	r.Mount("/queues", queuesRouter())
	r.Mount("/settings", settingsRouter())
	r.Mount("/stats", statsRouter())
	r.Get("/tags", statsTagsHandler)
	r.Mount("/upcoming", upcomingRouter())
	r.Get("/websocket/{id:[a-z]+}", websockets.WebsocketsHandler)

	// Files
	r.Get("/browserconfig.xml", rootFileHandler)
	r.Get("/robots.txt", rootFileHandler)
	r.Get("/sitemap.xml", siteMapHandler)
	r.Get("/site.webmanifest", rootFileHandler)

	// File server
	fileServer(r)

	// 404
	r.NotFound(Error404Handler)

	return http.ListenAndServe("0.0.0.0:"+viper.GetString("PORT"), r)
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func fileServer(r chi.Router) {

	path := "/assets"

	if strings.ContainsAny(path, "{}*") {
		logging.Info("FileServer does not permit URL parameters.")
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

	folder := viper.GetString("PATH")
	t, err := template.New("t").Funcs(getTemplateFuncMap()).ParseFiles(
		folder+"/templates/_header.gohtml",
		folder+"/templates/_header_esi.gohtml",
		folder+"/templates/_footer.gohtml",
		folder+"/templates/_stats_header.gohtml",
		folder+"/templates/_deals_header.gohtml",
		folder+"/templates/_flashes.gohtml",
		folder+"/templates/"+page+".gohtml",
	)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Something has gone wrong!", Error: err})
		return err
	}

	// Write a respone
	buf := &bytes.Buffer{}
	err = t.ExecuteTemplate(buf, page, pageData)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "Something has gone wrong!", Error: err})
		return err
	}

	w.WriteHeader(200)
	_, err = buf.WriteTo(w)
	logging.Error(err)

	return nil
}

func returnErrorTemplate(w http.ResponseWriter, r *http.Request, data errorTemplate) {

	if data.Title == "" {
		data.Title = "Error " + strconv.Itoa(data.Code)
	}

	if data.Code == 0 {
		data.Code = 500
	}

	logging.Error(data.Error)

	data.Fill(w, r, "Error")

	w.WriteHeader(data.Code)

	err := returnTemplate(w, r, "error", data)
	logging.Error(err)
}

type errorTemplate struct {
	GlobalTemplate
	Title   string
	Message string
	Code    int
	Error   error
}

func getTemplateFuncMap() map[string]interface{} {
	return template.FuncMap{
		"join":   func(a []string) string { return strings.Join(a, ", ") },
		"title":  func(a string) string { return strings.Title(a) },
		"comma":  func(a int) string { return humanize.Comma(int64(a)) },
		"commaf": func(a float64) string { return humanize.Commaf(a) },
		"slug":   func(a string) string { return slug.Make(a) },
		"apps": func(a []int, appsMap map[int]db.App) template.HTML {
			var apps []string
			for _, v := range a {
				apps = append(apps, "<a href=\"/games/"+strconv.Itoa(v)+"\">"+appsMap[v].GetName()+"</a>")
			}
			return template.HTML("Apps: " + strings.Join(apps, ", "))
		},
		"packages": func(a []int, packagesMap map[int]db.Package) template.HTML {
			var packages []string
			for _, v := range a {
				packages = append(packages, "<a href=\"/packages/"+strconv.Itoa(v)+"\">"+packagesMap[v].GetName()+"</a>")
			}
			return template.HTML("Packages: " + strings.Join(packages, ", "))
		},
		"tags": func(a []db.Tag) template.HTML {

			sort.Slice(a, func(i, j int) bool {
				return a[i].Name < a[j].Name
			})

			var tags []string
			for _, v := range a {
				tags = append(tags, "<a class=\"badge badge-success\" href=\"/games?tags="+strconv.Itoa(v.ID)+"\">"+v.GetName()+"</a>")
			}
			return template.HTML(strings.Join(tags, " "))
		},
		"genres": func(a []steam.AppDetailsGenre) template.HTML {

			sort.Slice(a, func(i, j int) bool {
				return a[i].Description < a[j].Description
			})

			var genres []string
			for _, v := range a {
				genres = append(genres, "<a class=\"badge badge-success\" href=\"/games?genres="+strconv.Itoa(v.ID)+"\">"+v.Description+"</a>")
			}
			return template.HTML(strings.Join(genres, " "))
		},
		"unix":       func(t time.Time) int64 { return t.Unix() },
		"startsWith": func(a string, b string) bool { return strings.HasPrefix(a, b) },
		"contains":   func(a string, b string) bool { return strings.Contains(a, b) },
		"max":        func(a int, b int) float64 { return math.Max(float64(a), float64(b)) },
		"json": func(v interface{}) (string, error) {
			b, err := json.Marshal(v)
			logging.Error(err)
			return string(b), err
		},
	}
}

// GlobalTemplate is added to every other template
type GlobalTemplate struct {
	Title       string // Page title
	Description string // Page description

	Avatar string
	Path   string // URL
	Env    string

	Toasts []Toast

	// User
	UserName           string // Username
	UserID             int
	UserLevel          int
	UserCountry        steam.CountryCode
	UserCurrencySymbol string

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
	logging.Error(err)

	if id == "" {
		t.UserID = 0
	} else {
		t.UserID, err = strconv.Atoi(id)
		logging.Error(err)
	}

	// User name
	t.UserName, err = session.Read(r, session.PlayerName)
	logging.Error(err)

	// Level
	level, err := session.Read(r, session.PlayerLevel)
	logging.Error(err)

	if level == "" {
		t.UserLevel = 0
	} else {
		t.UserLevel, err = strconv.Atoi(level)
		logging.Error(err)
	}

	// Country
	var code = session.GetCountryCode(r)
	t.UserCountry = code
	t.UserCurrencySymbol = helpers.CurrencySymbol(code)

	// Flashes
	t.FlashesGood, err = session.GetGoodFlashes(w, r)
	logging.Error(err)

	t.FlashesBad, err = session.GetBadFlashes(w, r)
	logging.Error(err)

	// All session data
	t.Session, err = session.ReadAll(r)
	logging.Error(err)
}

func (t GlobalTemplate) GetFooterText() (text string) {

	ts := time.Now()
	dayint, err := strconv.Atoi(ts.Format("2"))
	logging.Error(err)

	text = "Page created @ " + ts.Format("15:04:05") + " on " + ts.Format("Mon") + " " + humanize.Ordinal(dayint)

	// Get cashed
	if t.IsCache() {
		text += " from cache"
	}

	// Get time
	startTimeString := t.request.Header.Get("start-time")
	if startTimeString == "" {
		return text
	}

	startTimeInt, err := strconv.ParseInt(startTimeString, 10, 64)
	if err != nil {
		logging.Error(err)
		return text
	}

	d := time.Duration(time.Now().UnixNano() - startTimeInt)

	return text + " in " + d.String()
}

func (t GlobalTemplate) IsLoggedIn() bool {
	return t.UserID > 0
}

func (t GlobalTemplate) IsLocal() bool {
	return t.Env == "local"
}

func (t GlobalTemplate) IsCache() bool {
	return t.request.Header.Get("X-Cache") == "HIT"
}

func (t GlobalTemplate) IsProduction() bool {
	return t.Env == "production"
}

func (t GlobalTemplate) IsAdmin() bool {
	return t.request.Header.Get("Authorization") != ""
}

func (t GlobalTemplate) GetUserJSON() string {

	stringMap := map[string]interface{}{
		"userID":         t.UserID,
		"userLevel":      t.UserLevel,
		"userName":       t.UserName,
		"isLoggedIn":     t.IsLoggedIn(),
		"isLocal":        t.IsLocal(),
		"showAds":        t.ShowAd(),
		"country":        t.UserCountry,
		"currencySymbol": t.UserCurrencySymbol,
	}

	b, err := json.Marshal(stringMap)
	logging.Error(err)
	return string(b)
}

func (t GlobalTemplate) ShowAd() bool {
	return !t.IsLocal()
}

func (t *GlobalTemplate) AddToast(toast Toast) {
	t.Toasts = append(t.Toasts, toast)
}

type DataTablesAjaxResponse struct {
	Draw            string          `json:"draw"`
	RecordsTotal    string          `json:"recordsTotal"`
	RecordsFiltered string          `json:"recordsFiltered"`
	Data            [][]interface{} `json:"data"`
}

func (t *DataTablesAjaxResponse) AddRow(row []interface{}) {
	t.Data = append(t.Data, row)
}

func (t DataTablesAjaxResponse) output(w http.ResponseWriter) {

	if len(t.Data) == 0 {
		t.Data = make([][]interface{}, 0)
	}

	bytesx, err := json.Marshal(t)
	logging.Error(err)

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(bytesx)
	logging.Error(err)
}

type DataTablesQuery struct {
	Draw   string
	Order  map[string]map[string]interface{}
	Start  string
	Search map[string]interface{}
	Time   string `mapstructure:"_"`
}

func (q *DataTablesQuery) FillFromURL(url url.Values) (err error) {

	// Convert string into map
	queryMap, err := qs.Unmarshal(url.Encode())
	if err != nil {
		return err
	}

	// Convert map into struct
	err = mapstructure.Decode(queryMap, q)
	if err != nil {
		return err
	}

	return nil
}

func (q DataTablesQuery) GetSearchString(k string) (search string) {

	if val, ok := q.Search[k]; ok {
		if ok && val != "" {
			return val.(string)
		}
	}

	return ""
}

func (q DataTablesQuery) GetSearchSlice(k string) (search []string) {

	if val, ok := q.Search[k]; ok {
		if val != "" {
			for _, v := range val.([]interface{}) {
				search = append(search, v.(string))
			}
		}
	}

	return search
}

func (q DataTablesQuery) GetOrderSQL(columns map[string]string, code steam.CountryCode) (order string) {

	var ret []string

	for _, v := range q.Order {

		if col, ok := v["column"].(string); ok {
			if ok {

				if dir, ok := v["dir"].(string); ok {
					if ok {

						if col, ok := columns[col]; ok {
							if ok {

								if col == "price" {
									col = "JSON_EXTRACT(prices, \"$." + string(code) + ".final\")"
								}

								if dir == "asc" || dir == "desc" {
									ret = append(ret, col+" "+dir)
								}
							}
						}
					}
				}
			}
		}
	}

	return strings.Join(ret, ", ")
}

func (q DataTablesQuery) GetOrderDS(columns map[string]string, signed bool) (order string) {

	for _, v := range q.Order {

		if col, ok := v["column"].(string); ok {
			if ok {

				if dir, ok := v["dir"].(string); ok {
					if ok {

						if col, ok := columns[col]; ok {
							if ok {

								if dir == "desc" && signed {
									col = "-" + col
								}
								return col
							}
						}
					}
				}
			}
		}
	}

	return ""
}

func (q DataTablesQuery) SetOrderOffsetGorm(db *gorm.DB, code steam.CountryCode, columns map[string]string) *gorm.DB {

	db = db.Order(q.GetOrderSQL(columns, code))
	db = db.Offset(q.Start)

	return db
}

func (q DataTablesQuery) SetOrderOffsetDS(qu *datastore.Query, columns map[string]string) (*datastore.Query, error) {

	qu, err := q.SetOffsetDS(qu)
	if err != nil {
		return qu, err
	}

	order := q.GetOrderDS(columns, true)
	if order != "" {
		qu = qu.Order(order)
	}

	return qu, nil
}

func (q DataTablesQuery) SetOffsetDS(qu *datastore.Query) (*datastore.Query, error) {

	i, err := strconv.Atoi(q.Start)
	if err != nil {
		return qu, err
	}

	qu = qu.Offset(i)

	return qu, nil
}

func setNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	w.Header().Set("Expires", "0")                                         // Proxies.
}

// Toasts
type Toast struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	Link    string `json:"link"`
	Theme   string `json:"theme"`
	Timeout int    `json:"timeout"`
}
