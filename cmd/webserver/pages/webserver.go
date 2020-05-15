package pages

import (
	"bytes"
	"encoding/json"
	"html"
	"html/template"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/session-go/session"
	"github.com/dustin/go-humanize"
	sessionHelpers "github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/i18n"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gobuffalo/packr/v2"
	"github.com/gosimple/slug"
	"github.com/tdewolff/minify/v2"
	minhtml "github.com/tdewolff/minify/v2/html"
)

func setHeaders(w http.ResponseWriter, contentType string) {

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

func returnJSON(w http.ResponseWriter, r *http.Request, i interface{}) {

	setHeaders(w, "application/json")

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

var templatesBox = packr.New("templates", "../templates")

func returnTemplate(w http.ResponseWriter, r *http.Request, page string, pageData interface{}) {

	var err error

	// Set the last page
	if r.Method == "GET" && page != "error" && page != "login" && page != "forgot" {

		err = session.Set(r, sessionHelpers.SessionLastPage, r.URL.Path)
		if err != nil {
			log.Err(err, r)
		}
	}

	// Save the session
	sessionHelpers.Save(w, r)

	//
	setHeaders(w, "text/html")

	//
	t := template.New("t")
	t = t.Funcs(getTemplateFuncMap())

	templates := []string{
		"_webpack_header.gohtml",
		"_webpack_footer.gohtml",
		"_players_header.gohtml",
		"_header.gohtml",
		"_footer.gohtml",
		"_apps_header.gohtml",
		"_login_header.gohtml",
		"_flashes.gohtml",
		"_stats_header.gohtml",
		"_social.gohtml",
		page + ".gohtml",
	}

	for _, v := range templates {

		s, err := templatesBox.FindString(v)
		if err != nil {
			log.Err(err, r)
			continue
		}

		t, err = t.Parse(s)
		if err != nil {
			log.Err(err, r)
			continue
		}
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

func returnErrorTemplate(w http.ResponseWriter, r *http.Request, t errorTemplate) {

	if t.Title == "" {
		t.Title = "Error " + strconv.Itoa(t.Code)
	}

	if t.Code == 0 {
		t.Code = 500
	}

	t.fill(w, r, "Error", "Something has gone wrong!")

	w.WriteHeader(t.Code)

	returnTemplate(w, r, "error", t)
}

type errorTemplate struct {
	GlobalTemplate
	Title   string
	Message string
	Code    int
	DataID  int64 // To add to the page attribute
}

func getTemplateFuncMap() map[string]interface{} {
	return template.FuncMap{
		"bytes":        func(a uint64) string { return humanize.Bytes(a) },
		"comma":        func(a int) string { return humanize.Comma(int64(a)) },
		"comma64":      func(a int64) string { return humanize.Comma(a) },
		"commaf":       func(a float64) string { return humanize.FormatFloat("#,###.##", a) },
		"endsWith":     func(a string, b string) bool { return strings.HasSuffix(a, b) },
		"html":         func(html string) template.HTML { return helpers.RenderHTMLAndBBCode(html) },
		"htmlEscape":   func(text string) string { return html.EscapeString(text) },
		"https":        func(link string) string { return strings.Replace(link, "http://", "https://", 1) },
		"inc":          func(i int) int { return i + 1 },
		"join":         func(a []string) string { return strings.Join(a, ", ") },
		"json":         func(v interface{}) (string, error) { b, err := json.Marshal(v); log.Err(err); return string(b), err },
		"lower":        func(a string) string { return strings.ToLower(a) },
		"max":          func(a int, b int) float64 { return math.Max(float64(a), float64(b)) },
		"ordinalComma": func(i int) string { return helpers.OrdinalComma(i) },
		"pathEscape":   func(text string) string { return url.PathEscape(text) },
		"percent":      func(small, big int) float64 { return float64(small) / float64(big) * 100 },
		"round":        func(i int) string { return helpers.ShortHandNumber(int64(i)) },
		"round64":      func(i int64) string { return helpers.ShortHandNumber(i) },
		"seconds":      func(a int64) string { return humanize.RelTime(time.Now(), time.Now().Add(time.Second*time.Duration(a)), "", "") },
		"slug":         func(s string) string { return slug.Make(s) },
		"startsWith":   func(a string, b string) bool { return strings.HasPrefix(a, b) },
		"title":        func(a string) string { return strings.Title(a) },
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
	ProductCCs      []i18n.ProductCountryCode
	Continents      []i18n.Continent
	CurrentCC       string

	Background      string
	BackgroundTitle string
	BackgroundLink  string
	backgroundSet   bool

	FlashesGood []string
	FlashesBad  []string

	UserID        int
	UserName      string
	UserProductCC i18n.ProductCountryCode
	userLevel     int // Donation level of logged in user

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
	t.ProductCCs = i18n.GetProdCCs(true)
	t.Continents = i18n.Continents

	userIDString := sessionHelpers.Get(r, sessionHelpers.SessionUserID)
	if userIDString != "" {
		t.UserID, err = strconv.Atoi(userIDString)
		log.Err(err, r)
	}

	playerIDString := sessionHelpers.Get(r, sessionHelpers.SessionPlayerID)
	if playerIDString != "" {
		t.PlayerID, err = strconv.ParseInt(playerIDString, 10, 64)
		log.Err(err, r)
	}

	userLevel := sessionHelpers.Get(r, sessionHelpers.SessionUserLevel)
	if userLevel != "" {
		t.userLevel, err = strconv.Atoi(userLevel)
		if err != nil {
			log.Err(err, r)
		}
	}

	t.PlayerName = sessionHelpers.Get(r, sessionHelpers.SessionPlayerName)
	t.UserName = sessionHelpers.Get(r, sessionHelpers.SessionPlayerName)
	t.UserProductCC = i18n.GetProdCC(sessionHelpers.GetProductCC(r))
	t.CurrentCC = sessionHelpers.GetCountryCode(r)

	//
	t.setRandomBackground(true, false)
	t.setFlashes()
}

func (t *GlobalTemplate) setBackground(app mongo.App, title bool, link bool) {

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

	if t.request != nil && strings.HasPrefix(t.request.URL.Path, "/admin") {
		return
	}

	popularApps, err := mongo.PopularApps()
	if err != nil {
		log.Err(err, t.request)
		return
	}

	blacklist := []int{
		4000,   // Garry's Mod
		236850, // Europa Universalis IV
		242760, // The Forest
		431960, // Wallpaper Engine
	}

	whitelist := []mongo.App{
		{ID: 257420, Name: "Serious Sam 4", Background: "https://steamcdn-a.akamaihd.net/steam/apps/257420/library_hero.jpg"},
	}

	var filteredApps []mongo.App
	for _, app := range popularApps {
		if app.Background != "" && !helpers.SliceHasInt(blacklist, app.ID) {
			filteredApps = append(filteredApps, app)
		}
	}

	filteredApps = append(filteredApps, whitelist...)

	if len(filteredApps) > 0 {
		t.setBackground(filteredApps[rand.Intn(len(filteredApps))], title, link)
	}
}

func (t *GlobalTemplate) setFlashes() {

	t.FlashesGood = sessionHelpers.GetFlashes(t.request, sessionHelpers.SessionGood)
	t.FlashesBad = sessionHelpers.GetFlashes(t.request, sessionHelpers.SessionBad)
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

func (t GlobalTemplate) GetEventBadges() (badges []mongo.PlayerBadge) {

	for _, v := range mongo.GlobalBadges {
		if v.AppID == 0 {
			badges = append(badges, v)
		}
	}

	sort.Slice(badges, func(i, j int) bool {
		return badges[i].GetUniqueID() > badges[j].GetUniqueID()
	})

	return badges[0:3]
}

func (t GlobalTemplate) GetAppBadges() (badges []mongo.PlayerBadge) {

	for _, v := range mongo.GlobalBadges {
		if v.AppID > 0 {
			badges = append(badges, v)
		}
	}

	sort.Slice(badges, func(i, j int) bool {
		return badges[i].GetUniqueID() > badges[j].GetUniqueID()
	})

	return badges[0:3]
}

func (t GlobalTemplate) GetCookieFlag(key string) interface{} {

	c, err := t.request.Cookie("gamedb-session-2")

	if err == http.ErrNoCookie {
		return false
	} else if err != nil {
		log.Err(err, t.request)
		return false
	}

	c.Value, err = url.PathUnescape(c.Value)
	if err != nil {
		log.Err(err, t.request)
		return false
	}

	var vals = map[string]interface{}{}
	err = json.Unmarshal([]byte(c.Value), &vals)
	if err != nil {
		log.Err(err, t.request)
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
	return config.GetShortCommitHash()
}

func (t GlobalTemplate) IsAppsPage() bool {

	if strings.HasPrefix(t.Path, "/apps") {
		return true
	}
	return helpers.SliceHasString(strings.TrimPrefix(t.Path, "/"), []string{"new-releases", "upcoming", "wishlists", "packages", "bundles", "price-changes", "changes", "coop", "sales"})
}

func (t GlobalTemplate) IsStatsPage() bool {
	return helpers.SliceHasString(strings.TrimPrefix(t.Path, "/"), []string{"stats", "tags", "genres", "publishers", "developers"})
}

func (t GlobalTemplate) IsBadgesPage() bool {

	return strings.HasPrefix(t.Path, "/badges")
}

func (t GlobalTemplate) IsPlayersPage() bool {

	return strings.HasPrefix(t.Path, "/players")
}

func (t GlobalTemplate) IsSettingsPage() bool {

	if strings.HasPrefix(t.Path, "/signup") {
		return true
	}
	return helpers.SliceHasString(strings.TrimPrefix(t.Path, "/"), []string{"login", "logout", "forgot", "settings", "admin"})
}

func (t GlobalTemplate) IsMorePage() bool {

	if strings.HasPrefix(t.Path, "/chat") || strings.HasPrefix(t.Path, "/experience") {
		return true
	}
	return helpers.SliceHasString(strings.TrimPrefix(t.Path, "/"), []string{"achievements", "chat-bot", "contact", "info", "queues", "info", "steam-api", "api"})
}

func (t GlobalTemplate) IsSidebarPage() bool {
	return helpers.SliceHasString(strings.TrimPrefix(t.Path, "/"), []string{"api", "steam-api"})
}

func (t GlobalTemplate) IsLoggedIn() bool {
	return t.UserID > 0
}

func (t GlobalTemplate) IsAdmin() bool {
	return sessionHelpers.IsAdmin(t.request)
}

func (t GlobalTemplate) ShowAds() bool {

	if config.IsLocal() || t.userLevel > 0 {
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
	t.JSFiles = append(t.JSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/json2html/1.4.1/json2html.min.js", Integrity: "sha256-p1nDDwdo8QAOGc0Na5bpN1xNIXRxOZ6Pkm/7RkuGEK0="})
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

func (t *GlobalTemplate) addAssetMark() {
	t.JSFiles = append(t.JSFiles, Asset{URL: "https://cdnjs.cloudflare.com/ajax/libs/mark.js/8.11.1/jquery.mark.min.js", Integrity: "sha256-4HLtjeVgH0eIB3aZ9mLYF6E8oU5chNdjU6p6rrXpl9U="})
}

//
type Asset struct {
	URL       string
	Integrity string
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

	userID, err := sessionHelpers.GetUserIDFromSesion(r)
	if err != nil {
		return user, err
	}

	return sql.GetUserByID(userID)
}

// App bits
func GetAppTags(app mongo.App) (tags []sql.Tag, err error) {

	tags = []sql.Tag{} // Needed for marshalling into type

	if len(app.Tags) == 0 {
		return tags, nil
	}

	var item = memcache.MemcacheAppTags(app.ID)

	err = memcache.GetSetInterface(item.Key, item.Expiration, &tags, func() (interface{}, error) {
		return sql.GetTagsByID(app.Tags, []string{"id", "name"})
	})

	return tags, err
}

func GetAppGenres(app mongo.App) (genres []sql.Genre, err error) {

	genres = []sql.Genre{} // Needed for marshalling into type

	if len(app.Genres) == 0 {
		return genres, nil
	}

	var item = memcache.MemcacheAppGenres(app.ID)

	err = memcache.GetSetInterface(item.Key, item.Expiration, &genres, func() (interface{}, error) {
		return sql.GetGenresByID(app.Genres, []string{"id", "name"})
	})

	return genres, err
}

func GetDevelopers(app mongo.App) (developers []sql.Developer, err error) {

	developers = []sql.Developer{} // Needed for marshalling into type

	if len(app.Developers) == 0 {
		return developers, nil
	}

	var item = memcache.MemcacheAppDevelopers(app.ID)

	err = memcache.GetSetInterface(item.Key, item.Expiration, &developers, func() (interface{}, error) {
		return sql.GetDevelopersByID(app.Developers, []string{"id", "name"})
	})

	return developers, err
}

func GetPublishers(app mongo.App) (publishers []sql.Publisher, err error) {

	publishers = []sql.Publisher{} // Needed for marshalling into type

	if len(app.Publishers) == 0 {
		return publishers, nil
	}

	var item = memcache.MemcacheAppPublishers(app.ID)

	err = memcache.GetSetInterface(item.Key, item.Expiration, &publishers, func() (interface{}, error) {
		return sql.GetPublishersByID(app.Publishers, []string{"id", "name"})
	})

	return publishers, err
}

func GetAppCategories(app mongo.App) (categories []sql.Category, err error) {

	categories = []sql.Category{} // Needed for marshalling into type

	if len(app.Categories) == 0 {
		return categories, nil
	}

	var item = memcache.MemcacheAppCategories(app.ID)

	err = memcache.GetSetInterface(item.Key, item.Expiration, &categories, func() (interface{}, error) {

		return sql.GetCategoriesByID(app.Categories, []string{"id", "name"})
	})

	return categories, err
}

func GetAppBundles(app mongo.App) (bundles []sql.Bundle, err error) {

	bundles = []sql.Bundle{} // Needed for marshalling into type

	if len(app.Bundles) == 0 {
		return bundles, nil
	}

	var item = memcache.MemcacheAppBundles(app.ID)

	err = memcache.GetSetInterface(item.Key, item.Expiration, &bundles, func() (interface{}, error) {
		return sql.GetBundlesByID(app.Bundles, []string{})
	})

	return bundles, err
}

// Package bits
func GetPackageBundles(pack mongo.Package) (bundles []sql.Bundle, err error) {

	bundles = []sql.Bundle{} // Needed for marshalling into type

	if len(pack.Bundles) == 0 {
		return bundles, nil
	}

	var item = memcache.MemcachePackageBundles(pack.ID)

	err = memcache.GetSetInterface(item.Key, item.Expiration, &bundles, func() (interface{}, error) {
		return sql.GetBundlesByID(pack.Bundles, []string{})
	})

	return bundles, err
}
