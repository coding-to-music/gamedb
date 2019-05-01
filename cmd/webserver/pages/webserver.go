package pages

import (
	"bytes"
	"encoding/json"
	"html/template"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/derekstavis/go-qs"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/cmd/webserver/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/jinzhu/gorm"
	"github.com/mitchellh/mapstructure"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"
)

func setAllowedQueries(w http.ResponseWriter, r *http.Request, allowed []string) (redirect bool) {

	if allowed == nil {
		allowed = []string{}
	}

	allowed = append(allowed, "_") // jQuery caching

	query := r.URL.Query()
	oldPath := query.Encode()

	for k := range query {
		if !helpers.SliceHasString(allowed, k) {
			query.Del(k)
		}
	}

	newPath := query.Encode()
	if oldPath != newPath {
		http.Redirect(w, r, r.URL.Path+"?"+newPath, http.StatusFound)
		return true
	}

	return false
}

func setAllHeaders(w http.ResponseWriter, r *http.Request, contentType string) {

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Language", string(session.GetCountryCode(r))) // Used for varnish hash
	w.Header().Set("X-Content-Type-Options", "nosniff")           // Protection from malicious exploitation via MIME sniffing
	w.Header().Set("X-XSS-Protection", "1; mode=block")           // Block access to the entire page when an XSS attack is suspected
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")               // Protection from clickjacking

	if !strings.HasPrefix(r.URL.Path, "/esi") {
		w.Header().Set("Surrogate-Control", "ESI/1.0") // Enable ESI
	}

	setCacheHeaders(w, time.Hour*24) // Default cache headers
}

func setCacheHeaders(w http.ResponseWriter, duration time.Duration) {

	if w.Header().Get("Cache-Control") == "" || w.Header().Get("Expires") == "" {

		if duration == 0 || config.IsLocal() {

			w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=0")
		} else {

			w.Header().Set("Cache-Control", "public, max-age=0, s-maxage="+strconv.Itoa(int(duration.Seconds())))
		}
	}
}

func returnJSON(w http.ResponseWriter, r *http.Request, bytes []byte) (err error) {

	setAllHeaders(w, r, "application/json")

	_, err = w.Write(bytes)
	return err
}

func returnTemplate(w http.ResponseWriter, r *http.Request, page string, pageData interface{}) (err error) {

	setAllHeaders(w, r, "text/html")

	folder := config.Config.TemplatesPath.Get()
	t, err := template.New("t").Funcs(getTemplateFuncMap()).ParseFiles(
		folder+"/_apps_header.gohtml",
		folder+"/_current_apps.gohtml",
		folder+"/_flashes.gohtml",
		folder+"/_footer.gohtml",
		folder+"/_header.gohtml",
		folder+"/_stats_header.gohtml",
		folder+"/_social.gohtml",
		folder+"/"+page+".gohtml",
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

	if config.IsProd() {

		m := minify.New()
		m.Add("text/html", &html.Minifier{
			KeepConditionalComments: false,
			KeepDefaultAttrVals:     true,
			KeepDocumentTags:        true,
			KeepEndTags:             true,
			KeepWhitespace:          true,
		})

		err = m.Minify("text/html", w, buf)
		if err != nil {
			log.Err(err)
			return err
		}

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

	data.fill(w, r, "Error", "Something has gone wrong!")

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
	DataID  int64 // To add to the page attribute
}

func getTemplateFuncMap() map[string]interface{} {
	return template.FuncMap{
		"join":       func(a []string) string { return strings.Join(a, ", ") },
		"lower":      func(a string) string { return strings.ToLower(a) },
		"comma":      func(a int) string { return humanize.Comma(int64(a)) },
		"comma64":    func(a int64) string { return humanize.Comma(a) },
		"commaf":     func(a float64) string { return humanize.Commaf(a) },
		"bytes":      func(a uint64) string { return humanize.Bytes(a) },
		"seconds":    func(a int64) string { return humanize.RelTime(time.Now(), time.Now().Add(time.Second*time.Duration(a)), "", "") },
		"startsWith": func(a string, b string) bool { return strings.HasPrefix(a, b) },
		"endsWith":   func(a string, b string) bool { return strings.HasSuffix(a, b) },
		"max":        func(a int, b int) float64 { return math.Max(float64(a), float64(b)) },
		"html":       func(html string) template.HTML { return helpers.RenderHTMLAndBBCode(html) },
		"json":       func(v interface{}) (string, error) { b, err := json.Marshal(v); log.Err(err); return string(b), err },
	}
}

// GlobalTemplate is added to every other template
type GlobalTemplate struct {
	// These variables can be used in templates and cached
	Title              string        // Page title
	TitleWithIcons     string        // Page title
	Description        template.HTML // Page description
	Path               string        // URL path
	Env                string        // Environment
	CSSFiles           []Asset
	JSFiles            []Asset
	UserCountry        steam.CountryCode
	UserCurrencySymbol string

	// These variables can't!
	// Session
	userName    string
	userEmail   string
	userID      int
	userLevel   int
	flashesGood []interface{}
	flashesBad  []interface{}
	session     map[string]string

	//
	toasts      []Toast
	contactPage map[string]string
	loginPage   map[string]string

	// Internal
	request   *http.Request
	metaImage string
}

func (t *GlobalTemplate) fill(w http.ResponseWriter, r *http.Request, title string, description template.HTML) {

	var err error

	t.request = r

	if helpers.IsBot(r.UserAgent()) {
		t.Title = title
		t.TitleWithIcons = title
	} else {
		t.Title = title + " - Game DB"
		t.TitleWithIcons = title + " - 🅶🅰🅼🅴 🅳🅱"
	}

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

	// Country
	t.UserCountry = steam.CountryCode(r.URL.Query().Get("cc"))

	// Check if valid country
	if _, ok := steam.Countries[t.UserCountry]; !ok {
		t.UserCountry = session.GetCountryCode(r)
	}

	// Default country to session
	if t.UserCountry == "" {
		t.UserCountry = session.GetCountryCode(r)
	}

	// Currency
	locale, err := helpers.GetLocaleFromCountry(t.UserCountry)
	log.Err(err, r)
	if err == nil {
		t.UserCurrencySymbol = locale.CurrencySymbol
	}

	// Pages
	switch t.Path {
	case "/contact":

		// Details from form
		contactName, err := session.Read(r, "contact-name")
		log.Err(err)
		contactEmail, err := session.Read(r, "contact-email")
		log.Err(err)
		contactMessage, err := session.Read(r, "contact-message")
		log.Err(err)

		t.contactPage = map[string]string{
			"name":    contactName,
			"email":   contactEmail,
			"message": contactMessage,
		}

		// Email from logged in user
		t.userEmail, err = session.Read(r, session.UserEmail)
		log.Err(err, r)

	case "/login":

		loginEmail, err := session.Read(r, "login-email")
		log.Err(err)

		t.loginPage = map[string]string{
			"email": loginEmail,
		}

	case "/experience":

		level, err := session.Read(r, session.PlayerLevel)
		log.Err(err, r)

		if level == "" {
			t.userLevel = 0
		} else {
			t.userLevel, err = strconv.Atoi(level)
			log.Err(err, r)
		}
	}
}

func (t GlobalTemplate) GetMetaImage() (text string) {

	if t.metaImage == "" {
		return "https://gamedb.online/assets/img/sa-bg-500x500.png"
	}

	return t.metaImage
}

func (t GlobalTemplate) GetCanonical() (text string) {

	return "https://gamedb.online" + t.request.URL.Path + strings.TrimRight("?"+t.request.URL.Query().Encode(), "?")
}

func (t GlobalTemplate) GetFooterText() (text template.HTML) {

	// Page created time
	text += template.HTML(`Page created <span data-livestamp="` + strconv.FormatInt(time.Now().Unix(), 10) + `"></span>.`)

	text += " All times UTC."

	// // Page load time
	// startTimeInt, err := strconv.ParseInt(t.request.Header.Get("start-time"), 10, 64)
	// log.Err(err)
	//
	// durStr, err := durationfmt.Format(time.Duration(time.Now().UnixNano()-startTimeInt), "%ims")
	// log.Err(err)
	//
	// text += template.HTML(" in " + durStr + ".")

	// Deployed commit hash
	if len(config.Config.CommitHash.Get()) >= 7 {
		text += template.HTML(` <a href="/commits">v` + config.Config.CommitHash.Get()[0:7] + `</a>.`)
	}

	return text
}

func (t GlobalTemplate) IsCacheHit() bool {
	return t.request.Header.Get("X-Cache") == "HIT"
}

func (t GlobalTemplate) IsFromVarnish() bool {
	return t.request.Header.Get("X-Cache") != ""
}

func (t GlobalTemplate) IsAppsPage() bool {
	return helpers.SliceHasString([]string{"apps", "packages", "bundles"}, strings.TrimPrefix(t.Path, "/"))
}

func (t GlobalTemplate) IsStatsPage() bool {
	return helpers.SliceHasString([]string{"stats", "tags", "genres", "publishers", "developers"}, strings.TrimPrefix(t.Path, "/"))
}

func (t GlobalTemplate) IsTrendingPage() bool {
	return helpers.SliceHasString([]string{"upcoming", "new-releases", "trending"}, strings.TrimPrefix(t.Path, "/"))
}

func (t GlobalTemplate) IsMorePage() bool {
	return helpers.SliceHasString([]string{"changes", "chat", "chat-bot", "commits", "contact", "coop", "experience", "info", "queues", "steam-api"}, strings.TrimPrefix(t.Path, "/"))
}

func (t GlobalTemplate) isLoggedIn() bool {
	return t.userID > 0
}

func (t GlobalTemplate) isLocal() bool {
	return t.Env == string(config.EnvLocal)
}

func (t GlobalTemplate) isAdmin() bool {
	return isAdmin(t.request)
}

func (t GlobalTemplate) showAds() bool {
	return false
	// return !t.isLocal()
}

func (t *GlobalTemplate) addToast(toast Toast) {
	t.toasts = append(t.toasts, toast)
}

func (t *GlobalTemplate) addAssetChosen() {
	t.JSFiles = append(t.JSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/chosen/1.8.7/chosen.jquery.min.js", Integrity: "sha256-c4gVE6fn+JRKMRvqjoDp+tlG4laudNYrXI1GncbfAYY="})
}

func (t *GlobalTemplate) addAssetJSON2HTML() {
	t.JSFiles = append(t.JSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/json2html/1.2.0/json2html.min.js", Integrity: "sha256-5iWhgkOOkWSQMxoIXqSKvZQHOTJ1wYDBqhMTFm5DkDw="})
	t.JSFiles = append(t.JSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/jquery.json2html/1.2.0/jquery.json2html.min.js", Integrity: "sha256-NVPR5gsJCl/e6xUJ3Wv2+4Tui2vhZY6KBhx0RY0DNcs="})
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

func (t *GlobalTemplate) setFlashes(w http.ResponseWriter, r *http.Request, save bool) {

	var err error

	t.flashesGood, err = session.GetGoodFlashes(w, r)
	log.Err(err, r)

	t.flashesBad, err = session.GetBadFlashes(w, r)
	log.Err(err, r)

	if save && (len(t.flashesBad) > 0 || len(t.flashesBad) > 0) {
		err = session.Save(w, r)
		log.Err(err)
	}
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

	b, err := json.Marshal(t)
	log.Err(err, r)

	err = returnJSON(w, r, b)
	log.Err(err, r)
}

// DataTablesQuery
type DataTablesQuery struct {
	Draw    string
	Order   map[string]map[string]interface{}
	Start   string
	Search  map[string]interface{}
	Time    string `mapstructure:"_"`
	Columns []string
}

func (q *DataTablesQuery) fillFromURL(url url.Values) (err error) {

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

func (q DataTablesQuery) getSearchString(k string) (search string) {

	if val, ok := q.Search[k]; ok {
		if ok && val != "" {
			return val.(string)
		}
	}

	return ""
}

func (q DataTablesQuery) getSearchSlice(k string) (search []string) {

	if val, ok := q.Search[k]; ok {
		if val != "" {
			for _, v := range val.([]interface{}) {
				search = append(search, v.(string))
			}
		}
	}

	return search
}

func (q DataTablesQuery) getOrderSQL(columns map[string]string, code steam.CountryCode) (order string) {

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

func (q DataTablesQuery) getOrderMongo(columns map[string]string, colEdit func(string) string) mongo.D {

	for _, v := range q.Order {

		if col, ok := v["column"].(string); ok {
			if ok {

				if dir, ok := v["dir"].(string); ok {
					if ok {

						if col, ok := columns[col]; ok {
							if ok {

								if colEdit != nil {
									col = colEdit(col)
								}

								if dir == "desc" {
									return mongo.D{{col, -1}}
								} else {
									return mongo.D{{col, 1}}
								}
							}
						}
					}
				}
			}
		}
	}

	return mongo.D{}
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

func (q DataTablesQuery) setOrderOffsetGorm(db *gorm.DB, code steam.CountryCode, columns map[string]string) *gorm.DB {

	db = db.Order(q.getOrderSQL(columns, code))
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

	return int(i/perPage) + 1
}

// Toasts
type Toast struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	Link    string `json:"link"`
	Theme   string `json:"theme"`
	Timeout int    `json:"timeout"`
}

//
func isAdmin(r *http.Request) bool {

	id, err := session.Read(r, session.PlayerID)
	log.Err(err)

	return id == config.Config.AdminSteamID.Get()
}

//
func getPlayerIDFromSession(r *http.Request) (playerID int64, err error) {

	// Check if logged in
	loggedIn, err := session.IsLoggedIn(r)
	if err != nil {
		return playerID, errNotLoggedIn
	}

	if !loggedIn {
		return playerID, errNotLoggedIn
	}

	// Get session
	id, err := session.Read(r, session.PlayerID)
	if err != nil {
		return playerID, err
	}

	// Convert ID
	return strconv.ParseInt(id, 10, 64)
}

func getPlayerFromSession(r *http.Request) (player mongo.Player, err error) {

	playerID, err := getPlayerIDFromSession(r)
	if err != nil {
		return player, err
	}

	return mongo.GetPlayer(playerID)
}

func getUserFromSession(r *http.Request) (user sql.User, err error) {

	playerID, err := getPlayerIDFromSession(r)
	if err != nil {
		return user, err
	}

	return sql.GetOrCreateUser(playerID)
}
