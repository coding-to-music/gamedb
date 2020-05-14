package main

import (
	"compress/flate"
	"net/http"
	_ "net/http/pprof"
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
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
	chiMiddleware "github.com/go-chi/chi/middleware"
	"github.com/gobuffalo/packr/v2"
)

var (
	version string

	distBox  = packr.New("dist", "./assets/dist")
	filesBox = packr.New("files", "./assets/files")
	imgBox   = packr.New("img", "./assets/img")
)

func main() {

	//
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		log.Critical("GOOGLE_APPLICATION_CREDENTIALS not found")
		os.Exit(1)
	}

	config.Init(version, helpers.GetIP())
	log.Initialise([]log.LogName{log.LogNameWebserver})

	// Get API key
	err := sql.GetAPIKey("webserver")
	if err != nil {
		log.Critical(err)
		return
	}

	if false {
		go func() {
			mongo.CreateAppIndexes()
			mongo.CreatePackageIndexes()
			mongo.CreatePlayerIndexes()
			mongo.CreateGroupIndexes()
			mongo.CreateSaleIndexes()
			log.Info("Mongo indexes finished")
		}()
	}

	if false {
		go func() {
			// elastic.DeleteAndRebuildAppsIndex()
			// elastic.DeleteAndRebuildAchievementsIndex()
			// elastic.DeleteAndRebuildPlayersIndex()
			log.Info("Elastic indexes finished")
		}()
	}

	// Start queue producers to send to.
	// In a go routine so if Rabbit is not working, the webserver still starts
	go queue.Init(queue.WebserverDefinitions)

	// Setup Recaptcha
	recaptcha.SetSecret(config.Config.RecaptchaPrivate.Get())

	session.InitSession()

	// Routes
	r := chi.NewRouter()
	r.Use(middleware.MiddlewareDownMessage)
	r.Use(middleware.MiddlewareCors())
	r.Use(chiMiddleware.RedirectSlashes)
	r.Use(middleware.MiddlewareRealIP)
	r.Use(chiMiddleware.NewCompressor(flate.BestSpeed, "text/html", "text/css", "text/javascript", "application/json", "application/javascript").Handler)

	// Pages
	r.Get("/", pages.HomeHandler)
	r.Get("/currency/{id}", pages.CurrencyHandler)
	r.Mount("/achievements", pages.AchievementsRouter())
	r.Mount("/admin", pages.AdminRouter())
	r.Mount("/api", pages.APIRouter())
	r.Mount("/apps", pages.AppsRouter())
	r.Mount("/badges", pages.BadgesRouter())
	r.Mount("/bundles", pages.BundlesRouter())
	r.Mount("/categories", pages.CategoriesRouter())
	r.Mount("/changes", pages.ChangesRouter())
	r.Mount("/commits", pages.CommitsRouter())
	r.Mount("/contact", pages.ContactRouter())
	r.Mount("/coop", pages.CoopRouter())
	r.Mount("/depots", pages.DepotsRouter())
	r.Mount("/developers", pages.DevelopersRouter())
	r.Mount("/discord-bot", pages.ChatBotRouter())
	r.Mount("/discord-server", pages.ChatRouter())
	r.Mount("/donate", pages.DonateRouter())
	r.Mount("/experience", pages.ExperienceRouter())
	r.Mount("/forgot", pages.ForgotRouter())
	r.Mount("/franchise", pages.FranchiseRouter())
	r.Mount("/genres", pages.GenresRouter())
	r.Mount("/groups", pages.GroupsRouter())
	r.Mount("/health-check", pages.HealthCheckRouter())
	r.Mount("/home", pages.HomeRouter())
	r.Mount("/info", pages.InfoRouter())
	r.Mount("/login", pages.LoginRouter())
	r.Mount("/logout", pages.LogoutRouter())
	r.Mount("/new-releases", pages.NewReleasesRouter())
	r.Mount("/news", pages.NewsRouter())
	r.Mount("/packages", pages.PackagesRouter())
	r.Mount("/players", pages.PlayersRouter())
	r.Mount("/price-changes", pages.PriceChangeRouter())
	r.Mount("/product-keys", pages.ProductKeysRouter())
	r.Mount("/publishers", pages.PublishersRouter())
	r.Mount("/queues", pages.QueuesRouter())
	r.Mount("/sales", pages.SalesRouter())
	r.Mount("/settings", pages.SettingsRouter())
	r.Mount("/signup", pages.SignupRouter())
	r.Mount("/stats", pages.StatsRouter())
	r.Mount("/steam-api", pages.SteamAPIRouter())
	r.Mount("/tags", pages.TagsRouter())
	r.Mount("/upcoming", pages.UpcomingRouter())
	r.Mount("/webhooks", pages.WebhooksRouter())
	r.Mount("/websocket", pages.WebsocketsRouter())
	r.Mount("/wishlists", pages.WishlistsRouter())

	// Sitemaps, Google doesnt like having a sitemap in a sub directory
	r.Get("/sitemap.xml", pages.SiteMapIndexHandler)
	r.Get("/sitemap-pages.xml", pages.SiteMapPagesHandler)
	r.Get("/sitemap-games-by-score.xml", pages.SiteMapGamesByScoreHandler)
	r.Get("/sitemap-games-by-players.xml", pages.SiteMapGamesByPlayersHandler)
	r.Get("/sitemap-games-new.xml", pages.SiteMapGamesNewHandler)
	r.Get("/sitemap-games-upcoming.xml", pages.SiteMapGamesUpcomingHandler)
	r.Get("/sitemap-players-by-level.xml", pages.SiteMapPlayersByLevel)
	r.Get("/sitemap-players-by-games.xml", pages.SiteMapPlayersByGamesCount)
	r.Get("/sitemap-groups.xml", pages.SiteMapGroups)
	r.Get("/sitemap-badges.xml", pages.SiteMapBadges)

	// Shortcuts
	r.Get("/a{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/apps/"+chi.URLParam(r, "id"), http.StatusFound)
	})
	r.Get("/s{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/packages/"+chi.URLParam(r, "id"), http.StatusFound)
	})
	r.Get("/p{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/players/"+chi.URLParam(r, "id"), http.StatusFound)
	})
	r.Get("/b{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/bundles/"+chi.URLParam(r, "id"), http.StatusFound)
	})

	// Profiling
	r.Route("/debug", func(r chi.Router) {
		r.Use(middleware.MiddlewareAuthCheck())
		r.Use(middleware.MiddlewareAdminCheck(pages.Error404Handler))
		r.Mount("/", chiMiddleware.Profiler())
	})

	// Assets
	r.Route("/assets", func(r chi.Router) {
		r.Get("/img/*", rootFileHandler(imgBox, "/assets/img"))
		r.Get("/files/*", rootFileHandler(filesBox, "/assets/files"))
		r.Get("/dist/*", rootFileHandler(distBox, "/assets/dist"))
	})

	// Root files
	r.Get("/browserconfig.xml", rootFileHandler(filesBox, ""))
	r.Get("/robots.txt", rootFileHandler(filesBox, ""))
	r.Get("/site.webmanifest", rootFileHandler(filesBox, ""))
	// r.Get("/ads.txt", rootFileHandler)

	// Redirects
	r.Get("/sitemap/index.xml", redirectHandler("/sitemap.xml"))
	r.Get("/trending", redirectHandler("/apps/trending"))
	r.Get("/games", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/apps"+"?"+r.URL.RawQuery, http.StatusFound)
	})
	r.Get("/games/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/apps/"+chi.URLParam(r, "id"), http.StatusFound)
	})
	r.Get("/games/{id:[0-9]+}/{slug}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/apps/"+chi.URLParam(r, "id")+"/"+chi.URLParam(r, "id"), http.StatusFound)
	})
	r.Get("/chat-bot", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/discord-bot", http.StatusFound)
	})
	r.Get("/chat", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/discord-server", http.StatusFound)
	})
	r.Get("/chat/{id}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/discord-server", http.StatusFound)
	})

	// 404
	r.NotFound(pages.Error404Handler)

	log.Info("Starting Webserver on " + config.ListenOn())
	err = http.ListenAndServe(config.ListenOn(), r)
	log.Critical(err)
}

func redirectHandler(url string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, http.StatusFound)
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
