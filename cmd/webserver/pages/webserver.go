package pages

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"html"
	"html/template"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/session-go/session"
	"github.com/derekstavis/go-qs"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/tdewolff/minify/v2"
	minhtml "github.com/tdewolff/minify/v2/html"
	"go.mongodb.org/mongo-driver/bson"
)

func setHeaders(w http.ResponseWriter, r *http.Request, contentType string) {

	csp := []string{
		"default-src 'none'",
		"script-src 'self' 'unsafe-eval' 'unsafe-inline' https://cdnjs.cloudflare.com https://cdn.datatables.net https://www.googletagmanager.com https://www.google-analytics.com https://connect.facebook.net https://platform.twitter.com https://www.google.com https://www.gstatic.com https://*.infolinks.com https://*.patreon.com https://*.hotjar.com",
		"style-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com https://cdn.datatables.net https://fonts.googleapis.com",
		"media-src https://steamcdn-a.akamaihd.net",
		"font-src https://fonts.gstatic.com https://cdnjs.cloudflare.com",
		"frame-src https://platform.twitter.com https://*.facebook.com https://www.youtube.com https://www.google.com https://www.patreon.com https://router.infolinks.com https://vars.hotjar.com",
		"connect-src 'self' ws: wss: https://*.infolinks.com https://in.hotjar.com https://vc.hotjar.io https://www.google-analytics.com https://stats.g.doubleclick.net",
		"manifest-src 'self'",
		"img-src 'self' data: *", // * to hotlink news article images, info link images etc
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")                // MIME sniffing
	w.Header().Set("X-XSS-Protection", "1; mode=block")                // XSS
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")                    // Clickjacking
	w.Header().Set("Content-Security-Policy", strings.Join(csp, "; ")) // XSS
	w.Header().Set("Referrer-Policy", "no-referrer-when-downgrade")
	w.Header().Set("Feature-Policy", "geolocation 'none'; midi 'none'; sync-xhr 'none'; microphone 'none'; camera 'none'; magnetometer 'none'; gyroscope 'none'; speaker 'none'; fullscreen 'none'; payment 'none';")
	w.Header().Set("Server", "")
}

func SetCacheHeaders(w http.ResponseWriter, duration time.Duration) {

	if w.Header().Get("Cache-Control") == "" || w.Header().Get("Expires") == "" {

		if duration == 0 || config.IsLocal() {

			w.Header().Set("Cache-Control", "max-age=0")
			w.Header().Set("Expires", time.Now().AddDate(0, 0, -2).Format(time.RFC1123))

		} else {

			w.Header().Set("Cache-Control", "max-age="+strconv.Itoa(int(duration.Seconds())))
			w.Header().Set("Expires", time.Now().Add(duration).Format(time.RFC1123))
		}
	}
}

func returnJSON(w http.ResponseWriter, r *http.Request, i interface{}) {

	setHeaders(w, r, "application/json")

	b, err := json.Marshal(i)
	if err != nil {
		log.Err(err, r)
		return
	}

	_, err = w.Write(b)
	if err != nil && !strings.Contains(err.Error(), "write: broken pipe") {
		log.Critical(err)
	}
}

func returnTemplate(w http.ResponseWriter, r *http.Request, page string, pageData interface{}) {

	var err error

	// Set the last page
	if r.Method == "GET" && page != "error" && page != "login" {

		err = session.Set(r, helpers.SessionLastPage, r.URL.Path)
		if err != nil {
			log.Err(err, r)
		}
	}

	// Save the session
	err = session.Save(w, r)
	if _, ok := err.(base64.CorruptInputError); ok {
		log.Info(err)
	} else if err != nil {
		log.Err(err, r)
	}

	//
	setHeaders(w, r, "text/html")

	//
	t, err := template.New("t").Funcs(getTemplateFuncMap()).ParseFiles(
		"./templates/_webpack_header.gohtml",
		"./templates/_webpack_footer.gohtml",
		"./templates/_players_header.gohtml",
		"./templates/_header.gohtml",
		"./templates/_footer.gohtml",
		"./templates/_apps_header.gohtml",
		"./templates/_login_header.gohtml",
		"./templates/_flashes.gohtml",
		"./templates/_stats_header.gohtml",
		"./templates/_social.gohtml",
		"./templates/"+page+".gohtml",
	)
	if err != nil {
		log.Critical(err, r)
		return
	}

	// Write a respone
	buf := &bytes.Buffer{}
	err = t.ExecuteTemplate(buf, page, pageData)
	if err != nil {
		log.Critical(err, r)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "Looks like I messed something up, will be fixed soon!"})
		return
	}

	if config.IsProd() {

		m := minify.New()
		m.Add("text/html", &minhtml.Minifier{
			KeepConditionalComments: false,
			KeepDefaultAttrVals:     true,
			KeepDocumentTags:        true,
			KeepEndTags:             true,
			KeepWhitespace:          true,
		})

		err = m.Minify("text/html", w, buf)
		if err != nil && !strings.Contains(err.Error(), "write: broken pipe") {
			log.Critical(err, r)
		}

	} else {
		_, err = buf.WriteTo(w)
		log.Critical(err, r)
	}
}

func returnErrorTemplate(w http.ResponseWriter, r *http.Request, data errorTemplate) {

	if data.Title == "" {
		data.Title = "Error " + strconv.Itoa(data.Code)
	}

	if data.Code == 0 {
		data.Code = 500
	}

	log.Err(data.Error)

	data.fill(w, r, "Error", "Something has gone wrong!")

	w.WriteHeader(data.Code)

	returnTemplate(w, r, "error", data)
}

type errorTemplate struct {
	GlobalTemplate
	Title   string
	Message string
	Code    int
	Error   error
	DataID  int64 // To add to the page attribute
}

func getTemplateFuncMap() map[string]interface{} {
	return template.FuncMap{
		"join":         func(a []string) string { return strings.Join(a, ", ") },
		"lower":        func(a string) string { return strings.ToLower(a) },
		"comma":        func(a int) string { return humanize.Comma(int64(a)) },
		"comma64":      func(a int64) string { return humanize.Comma(a) },
		"commaf":       func(a float64) string { return humanize.Commaf(a) },
		"bytes":        func(a uint64) string { return humanize.Bytes(a) },
		"seconds":      func(a int64) string { return humanize.RelTime(time.Now(), time.Now().Add(time.Second*time.Duration(a)), "", "") },
		"startsWith":   func(a string, b string) bool { return strings.HasPrefix(a, b) },
		"endsWith":     func(a string, b string) bool { return strings.HasSuffix(a, b) },
		"max":          func(a int, b int) float64 { return math.Max(float64(a), float64(b)) },
		"html":         func(html string) template.HTML { return helpers.RenderHTMLAndBBCode(html) },
		"json":         func(v interface{}) (string, error) { b, err := json.Marshal(v); log.Err(err); return string(b), err },
		"title":        func(a string) string { return strings.Title(a) },
		"inc":          func(i int) int { return i + 1 },
		"ordinalComma": func(i int) string { return helpers.OrdinalComma(i) },
		"https":        func(link string) string { return strings.Replace(link, "http://", "https://", 1) },
		"htmlEscape":   func(text string) string { return html.EscapeString(text) },
		"pathEscape":   func(text string) string { return url.PathEscape(text) },
		"round":        func(i int) string { return helpers.ShortHandNumber(int64(i)) },
		"round64":      func(i int64) string { return helpers.ShortHandNumber(i) },
		"slug":         func(s string) string { return slug.Make(s) },
		"sum": func(i ...int) (total int) {
			for _, v := range i {
				total += v
			}
			return total
		},
	}
}

// GlobalTemplate is added to every other template
type GlobalTemplate struct {
	Title           string        // Page title
	Description     template.HTML // Page description
	Path            string        // URL path
	Env             string        // Environment
	CSSFiles        []Asset
	JSFiles         []Asset
	IncludeSocialJS bool
	Canonical       string
	ProductCCs      []helpers.ProductCountryCode

	Background      string
	BackgroundTitle string
	BackgroundLink  string
	backgroundSet   bool

	FlashesGood []string
	FlashesBad  []string

	UserID        int
	UserName      string
	UserProductCC helpers.ProductCountryCode
	userLevel     int // Donation level

	PlayerID   int64
	PlayerName string

	// Internal
	request   *http.Request
	response  http.ResponseWriter
	metaImage string
	toasts    []Toast
	hideAds   bool
}

func (t *GlobalTemplate) fill(w http.ResponseWriter, r *http.Request, title string, description template.HTML) {

	var err error

	t.request = r
	t.response = w

	t.Title = title + " - Game DB"
	t.Description = description
	t.Env = config.Config.Environment.Get()
	t.Path = r.URL.Path
	t.ProductCCs = helpers.GetProdCCs(true)

	val, err := session.Get(r, helpers.SessionUserID)
	if err != nil {
		log.Err(err, r)
	}

	if val != "" {
		t.UserID, err = strconv.Atoi(val)
		log.Err(err, r)
	}

	val, err = session.Get(r, helpers.SessionPlayerID)
	if err != nil {
		log.Err(err, r)
	}

	if val != "" {
		t.PlayerID, err = strconv.ParseInt(val, 10, 64)
		log.Err(err, r)
	}

	val, err = session.Get(r, helpers.SessionUserLevel)
	if err != nil {
		log.Err(err, r)
	}

	if val != "" {
		t.userLevel, err = strconv.Atoi(val)
		if err != nil {
			log.Err(err, r)
		}
	}

	t.PlayerName, err = session.Get(r, helpers.SessionPlayerName)
	if err != nil {
		log.Err(err, r)
	}

	t.UserName, err = session.Get(r, helpers.SessionPlayerName)
	if err != nil {
		log.Err(err, r)
	}

	t.UserProductCC = helpers.GetProdCC(helpers.GetProductCC(r))
	if err != nil {
		log.Err(err, r)
	}

	//
	t.setRandomBackground(true, false)
	t.setFlashes()
}

func (t *GlobalTemplate) setBackground(app sql.App, title bool, link bool) {

	t.backgroundSet = true

	if app.Background != "" {
		t.Background = app.Background
	} else {
		return
	}
	if title {
		t.BackgroundTitle = app.GetName()
	} else {
		t.BackgroundTitle = ""
	}
	if link {
		t.BackgroundLink = app.GetPath()
	} else {
		t.BackgroundLink = ""
	}
}

func (t *GlobalTemplate) setRandomBackground(title bool, link bool) {

	if t.backgroundSet {
		return
	}

	if strings.HasPrefix(t.request.URL.Path, "/admin") {
		return
	}

	popularApps, err := sql.PopularApps()
	if err != nil {
		log.Err(err, t.request)
		return
	}

	blacklist := []int{
		4000,   // Garry's Mod
		236850, // Europa Universalis IV
		289070, // Civilization VI
		431960, // Wallpaper Engine
		582010, // Monster Hunter
	}

	var filteredApps []mongo.App
	for _, app := range popularApps {
		if app.Background != "" && !helpers.SliceHasInt(blacklist, app.ID) {
			filteredApps = append(filteredApps, app)
		}
	}

	if len(filteredApps) > 0 {
		t.setBackground(filteredApps[rand.Intn(len(filteredApps))], title, link)
	}
}

func (t *GlobalTemplate) setFlashes() {

	var r = t.request
	var err error

	t.FlashesGood, err = session.GetFlashes(r, helpers.SessionGood)
	if err != nil {
		log.Err(err, r)
	}

	t.FlashesBad, err = session.GetFlashes(r, helpers.SessionBad)
	if err != nil {
		log.Err(err, r)
	}
}

func (t GlobalTemplate) GetUserJSON() string {

	stringMap := map[string]interface{}{
		"prodCC":             t.UserProductCC.ProductCode,
		"userCurrencySymbol": t.UserProductCC.Symbol,
		"toasts":             t.toasts,
		"log":                config.IsLocal() || t.IsAdmin(),
		"isLoggedIn":         t.IsLoggedIn(),
		"isProd":             config.IsProd(),
	}

	b, err := json.Marshal(stringMap)
	log.Err(err, t.request)

	return string(b)
}

func (t GlobalTemplate) GetMetaImage() (text string) {

	if t.metaImage == "" {
		return "https://gamedb.online/assets/img/sa-bg-500x500.png"
	}

	return t.metaImage
}

func (t GlobalTemplate) GetCookieFlag(key string) interface{} {

	c, err := t.request.Cookie("gamedb-session-2")

	if err == http.ErrNoCookie {
		return false
	} else if err != nil {
		log.Err(err)
		return false
	}

	c.Value, err = url.PathUnescape(c.Value)
	if err != nil {
		log.Err(err)
		return false
	}

	var vals = map[string]interface{}{}
	err = json.Unmarshal([]byte(c.Value), &vals)
	if err != nil {
		log.Err(err)
		return false
	}

	if val, ok := vals[key]; ok {
		return val
	}

	return false
}

func (t GlobalTemplate) GetCanonical() (text string) {

	if t.Canonical != "" {
		return "https://gamedb.online" + t.Canonical
	}
	return "https://gamedb.online" + t.request.URL.Path + strings.TrimRight("?"+t.request.URL.Query().Encode(), "?")
}

func (t GlobalTemplate) GetVersionHash() string {

	if config.IsLocal() {
		return "local"
	}

	if len(config.Config.CommitHash.Get()) >= 7 {
		return config.Config.CommitHash.Get()[0:7]
	}
	return ""
}

func (t GlobalTemplate) IsAppsPage() bool {
	return helpers.SliceHasString([]string{"apps", "new-releases", "upcoming", "apps/trending", "achievements", "packages", "bundles", "price-changes", "changes", "coop"}, strings.TrimPrefix(t.Path, "/"))
}

func (t GlobalTemplate) IsStatsPage() bool {
	return helpers.SliceHasString([]string{"stats", "tags", "genres", "publishers", "developers"}, strings.TrimPrefix(t.Path, "/"))
}

func (t GlobalTemplate) IsSettingsPage() bool {

	if strings.HasPrefix(t.Path, "/signup") {
		return true
	}

	return helpers.SliceHasString([]string{"login", "logout", "forgot", "settings", "admin"}, strings.TrimPrefix(t.Path, "/"))
}

func (t GlobalTemplate) IsMorePage() bool {

	if strings.HasPrefix(t.Path, "/chat") {
		return true
	}

	if strings.HasPrefix(t.Path, "/experience") {
		return true
	}

	return helpers.SliceHasString([]string{"chat-bot", "contact", "info", "queues", "info", "steam-api", "api"}, strings.TrimPrefix(t.Path, "/"))
}

func (t GlobalTemplate) IsSidebarPage() bool {
	return helpers.SliceHasString([]string{"api", "steam-api"}, strings.TrimPrefix(t.Path, "/"))
}

func (t GlobalTemplate) IsLoggedIn() bool {
	return t.UserID > 0
}

func (t GlobalTemplate) IsAdmin() bool {
	return helpers.IsAdmin(t.request)
}

func (t GlobalTemplate) ShowAds() bool {

	if config.IsLocal() {
		return false
	}

	if t.userLevel > 0 {
		return false
	}

	return !t.hideAds
}

func (t *GlobalTemplate) addToast(toast Toast) {
	t.toasts = append(t.toasts, toast)
}

func (t *GlobalTemplate) addAssetChosen() {
	t.JSFiles = append(t.JSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/chosen/1.8.7/chosen.jquery.min.js", Integrity: "sha256-c4gVE6fn+JRKMRvqjoDp+tlG4laudNYrXI1GncbfAYY="})
}

func (t *GlobalTemplate) addAssetCountdown() {
	t.JSFiles = append(t.JSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/jquery.countdown/2.2.0/jquery.countdown.min.js", Integrity: "sha256-Ikk5myJowmDQaYVCUD0Wr+vIDkN8hGI58SGWdE671A8="})
}

func (t *GlobalTemplate) addAssetJSON2HTML() {
	// This is included in webpack now
	// t.JSFiles = append(t.JSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/json2html/1.3.0/json2html.min.js", Integrity: "sha256-99iKvXmXDqqj9wm3jtTv5Iwn7MTPw4npjQhIY3gY2rw="})
}

func (t *GlobalTemplate) addAssetHighCharts() {
	t.JSFiles = append(t.JSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/highcharts/7.0.1/highcharts.js", Integrity: "sha256-j3WPKr23emLOeDVvf5mbfGs5xE+GERqV1vCz+Wx6n74="})
	t.JSFiles = append(t.JSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/highcharts/7.0.1/modules/data.js", Integrity: "sha256-CYgititANzm6qnx8M/4TpaGqfa8xFOIbHfWbtvKAg4w="})
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

//
type Asset struct {
	URL       string
	Integrity string
}

// DataTablesAjaxResponse
type DataTablesAjaxResponse struct {
	Draw            string          `json:"draw"`
	RecordsTotal    int64           `json:"recordsTotal,string"`
	RecordsFiltered int64           `json:"recordsFiltered,string"`
	LevelLimited    bool            `json:"limited"`
	Data            [][]interface{} `json:"data"`
}

func (t *DataTablesAjaxResponse) AddRow(row []interface{}) {
	t.Data = append(t.Data, row)
}

func (t DataTablesAjaxResponse) output(w http.ResponseWriter, r *http.Request) {

	if len(t.Data) == 0 {
		t.Data = make([][]interface{}, 0)
	}

	returnJSON(w, r, t)
}

func (t *DataTablesAjaxResponse) limit(r *http.Request) {

	level := sql.UserLevel(helpers.GetUserLevel(r))
	max := level.MaxResults(100)

	if max > 0 && max < t.RecordsFiltered {
		t.RecordsFiltered = max
		t.LevelLimited = true
	}
}

// DataTablesQuery
type DataTablesQuery struct {
	Draw   string                            `json:"draw"`
	Order  map[string]map[string]interface{} `json:"order"`
	Start  string                            `json:"start"`
	Search map[string]interface{}            `json:"search"`
	// Time   string `json:"_"`
	// Columns []string
}

func (q *DataTablesQuery) fillFromURL(url url.Values) (err error) {

	// Convert string into map
	queryMap, err := qs.Unmarshal(url.Encode())
	if err != nil {
		return err
	}

	// Convert map into struct
	return helpers.MarshalUnmarshal(queryMap, q)
}

func (q *DataTablesQuery) limit(r *http.Request) {

	level := sql.UserLevel(helpers.GetUserLevel(r))
	max := level.MaxOffset(100)

	start, _ := strconv.Atoi(q.Start)

	if max > 0 && int64(start) > max {
		q.Start = strconv.FormatInt(int64(start), 10)
	}
}

func (q DataTablesQuery) getSearchString(k string) (search string) {

	if val, ok := q.Search[k]; ok {
		if ok && val != "" {
			if val, ok := val.(string); ok {
				if ok {
					return val
				}
			}
		}
	}

	return ""
}

func (q DataTablesQuery) getSearchSlice(k string) (search []string) {

	if val, ok := q.Search[k]; ok {
		if val != "" {

			if val, ok := val.([]interface{}); ok {
				for _, v := range val {
					search = append(search, v.(string))
				}
			}
		}
	}

	return search
}

func (q DataTablesQuery) getOrderSQL(columns map[string]string, defaultCol string) (order string) {

	var ret []string

	for _, v := range q.Order {

		col, ok := v["column"].(string)
		if !ok || col == "" {
			col = defaultCol
		}

		if dir, ok := v["dir"].(string); ok {
			if ok {

				if columns != nil {
					col, ok := columns[col]
					if ok {
						if dir == "asc" || dir == "desc" {
							if strings.Contains(col, "$dir") {
								ret = append(ret, strings.Replace(col, "$dir", dir, 1))
							} else {
								ret = append(ret, col+" "+dir)
							}
						}
					}
				}
			}
		}
	}

	return strings.Join(ret, ", ")
}

func (q DataTablesQuery) getOrderMongo(columns map[string]string) bson.D {

	for _, v := range q.Order {

		if col, ok := v["column"].(string); ok {
			if ok {

				if dir, ok := v["dir"].(string); ok {
					if ok {

						if col, ok := columns[col]; ok {
							if ok {

								if dir == "desc" {
									return bson.D{{Key: col, Value: -1}}
								}

								return bson.D{{Key: col, Value: 1}}
							}
						}
					}
				}
			}
		}
	}

	return bson.D{}
}

func (q DataTablesQuery) getOrderString(columns map[string]string) (col string) {

	for _, v := range q.Order {

		if col, ok := v["column"].(string); ok {
			if ok {
				if col, ok := columns[col]; ok {
					if ok {
						return col
					}
				}
			}
		}
	}

	return col
}

func (q DataTablesQuery) setOrderOffsetGorm(db *gorm.DB, columns map[string]string, defaultCol string) *gorm.DB {

	db = db.Order(q.getOrderSQL(columns, defaultCol))
	db = db.Offset(q.Start)

	return db
}

func (q DataTablesQuery) getOffset() int {
	i, _ := strconv.Atoi(q.Start)
	return i
}

func (q DataTablesQuery) getOffset64() int64 {
	i, _ := strconv.ParseInt(q.Start, 10, 64)
	return i
}

func (q DataTablesQuery) getPage(perPage int) int {

	i, _ := strconv.Atoi(q.Start)

	if i == 0 {
		return 1
	}

	return (i / perPage) + 1
}

// Toasts
type Toast struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	Link    string `json:"link"`
	Theme   string `json:"theme"`
	Timeout int    `json:"timeout"`
}

func getUserFromSession(r *http.Request) (user sql.User, err error) {

	userID, err := helpers.GetUserIDFromSesion(r)
	if err != nil {
		return user, err
	}

	return sql.GetUserByID(userID)
}
