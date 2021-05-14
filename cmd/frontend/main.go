package main

import (
	"compress/flate"
	"errors"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/Jleagle/rate-limit-go"
	"github.com/gamedb/gamedb/cmd/frontend/handlers"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/email"
	handlers2 "github.com/gamedb/gamedb/cmd/frontend/helpers/handlers"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/middleware"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/session"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/gobuffalo/packr/v2"
)

var (
	filesBox = packr.New("files", "./assets")
	distBox  = packr.New("dist", "./assets/dist")
	imgBox   = packr.New("img", "./assets/img")
)

func main() {

	rand.Seed(time.Now().UnixNano())

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
	consumers.Init(consumers.FrontendDefinitions)
	session.Init()
	handlers.Init()
	email.Init()

	// Clear caches on process restart
	if config.IsProd() {
		keys := []string{
			memcache.ItemCommitsPage(1).Key,
		}
		err = memcache.Client().Delete(keys...)
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
	r.Use(chiMiddleware.Compress(flate.DefaultCompression))
	r.Use(rateLimitMiddleware)

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
		r.Get("/img/*", handlers2.RootFileHandler(imgBox, "/assets/img"))
		r.Get("/dist/*", handlers2.RootFileHandler(distBox, "/assets/dist"))
	})

	// Root files
	r.Get("/ads.txt", handlers2.RootFileHandler(filesBox, ""))
	r.Get("/browserconfig.xml", handlers2.RootFileHandler(filesBox, ""))
	r.Get("/robots.txt", handlers2.RootFileHandler(filesBox, ""))
	r.Get("/site.webmanifest", handlers2.RootFileHandler(filesBox, ""))

	// Shortcuts
	r.Get("/a{id:[0-9]+}", handlers2.RedirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/games/" + chi.URLParam(r, "id") }))
	r.Get("/g{id:[0-9]+}", handlers2.RedirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/games/" + chi.URLParam(r, "id") }))
	r.Get("/s{id:[0-9]+}", handlers2.RedirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/packages/" + chi.URLParam(r, "id") }))
	r.Get("/p{id:[0-9]+}", handlers2.RedirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/players/" + chi.URLParam(r, "id") }))
	r.Get("/b{id:[0-9]+}", handlers2.RedirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/bundles/" + chi.URLParam(r, "id") }))

	// Redirects
	r.Get("/apps", handlers2.RedirectHandler("/games"))
	r.Get("/apps/{one}", handlers2.RedirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/games/" + chi.URLParam(r, "one") }))
	r.Get("/apps/{one}/{two}", handlers2.RedirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/games/" + chi.URLParam(r, "one") + "/" + chi.URLParam(r, "two") }))
	r.Get("/subs", handlers2.RedirectHandler("/packages"))
	r.Get("/subs/{one}", handlers2.RedirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/packages/" + chi.URLParam(r, "one") }))
	r.Get("/subs/{one}/{two}", handlers2.RedirectHandlerFunc(func(w http.ResponseWriter, r *http.Request) string { return "/packages/" + chi.URLParam(r, "one") + "/" + chi.URLParam(r, "two") }))
	r.Get("/coop", handlers2.RedirectHandler("/games/coop"))
	r.Get("/chat", handlers2.RedirectHandler("/discord-server"))
	r.Get("/chat-bot", handlers2.RedirectHandler("/discord-bot"))
	r.Get("/chat/{id}", handlers2.RedirectHandler("/discord-server"))
	r.Get("/sitemap/index.xml", handlers2.RedirectHandler("/sitemap.xml"))
	r.Get("/steam-api", handlers2.RedirectHandler("/api/steam"))
	r.Get("/api", handlers2.RedirectHandler("/api/gamedb"))
	r.Get("/discord", handlers2.RedirectHandler("/discord-bot")) // Used in discord messages

	// Game Redirects
	r.Get("/new-releases", handlers2.RedirectHandler("/games/new-releases"))
	r.Get("/random", handlers2.RedirectHandler("/games/random"))
	r.Get("/sales", handlers2.RedirectHandler("/games/sales"))
	r.Get("/trending", handlers2.RedirectHandler("/games/trending"))
	r.Get("/upcoming", handlers2.RedirectHandler("/games/upcoming"))
	r.Get("/wishlists", handlers2.RedirectHandler("/games/wishlists"))

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
		memcache.Close,
	)
}

var limiters = rate.New(time.Second, rate.WithBurst(10))

func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		err := limiters.GetLimiter(r.RemoteAddr).Wait(r.Context())
		if err != nil {
			log.ErrS(err)
			handlers.Error500Handler(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
