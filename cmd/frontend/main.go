package main

import (
	"compress/flate"
	"errors"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gamedb/gamedb/cmd/frontend/handlers"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/email"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/middleware"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/go-chi/chi"
	chiMiddleware "github.com/go-chi/chi/middleware"
	"github.com/gobuffalo/packr/v2"
)

var (
	filesBox = packr.New("files", "./assets")
	distBox  = packr.New("dist", "./assets/dist")
	imgBox   = packr.New("img", "./assets/img")
)

func main() {

	rand.Seed(time.Now().Unix())

	err := config.Init(helpers.GetIP())
	log.InitZap(log.LogNameFrontend)
	defer log.Flush()
	if err != nil {
		log.ErrS(err)
		return
	}

	if config.C.MailjetPublic == "" || config.C.MailjetPrivate == "" {
		log.ErrS(errors.New("missing mailjet environment variables"))
		return
	}

	// Profiling
	go func() {
		err := http.ListenAndServe(":6064", nil)
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Get API key
	err = mysql.GetConsumer("frontend")
	if err != nil {
		log.ErrS(err)
		return
	}

	// Init modules
	queue.Init(queue.FrontendDefinitions)
	session.Init()
	handlers.Init()
	email.Init()

	// Clear caches on process restart
	if config.IsProd() {
		keys := []string{
			memcache.ItemCommitsPage(1).Key,
		}
		err = memcache.Delete(keys...)
		if err != nil {
			log.ErrS(err)
		}
	}

	// Routes
	r := chi.NewRouter()
	r.Use(chiMiddleware.RedirectSlashes)
	r.Use(middleware.MiddlewareDownMessage)
	r.Use(middleware.MiddlewareCors())
	r.Use(middleware.RealIP)
	r.Use(chiMiddleware.NewCompressor(flate.DefaultCompression, "text/html", "text/css", "text/javascript", "application/json", "application/javascript").Handler)
	r.Use(middleware.RateLimiterWait(time.Second, 10))

	// Pages
	r.Mount("/{type:(categories|developers|genres|publishers|tags)}", handlers.StatsListRouter())
	r.Mount("/achievements", handlers.AchievementsRouter())
	r.Mount("/admin", handlers.AdminRouter())
	r.Mount("/api", handlers.APIRouter())
	r.Mount("/badges", handlers.BadgesRouter())
	r.Mount("/bundles", handlers.BundlesRouter())
	r.Mount("/changes", handlers.ChangesRouter())
	r.Mount("/commits", handlers.CommitsRouter())
	r.Mount("/contact", handlers.ContactRouter())
	r.Mount("/discord-bot", handlers.ChatBotRouter())
	r.Mount("/discord-server", handlers.ChatRouter())
	r.Mount("/donate", handlers.DonateRouter())
	r.Mount("/experience", handlers.ExperienceRouter())
	r.Mount("/features", handlers.FeaturesRouter())
	r.Mount("/forgot", handlers.ForgotRouter())
	r.Mount("/franchise", handlers.FranchiseRouter())
	r.Mount("/games", handlers.GamesRouter())
	r.Mount("/groups", handlers.GroupsRouter())
	r.Mount("/health-check", handlers.HealthCheckRouter())
	r.Mount("/home", handlers.HomeRouter())
	r.Mount("/info", handlers.InfoRouter())
	r.Mount("/login", handlers.LoginRouter())
	r.Mount("/logout", handlers.LogoutRouter())
	r.Mount("/news", handlers.NewsRouter())
	r.Mount("/oauth", handlers.OauthRouter())
	r.Mount("/packages", handlers.PackagesRouter())
	r.Mount("/players", handlers.PlayersRouter())
	r.Mount("/price-changes", handlers.PriceChangeRouter())
	r.Mount("/product-keys", handlers.ProductKeysRouter())
	r.Mount("/queues", handlers.QueuesRouter())
	r.Mount("/settings", handlers.SettingsRouter())
	r.Mount("/signup", handlers.SignupRouter())
	r.Mount("/stats", handlers.StatsRouter())
	r.Mount("/terms", handlers.TermsRouter())
	r.Mount("/webhooks", handlers.WebhooksRouter())
	r.Mount("/websocket", handlers.WebsocketsRouter())

	r.Get("/", handlers.HomeHandler)
	r.Get("/currency/{id}", handlers.CurrencyHandler)

	// Sitemaps, Google doesnt like having a sitemap in a sub directory
	r.Get("/sitemap-badges.xml", handlers.SiteMapBadges)
	r.Get("/sitemap-games-by-players.xml", handlers.SiteMapGamesByPlayersHandler)
	r.Get("/sitemap-games-by-score.xml", handlers.SiteMapGamesByScoreHandler)
	r.Get("/sitemap-games-new.xml", handlers.SiteMapGamesNewHandler)
	r.Get("/sitemap-games-upcoming.xml", handlers.SiteMapGamesUpcomingHandler)
	r.Get("/sitemap-groups.xml", handlers.SiteMapGroups)
	r.Get("/sitemap-pages.xml", handlers.SiteMapPagesHandler)
	r.Get("/sitemap-players-by-games.xml", handlers.SiteMapPlayersByGamesCount)
	r.Get("/sitemap-players-by-level.xml", handlers.SiteMapPlayersByLevel)
	r.Get("/sitemap.xml", handlers.SiteMapIndexHandler)

	// Assets
	r.Route("/assets", func(r chi.Router) {
		r.Get("/img/*", rootFileHandler(imgBox, "/assets/img"))
		r.Get("/dist/*", rootFileHandler(distBox, "/assets/dist"))
	})

	// Root files
	r.Get("/ads.txt", rootFileHandler(filesBox, ""))
	r.Get("/browserconfig.xml", rootFileHandler(filesBox, ""))
	r.Get("/robots.txt", rootFileHandler(filesBox, ""))
	r.Get("/site.webmanifest", rootFileHandler(filesBox, ""))

	// Shortcuts
	r.Get("/a{id:[0-9]+}", redirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/games/" + chi.URLParam(r, "id") }))
	r.Get("/g{id:[0-9]+}", redirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/games/" + chi.URLParam(r, "id") }))
	r.Get("/s{id:[0-9]+}", redirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/packages/" + chi.URLParam(r, "id") }))
	r.Get("/p{id:[0-9]+}", redirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/players/" + chi.URLParam(r, "id") }))
	r.Get("/b{id:[0-9]+}", redirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/bundles/" + chi.URLParam(r, "id") }))

	// Redirects
	r.Get("/apps", redirectHandler("/games"))
	r.Get("/apps/{one}", redirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/games/" + chi.URLParam(r, "one") }))
	r.Get("/apps/{one}/{two}", redirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/games/" + chi.URLParam(r, "one") + "/" + chi.URLParam(r, "two") }))
	r.Get("/subs", redirectHandler("/packages"))
	r.Get("/subs/{one}", redirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/packages/" + chi.URLParam(r, "one") }))
	r.Get("/subs/{one}/{two}", redirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/packages/" + chi.URLParam(r, "one") + "/" + chi.URLParam(r, "two") }))
	r.Get("/coop", redirectHandler("/games/coop"))
	r.Get("/chat", redirectHandler("/discord-server"))
	r.Get("/chat-bot", redirectHandler("/discord-bot"))
	r.Get("/chat/{id}", redirectHandler("/discord-server"))
	r.Get("/sitemap/index.xml", redirectHandler("/sitemap.xml"))
	r.Get("/steam-api", redirectHandler("/api/steam"))
	r.Get("/api", redirectHandler("/api/gamedb"))
	r.Get("/discord", redirectHandler("/discord-bot")) // Used in discord messages
	r.Get("/api/gamedb", redirectHandler("/api/globalsteam"))
	r.Get("/api/gamedb.json", redirectHandler("/api/globalsteam.json"))
	r.Get("/api/gamedb.yaml", redirectHandler("/api/globalsteam.yaml"))

	// Game Redirects
	r.Get("/new-releases", redirectHandler("/games/new-releases"))
	r.Get("/random", redirectHandler("/games/random"))
	r.Get("/sales", redirectHandler("/games/sales"))
	r.Get("/trending", redirectHandler("/games/trending"))
	r.Get("/upcoming", redirectHandler("/games/upcoming"))
	r.Get("/wishlists", redirectHandler("/games/wishlists"))

	// 404
	r.NotFound(handlers.Error404Handler)

	// Serve
	if config.C.FrontendPort == "" {
		log.Err("Missing environment variables")
		return
	}

	s := &http.Server{
		Addr:              "0.0.0.0:" + config.C.FrontendPort,
		Handler:           r,
		ReadTimeout:       2 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	log.Info("Starting Frontend on " + "http://" + s.Addr)

	go func() {
		err = s.ListenAndServe()
		if err != nil {
			log.ErrS(err)
		}
	}()

	helpers.KeepAlive(
		mysql.Close,
		mongo.Close,
		func() {
			influx.GetWriter().Flush()
		})
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
			if err != nil {
				log.ErrS(err)
			}
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
		if err != nil {
			log.ErrS(err)
		}
	}
}
