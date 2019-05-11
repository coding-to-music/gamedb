package main

import (
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/recaptcha-go"
	"github.com/gamedb/gamedb/cmd/webserver/pages"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

//noinspection GoUnusedGlobalVariable
var version string

func main() {

	log.Info("Starting PubSub")
	go websockets.ListenToPubSub()

	log.Info("Starting webserver")

	config.Config.CommitHash.SetDefault(version)

	recaptcha.SetSecret(config.Config.RecaptchaPrivate.Get())

	rand.Seed(time.Now().UnixNano())

	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		log.Critical("GOOGLE_APPLICATION_CREDENTIALS not found")
		os.Exit(1)
	}

	r := chi.NewRouter()

	r.Use(middlewareTime)
	r.Use(middlewareCors())
	r.Use(middleware.RealIP)
	// r.Use(middleware.DefaultCompress) // http: superfluous response.WriteHeader call from github.com/go-chi/chi/middleware.(*compressResponseWriter).Write (compress.go:228)
	r.Use(middleware.RedirectSlashes)
	r.Use(middlewareLog)

	// Pages
	r.Get("/", pages.HomeHandler)
	r.Mount("/admin", pages.AdminRouter())
	r.Mount("/api", pages.APIRouter())
	r.Mount("/apps", pages.AppsRouter())
	r.Mount("/bundles", pages.BundlesRouter())
	r.Mount("/changes", pages.ChangesRouter())
	r.Mount("/chat", pages.ChatRouter())
	r.Mount("/chat-bot", pages.ChatBotRouter())
	r.Mount("/commits", pages.CommitsRouter())
	r.Mount("/contact", pages.ContactRouter())
	r.Mount("/coop", pages.CoopRouter())
	r.Mount("/depots", pages.DepotsRouter())
	r.Mount("/developers", pages.DevelopersRouter())
	r.Mount("/donate", pages.DonateRouter())
	r.Mount("/experience", pages.ExperienceRouter())
	r.Mount("/franchise", pages.FranchiseRouter())
	r.Mount("/genres", pages.GenresRouter())
	r.Mount("/health-check", pages.HealthCheckRouter())
	r.Mount("/home", pages.HomeRouter())
	r.Mount("/info", pages.InfoRouter())
	r.Mount("/login", pages.LoginRouter())
	r.Mount("/logout", pages.LogoutRouter())
	r.Mount("/new-releases", pages.NewReleasesRouter())
	r.Mount("/news", pages.NewsRouter())
	r.Mount("/packages", pages.PackagesRouter())
	r.Mount("/patreon", pages.PatreonRouter())
	r.Mount("/players", pages.PlayersRouter())
	r.Mount("/price-changes", pages.PriceChangeRouter())
	r.Mount("/product-keys", pages.ProductKeysRouter())
	r.Mount("/publishers", pages.PublishersRouter())
	r.Mount("/queues", pages.QueuesRouter())
	r.Mount("/settings", pages.SettingsRouter())
	r.Mount("/signup", pages.SignupRouter())
	r.Mount("/sitemap", pages.SiteMapRouter())
	r.Mount("/stats", pages.StatsRouter())
	r.Mount("/steam-api", pages.SteamAPIRouter())
	r.Mount("/tags", pages.TagsRouter())
	r.Mount("/trending", pages.TrendingRouter())
	r.Mount("/twitter", pages.TwitterRouter())
	r.Mount("/upcoming", pages.UpcomingRouter())
	r.Mount("/websocket", pages.WebsocketsRouter())

	// Profiling
	if config.IsLocal() {
		r.Mount("/debug", middleware.Profiler())
	}

	// Files
	r.Get("/browserconfig.xml", pages.RootFileHandler)
	r.Get("/robots.txt", pages.RootFileHandler)
	r.Get("/site.webmanifest", pages.RootFileHandler)

	// File server
	fileServer(r, "/assets", http.Dir(config.Config.AssetsPath.Get()))

	// 404
	r.NotFound(pages.Error404Handler)

	err := http.ListenAndServe(config.ListenOn(), r)
	log.Critical(err)

	helpers.KeepAlive()
}

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.IsLocal() {
			// log.Info(log.LogNameRequests, r.Method+" "+r.URL.Path)
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

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func fileServer(r chi.Router, path string, root http.FileSystem) {

	if strings.ContainsAny(path, "{}*") {
		log.Info("Invalid URL " + path)
		return
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusFound).ServeHTTP)
		path += "/"
	}
	path += "*"

	if strings.Contains(path, "..") {
		log.Info("Invalid URL " + path)
		return
	}

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
