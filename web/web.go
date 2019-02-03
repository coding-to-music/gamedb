package web

import (
	"bytes"
	"encoding/json"
	"html/template"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/steam-go/steam"
	"github.com/derekstavis/go-qs"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
	"github.com/gamedb/website/websockets"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/jinzhu/gorm"
	"github.com/mitchellh/mapstructure"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"
)

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.Config.IsProd() {
			log.Info(log.ServiceGoogle, log.LogNameRequests, r.Method+" "+r.URL.Path)
		}
		next.ServeHTTP(w, r)
	})
}

func middlewareTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		r.Header.Set("start-time", strconv.FormatInt(time.Now().UnixNano(), 10))

		next.ServeHTTP(w, r)
	})
}

func middlewareCors() func(next http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins: []string{config.Config.GameDBDomain.Get()}, // Use this to allow specific origin hosts
		AllowedMethods: []string{"GET", "POST"},
	}).Handler
}

func Serve() error {

	r := chi.NewRouter()

	r.Use(middlewareTime)
	r.Use(middlewareCors())
	r.Use(middleware.RealIP)
	r.Use(middleware.DefaultCompress)
	r.Use(middleware.RedirectSlashes)
	r.Use(middlewareLog)

	// Pages
	r.Get("/", homeRedirectHandler)
	r.Get("/commits", commitsHandler)
	r.Get("/coop", coopHandler)
	r.Get("/discounts", discountsHandler)
	r.Get("/developers", statsDevelopersHandler)
	r.Get("/donate", donateHandler)
	r.Get("/esi/header", headerHandler)
	r.Get("/genres", statsGenresHandler)
	r.Get("/health-check", healthCheckHandler)
	r.Get("/info", infoHandler)
	r.Get("/logout", logoutHandler)
	r.Get("/news", newsHandler)
	r.Get("/news/ajax", newsAjaxHandler)
	r.Get("/publishers", statsPublishersHandler)
	r.Get("/tags", statsTagsHandler)
	r.Get("/websocket/{id:[a-z]+}", websockets.WebsocketsHandler)
	r.Mount("/admin", adminRouter())
	r.Mount("/apps", appsRouter())
	r.Mount("/bundles", bundlesRouter())
	r.Mount("/changes", changesRouter())
	r.Mount("/chat", chatRouter())
	r.Mount("/contact", contactRouter())
	r.Mount("/depots", depotsRouter())
	r.Mount("/experience", experienceRouter())
	r.Mount("/games", gamesRouter())
	r.Mount("/login", loginRouter())
	r.Mount("/packages", packagesRouter())
	r.Mount("/players", playersRouter())
	r.Mount("/price-changes", priceChangeRouter())
	r.Mount("/product-keys", productKeysRouter())
	r.Mount("/queues", queuesRouter())
	r.Mount("/settings", settingsRouter())
	r.Mount("/stats", statsRouter())
	r.Mount("/upcoming", upcomingRouter())

	// Files
	r.Get("/browserconfig.xml", rootFileHandler)
	r.Get("/robots.txt", rootFileHandler)
	r.Get("/sitemap.xml", siteMapHandler)
	r.Get("/site.webmanifest", rootFileHandler)

	// File server
	fileServer(r, "/assets", http.Dir("assets"))

	// 404
	r.NotFound(error404Handler)

	return http.ListenAndServe(config.Config.ListenOn(), r)
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func fileServer(r chi.Router, path string, root http.FileSystem) {

	if strings.ContainsAny(path, "{}*") {
		log.Info("Invalid URL " + path)
		return
	}
	if strings.Contains(path, "..") {
		log.Info("Invalid URL " + path)
		return
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}

func setNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	w.Header().Set("Expires", "0")                                         // Proxies.
}

func returnJSON(w http.ResponseWriter, r *http.Request, bytes []byte) (err error) {

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Language", string(session.GetCountryCode(r))) // Used for varnish hash

	_, err = w.Write(bytes)
	return err
}

func returnTemplate(w http.ResponseWriter, r *http.Request, page string, pageData interface{}) (err error) {

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Language", string(session.GetCountryCode(r))) // Used for varnish hash
	w.WriteHeader(200)

	folder := config.Config.GameDBDirectory.Get()
	t, err := template.New("t").Funcs(getTemplateFuncMap()).ParseFiles(
		folder+"/templates/_header.gohtml",
		folder+"/templates/_header_esi.gohtml",
		folder+"/templates/_footer.gohtml",
		folder+"/templates/_stats_header.gohtml",
		folder+"/templates/_deals_header.gohtml",
		folder+"/templates/_apps_header.gohtml",
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

	if config.Config.IsProd() {

		m := minify.New()
		m.Add("text/html", &html.Minifier{
			KeepConditionalComments: false,
			KeepDefaultAttrVals:     true,
			KeepDocumentTags:        true,
			KeepEndTags:             true,
			KeepWhitespace:          true,
		})

		buf2 := &bytes.Buffer{}
		err = m.Minify("text/html", buf2, buf)
		if err != nil {
			log.Err(err)
			return err
		}

		_, err = buf2.WriteTo(w)

	} else {

		_, err = buf.WriteTo(w)

	}

	log.Err(err)
	return err
}

func returnErrorTemplate(w http.ResponseWriter, r *http.Request, data errorTemplate) {

	if data.Title == "" {
		data.Title = "Error " + strconv.Itoa(data.Code)
	}

	if data.Code == 0 {
		data.Code = 500
	}

	log.Err(data.Error)

	data.Fill(w, r, "Error", "Something has gone wrong!")

	w.WriteHeader(data.Code)

	err := returnTemplate(w, r, "error", data)
	log.Err(err, r)
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
		"join":       func(a []string) string { return strings.Join(a, ", ") },
		"comma":      func(a int) string { return humanize.Comma(int64(a)) },
		"commaf":     func(a float64) string { return humanize.Commaf(a) },
		"bytes":      func(a uint64) string { return humanize.Bytes(a) },
		"seconds":    func(a int64) string { return humanize.RelTime(time.Now(), time.Now().Add(time.Second*time.Duration(a)), "", "") },
		"startsWith": func(a string, b string) bool { return strings.HasPrefix(a, b) },
		"endsWith":   func(a string, b string) bool { return strings.HasSuffix(a, b) },
		"max":        func(a int, b int) float64 { return math.Max(float64(a), float64(b)) },
		"json": func(v interface{}) (string, error) {
			b, err := json.Marshal(v)
			log.Err(err)
			return string(b), err
		},
	}
}

// GlobalTemplate is added to every other template
type GlobalTemplate struct {
	// These variables can be used in templates and cached
	Title       string        // Page title
	Description template.HTML // Page description
	Path        string        // URL path
	Env         string        // Environment
	CSSFiles    []Asset
	JSFiles     []Asset
	MetaImage   string

	// These variables can't!
	// Session
	userName           string
	userEmail          string
	userID             int
	userLevel          int
	userCountry        steam.CountryCode
	userCurrencySymbol string

	// Session
	flashesGood []interface{}
	flashesBad  []interface{}
	session     map[string]string

	//
	toasts []Toast

	//
	request *http.Request // Internal
}

func (t *GlobalTemplate) Fill(w http.ResponseWriter, r *http.Request, title string, description template.HTML) {

	var err error

	t.request = r

	t.Title = title + " - Game DB"
	t.Description = description
	t.Env = config.Config.Environment.Get()
	t.Path = r.URL.Path

	// User ID
	id, err := session.Read(r, session.PlayerID)
	log.Err(err, r)

	if id == "" {
		t.userID = 0
	} else {
		t.userID, err = strconv.Atoi(id)
		log.Err(err, r)
	}

	// User name
	t.userName, err = session.Read(r, session.PlayerName)
	log.Err(err, r)

	// Email
	t.userEmail, err = session.Read(r, session.UserEmail)
	log.Err(err, r)

	// Level
	level, err := session.Read(r, session.PlayerLevel)
	log.Err(err, r)

	if level == "" {
		t.userLevel = 0
	} else {
		t.userLevel, err = strconv.Atoi(level)
		log.Err(err, r)
	}

	// Country
	t.userCountry = steam.CountryCode(r.URL.Query().Get("cc"))

	// Check if valid country
	if _, ok := steam.Countries[t.userCountry]; !ok {
		t.userCountry = session.GetCountryCode(r)
	}

	// Default country to session
	if t.userCountry == "" {
		t.userCountry = session.GetCountryCode(r)
	}

	locale, err := helpers.GetLocaleFromCountry(t.userCountry)
	log.Err(err, r)

	t.userCurrencySymbol = locale.CurrencySymbol

	// Flashes
	t.flashesGood, err = session.GetGoodFlashes(w, r)
	log.Err(err, r)

	t.flashesBad, err = session.GetBadFlashes(w, r)
	log.Err(err, r)

	// All session data, todo, remove this, security etc
	t.session, err = session.ReadAll(r)
	log.Err(err, r)
}

func (t GlobalTemplate) GetUserJSON() string {

	stringMap := map[string]interface{}{
		"userID":         strconv.Itoa(t.userID), // Too long for JS int
		"userLevel":      t.userLevel,
		"userName":       t.userName,
		"userEmail":      t.userEmail,
		"isLoggedIn":     t.isLoggedIn(),
		"isLocal":        t.isLocal(),
		"isAdmin":        t.isAdmin(),
		"showAds":        t.showAds(),
		"country":        t.userCountry,
		"currencySymbol": t.userCurrencySymbol,
		"flashesGood":    t.flashesGood,
		"flashesBad":     t.flashesBad,
		"toasts":         t.toasts,
		"session":        t.session,
	}

	b, err := json.Marshal(stringMap)
	log.Err(err)

	return string(b)
}
func (t GlobalTemplate) GetMetaImage() (text string) {

	if t.MetaImage == "" {
		return "/assets/img/sa-bg-500x500.png"
	}

	return t.MetaImage
}

func (t GlobalTemplate) GetFooterText() (text string) {

	ts := time.Now()
	dayint, err := strconv.Atoi(ts.Format("2"))
	log.Err(err)

	text = "Page created on " + ts.Format("Mon") + " the " + humanize.Ordinal(dayint) + " @ " + ts.Format("15:04:05")

	// Get cashed
	if t.IsCacheHit() {
		text += " from cache"
	}

	// Get time
	startTimeString := t.request.Header.Get("start-time")
	if startTimeString == "" {
		return text
	}

	startTimeInt, err := strconv.ParseInt(startTimeString, 10, 64)
	if err != nil {
		log.Err(err)
		return text
	}

	d := time.Duration(time.Now().UnixNano() - startTimeInt)

	return text + " in " + d.String()
}

func (t GlobalTemplate) IsCacheHit() bool {
	return t.request.Header.Get("X-Cache") == "HIT"
}

func (t GlobalTemplate) IsFromVarnish() bool {
	return t.request.Header.Get("X-From-Varnish") == "true"
}

func (t GlobalTemplate) IsStatsPage() bool {
	return helpers.SliceHasString([]string{"stats", "tags", "genres", "publishers", "developers"}, strings.TrimPrefix(t.Path, "/"))
}

func (t GlobalTemplate) IsMorePage() bool {
	return helpers.SliceHasString([]string{"contact", "experience", "changes", "queues", "info"}, strings.TrimPrefix(t.Path, "/"))
}

func (t GlobalTemplate) isLoggedIn() bool {
	return t.userID > 0
}

func (t GlobalTemplate) isLocal() bool {
	return t.Env == string(log.EnvLocal)
}

func (t GlobalTemplate) isAdmin() bool {
	return isAdmin(t.request)
}

func (t GlobalTemplate) showAds() bool {
	return !t.isLocal()
}

func (t *GlobalTemplate) addToast(toast Toast) {
	t.toasts = append(t.toasts, toast)
}

func (t *GlobalTemplate) addAssetChosen() {
	t.JSFiles = append(t.JSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/chosen/1.8.7/chosen.jquery.min.js", Integrity: "sha256-c4gVE6fn+JRKMRvqjoDp+tlG4laudNYrXI1GncbfAYY="})
	t.CSSFiles = append(t.CSSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/chosen/1.8.7/chosen.min.css", Integrity: "sha256-EH/CzgoJbNED+gZgymswsIOrM9XhIbdSJ6Hwro09WE4="})
}

func (t *GlobalTemplate) addAssetJSON2HTML() {
	t.JSFiles = append(t.JSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/json2html/1.2.0/json2html.min.js", Integrity: "sha256-5iWhgkOOkWSQMxoIXqSKvZQHOTJ1wYDBqhMTFm5DkDw="})
	t.JSFiles = append(t.JSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/jquery.json2html/1.2.0/jquery.json2html.min.js", Integrity: "sha256-NVPR5gsJCl/e6xUJ3Wv2+4Tui2vhZY6KBhx0RY0DNcs="})
}

func (t *GlobalTemplate) addAssetHighCharts() {
	t.JSFiles = append(t.JSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/highcharts/7.0.1/highcharts.js", Integrity: "sha256-j3WPKr23emLOeDVvf5mbfGs5xE+GERqV1vCz+Wx6n74="})
	t.JSFiles = append(t.JSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/highcharts/7.0.1/modules/data.js", Integrity: "sha256-CYgititANzm6qnx8M/4TpaGqfa8xFOIbHfWbtvKAg4w="})
	// t.JSFiles = append(t.JSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/highcharts/7.0.1/modules/heatmap.js", Integrity: "sha256-HgUQ2+RnyQrmj1venzdV9Q6/ahkZ8h4HYoXNbGu7dpo="})
}

func (t *GlobalTemplate) addAssetSlider() {
	t.JSFiles = append(t.JSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/noUiSlider/12.1.0/nouislider.min.js", Integrity: "sha256-V76+FCDgnqVqafUQ74coiR7qA3Gd6ZlVuFgdwcGCGlc="})
	t.CSSFiles = append(t.CSSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/noUiSlider/12.1.0/nouislider.min.css", Integrity: "sha256-MyPOSprr9/vRwXTYc0saw86ylzGM2HVRKWUfHIFta74="})
}

func (t *GlobalTemplate) addAssetCarousel() {
	t.JSFiles = append(t.JSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/slick-carousel/1.9.0/slick.min.js", Integrity: "sha256-NXRS8qVcmZ3dOv3LziwznUHPegFhPZ1F/4inU7uC8h0="})
	t.CSSFiles = append(t.CSSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/slick-carousel/1.9.0/slick.min.css", Integrity: "sha256-UK1EiopXIL+KVhfbFa8xrmAWPeBjMVdvYMYkTAEv/HI="})
	t.CSSFiles = append(t.CSSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/slick-carousel/1.9.0/slick-theme.min.css", Integrity: "sha256-4hqlsNP9KM6+2eA8VUT0kk4RsMRTeS7QGHIM+MZ5sLY="})
}

func (t *GlobalTemplate) addAssetPasswordStrength() {
	t.JSFiles = append(t.JSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/pwstrength-bootstrap/3.0.2/pwstrength-bootstrap.min.js", Integrity: "sha256-BPKP4P2AbrV7hf80SHJAJkIvjt7X7MKFEPpA99uU6uQ="})
}

type Asset struct {
	URL       string
	Integrity string
}

// DataTablesAjaxResponse
type DataTablesAjaxResponse struct {
	Draw            string          `json:"draw"`
	RecordsTotal    string          `json:"recordsTotal"`
	RecordsFiltered string          `json:"recordsFiltered"`
	Data            [][]interface{} `json:"data"`
}

func (t *DataTablesAjaxResponse) AddRow(row []interface{}) {
	t.Data = append(t.Data, row)
}

func (t DataTablesAjaxResponse) output(w http.ResponseWriter, r *http.Request) {

	if len(t.Data) == 0 {
		t.Data = make([][]interface{}, 0)
	}

	bytesx, err := json.Marshal(t)
	log.Err(err, r)

	err = returnJSON(w, r, bytesx)
	log.Err(err, r)
}

// DataTablesQuery
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

// Toasts
type Toast struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	Link    string `json:"link"`
	Theme   string `json:"theme"`
	Timeout int    `json:"timeout"`
}

// Get prices ajax
func productPricesAjaxHandler(w http.ResponseWriter, r *http.Request, productType db.ProductType) {

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Err("invalid id")
		return
	}

	idx, err := strconv.Atoi(id)
	if err != nil {
		log.Err("invalid id")
		return
	}

	// Get product
	var product db.ProductInterface

	if productType == db.ProductTypeApp {
		product, err = db.GetApp(idx, []string{})
	} else {
		product, err = db.GetPackage(idx, []string{"id", "product_type", "prices"})
	}
	if err != nil {
		log.Err(err)
		return
	}

	// Get code
	code := steam.CountryCode(r.URL.Query().Get("code"))
	if code == "" {
		code = session.GetCountryCode(r)
	}

	if code == "" {
		log.Err("no code given")
		return
	}

	// Get prices from datastore
	pricesResp, err := db.GetProductPrices(product.GetID(), product.GetProductType(), code)
	if err != nil {
		log.Err(err, r)
		return
	}

	// Get locale
	locale, err := helpers.GetLocaleFromCountry(code)
	if err != nil {
		log.Err(err, r)
		return
	}

	// Make JSON response
	var response productPricesAjaxResponse
	response.Symbol = locale.CurrencySymbol

	for _, v := range pricesResp {
		response.Prices = append(response.Prices, []float64{float64(v.CreatedAt.Unix() * 1000), float64(v.PriceAfter) / 100})
	}

	// Add current price
	price, err := product.GetPrice(code)
	err = helpers.IgnoreErrors(err, db.ErrMissingCountryCode)
	if err != nil {
		log.Err(err, r)
		return
	}

	response.Prices = append(response.Prices, []float64{float64(time.Now().Unix()) * 1000, float64(price.Final) / 100})

	// Sort prices for Highcharts
	sort.Slice(response.Prices, func(i, j int) bool {
		return response.Prices[i][0] < response.Prices[j][0]
	})

	// Return
	pricesBytes, err := json.Marshal(response)
	if err != nil {
		log.Err(err, r)
		return
	}

	err = returnJSON(w, r, pricesBytes)
	log.Err(err)
}

type productPricesAjaxResponse struct {
	Prices [][]float64 `json:"prices"`
	Symbol string      `json:"symbol"`
}

func isAdmin(r *http.Request) bool {
	return r.Header.Get("Authorization") != ""
}
