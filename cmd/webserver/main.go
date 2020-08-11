package main

import (
	"compress/flate"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/recaptcha-go"
	"github.com/gamedb/gamedb/cmd/webserver/pages"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/middleware"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/go-chi/chi"
	chiMiddleware "github.com/go-chi/chi/middleware"
	"github.com/gobuffalo/packr/v2"
)

var (
	version string
	commits string

	distBox  = packr.New("dist", "./assets/dist")
	filesBox = packr.New("files", "./assets/files")
	imgBox   = packr.New("img", "./assets/img")
)

func main() {

	config.Init(version, commits, helpers.GetIP())
	log.Initialise(log.LogNameWebserver)

	//
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		log.Critical("GOOGLE_APPLICATION_CREDENTIALS not found")
		os.Exit(1)
	}

	// Profiling
	log.Info("Starting webserver profiling")
	go func() {
		err := http.ListenAndServe(":6064", nil)
		log.Critical(err)
	}()

	// Get API key
	err := mysql.GetConsumer("webserver")
	if err != nil {
		log.Critical(err)
		return
	}

	// Start queue producers to send to.
	// In a go routine so if Rabbit is not working, the webserver still starts
	go queue.Init(queue.WebserverDefinitions)

	// Setup Recaptcha
	recaptcha.SetSecret(config.Config.RecaptchaPrivate.Get())

	session.InitSession()

	// Clear caches on process restart
	if config.IsProd() {
		keys := []string{
			memcache.MemcacheCommitsPage(1).Key,
		}
		err = memcache.Delete(keys...)
		log.Err(err)
	}

	// Routes
	r := chi.NewRouter()
	r.Use(chiMiddleware.RedirectSlashes)
	r.Use(middleware.MiddlewareDownMessage)
	r.Use(middleware.MiddlewareCors())
	r.Use(middleware.MiddlewareRealIP)
	r.Use(chiMiddleware.NewCompressor(flate.DefaultCompression, "text/html", "text/css", "text/javascript", "application/json", "application/javascript").Handler)

	// Pages
	r.Get("/", pages.HomeHandler)
	r.Get("/currency/{id}", pages.CurrencyHandler)
	r.Mount("/achievements", pages.AchievementsRouter())
	r.Mount("/admin", pages.AdminRouter())
	r.Mount("/api", pages.APIRouter())
	r.Mount("/badges", pages.BadgesRouter())
	r.Mount("/bundles", pages.BundlesRouter())
	r.Mount("/categories", pages.CategoriesRouter())
	r.Mount("/changes", pages.ChangesRouter())
	r.Mount("/commits", pages.CommitsRouter())
	r.Mount("/contact", pages.ContactRouter())
	r.Mount("/depots", pages.DepotsRouter())
	r.Mount("/developers", pages.DevelopersRouter())
	r.Mount("/discord-bot", pages.ChatBotRouter())
	r.Mount("/discord-server", pages.ChatRouter())
	r.Mount("/donate", pages.DonateRouter())
	r.Mount("/experience", pages.ExperienceRouter())
	r.Mount("/forgot", pages.ForgotRouter())
	r.Mount("/franchise", pages.FranchiseRouter())
	r.Mount("/games", pages.GamesRouter())
	r.Mount("/genres", pages.GenresRouter())
	r.Mount("/groups", pages.GroupsRouter())
	r.Mount("/health-check", pages.HealthCheckRouter())
	r.Mount("/home", pages.HomeRouter())
	r.Mount("/info", pages.InfoRouter())
	r.Mount("/login", pages.LoginRouter())
	r.Mount("/logout", pages.LogoutRouter())
	r.Mount("/news", pages.NewsRouter())
	r.Mount("/packages", pages.PackagesRouter())
	r.Mount("/players", pages.PlayersRouter())
	r.Mount("/price-changes", pages.PriceChangeRouter())
	r.Mount("/product-keys", pages.ProductKeysRouter())
	r.Mount("/publishers", pages.PublishersRouter())
	r.Mount("/queues", pages.QueuesRouter())
	r.Mount("/settings", pages.SettingsRouter())
	r.Mount("/signup", pages.SignupRouter())
	r.Mount("/stats", pages.StatsRouter())
	r.Mount("/tags", pages.TagsRouter())
	r.Mount("/terms", pages.TermsRouter())
	r.Mount("/webhooks", pages.WebhooksRouter())
	r.Mount("/websocket", pages.WebsocketsRouter())

	// Assets
	r.Route("/assets", func(r chi.Router) {
		r.Get("/img/*", rootFileHandler(imgBox, "/assets/img"))
		r.Get("/files/*", rootFileHandler(filesBox, "/assets/files"))
		r.Get("/dist/*", rootFileHandler(distBox, "/assets/dist"))
	})

	// Sitemaps, Google doesnt like having a sitemap in a sub directory
	r.Get("/sitemap-badges.xml", pages.SiteMapBadges)
	r.Get("/sitemap-games-by-players.xml", pages.SiteMapGamesByPlayersHandler)
	r.Get("/sitemap-games-by-score.xml", pages.SiteMapGamesByScoreHandler)
	r.Get("/sitemap-games-new.xml", pages.SiteMapGamesNewHandler)
	r.Get("/sitemap-games-upcoming.xml", pages.SiteMapGamesUpcomingHandler)
	r.Get("/sitemap-groups.xml", pages.SiteMapGroups)
	r.Get("/sitemap-pages.xml", pages.SiteMapPagesHandler)
	r.Get("/sitemap-players-by-games.xml", pages.SiteMapPlayersByGamesCount)
	r.Get("/sitemap-players-by-level.xml", pages.SiteMapPlayersByLevel)
	r.Get("/sitemap.xml", pages.SiteMapIndexHandler)

	// Root files
	r.Get("/browserconfig.xml", rootFileHandler(filesBox, ""))
	r.Get("/robots.txt", rootFileHandler(filesBox, ""))
	r.Get("/site.webmanifest", rootFileHandler(filesBox, ""))
	// r.Get("/ads.txt", rootFileHandler)

	// Shortcuts
	r.Get("/a{id:[0-9]+}", redirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/games/" + chi.URLParam(r, "id") }))
	r.Get("/g{id:[0-9]+}", redirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/games/" + chi.URLParam(r, "id") }))
	r.Get("/s{id:[0-9]+}", redirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/packages/" + chi.URLParam(r, "id") }))
	r.Get("/p{id:[0-9]+}", redirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/players/" + chi.URLParam(r, "id") }))
	r.Get("/b{id:[0-9]+}", redirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/bundles/" + chi.URLParam(r, "id") }))

	// Redirects
	r.Get("/apps", redirectHandler("/games"))
	r.Get("/apps/{id:[0-9]+}", redirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/games/" + chi.URLParam(r, "id") }))
	r.Get("/apps/{id:[0-9]+}/{slug}", redirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/games/" + chi.URLParam(r, "id") + "/" + chi.URLParam(r, "slug") }))
	r.Get("/subs", redirectHandler("/packages"))
	r.Get("/subs/{id:[0-9]+}", redirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/packages/" + chi.URLParam(r, "id") }))
	r.Get("/subs/{id:[0-9]+}/{slug}", redirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/packages/" + chi.URLParam(r, "id") + "/" + chi.URLParam(r, "slug") }))
	r.Get("/coop", redirectHandler("/games/coop"))
	r.Get("/chat", redirectHandler("/discord-server"))
	r.Get("/chat-bot", redirectHandler("/discord-bot"))
	r.Get("/chat/{id}", redirectHandler("/discord-server"))
	r.Get("/sitemap/index.xml", redirectHandler("/sitemap.xml"))
	r.Get("/steam-api", redirectHandler("/api/steam"))
	r.Get("/api", redirectHandler("/api/gamedb"))

	// Game Redirects
	r.Get("/new-releases", redirectHandler("/games/new-releases"))
	r.Get("/random", redirectHandler("/games/random"))
	r.Get("/sales", redirectHandler("/games/sales"))
	r.Get("/trending", redirectHandler("/games/trending"))
	r.Get("/upcoming", redirectHandler("/games/upcoming"))
	r.Get("/wishlists", redirectHandler("/games/wishlists"))

	// 404
	r.NotFound(pages.Error404Handler)

	log.Info("Starting Webserver on " + "http://" + config.ListenOn())

	s := &http.Server{
		Addr:              config.ListenOn(),
		Handler:           r,
		ReadTimeout:       2 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	err = s.ListenAndServe()
	log.Critical(err)
}

func redirectHandler(path string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		u, _ := url.Parse(path)
		q := u.Query()

		for k, v := range r.URL.Query() {
			for _, vv := range v {
				q.Add(k, vv)
			}
		}

		u.RawQuery = q.Encode()

		http.Redirect(w, r, u.String(), http.StatusFound)
	}
}

func redirectHandlerFunc(f func(w http.ResponseWriter, r *http.Request) string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		u, _ := url.Parse(f(w, r))
		q := u.Query()

		for k, v := range r.URL.Query() {
			for _, vv := range v {
				q.Add(k, vv)
			}
		}

		u.RawQuery = q.Encode()

		http.Redirect(w, r, u.String(), http.StatusFound)
	}
}

func rootFileHandler(box *packr.Box, path string) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("X-Content-Type-Options", "nosniff")

		b, err := box.Find(strings.TrimPrefix(r.URL.Path, path))
		if err != nil {
			w.WriteHeader(404)
			_, err := w.Write([]byte("Unable to read file."))
			log.Err(err, r)
			return
		}

		types := map[string]string{
			".js":  "text/javascript",
			".css": "text/css",
			".png": "image/png",
			".jpg": "image/jpeg",
		}

		if val, ok := types[filepath.Ext(r.URL.Path)]; ok {

			// Cache for ages
			duration := time.Hour * 1000
			w.Header().Set("Cache-Control", "max-age="+strconv.Itoa(int(duration.Seconds())))
			w.Header().Set("Expires", time.Now().Add(duration).Format(time.RFC1123))

			// Fix headers, packr seems to break them
			w.Header().Add("Content-Type", val)
		}

		// Output
		_, err = w.Write(b)
		log.Err(err, r)
	}
}
