package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"strings"

	"github.com/99designs/basicauth-go"
	"github.com/Jleagle/go-helpers/logger"
	"github.com/go-chi/chi"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/pics"
	"github.com/steam-authority/steam-authority/queue"
	"github.com/steam-authority/steam-authority/web"
	"github.com/steam-authority/steam-authority/websockets"
)

func main() {

	logger.SetRollbarKey(os.Getenv("STEAM_ROLLBAR_PRIVATE"))

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", os.Getenv("STEAM_GOOGLE_APPLICATION_CREDENTIALS"))
	if os.Getenv("ENV") == "local" {
		os.Setenv("STEAM_DOMAIN", os.Getenv("STEAM_DOMAIN_LOCAL"))
	} else {
		os.Setenv("STEAM_DOMAIN", "https://steamauthority.net")
	}

	// Flags
	flagDebug := flag.Bool("debug", false, "Debug")
	flagPics := flag.Bool("pics", false, "Pics")
	flagConsumers := flag.Bool("consumers", false, "Consumers")
	flagPprof := flag.Bool("pprof", false, "PProf")

	flag.Parse()

	if *flagPprof {
		go http.ListenAndServe(":8080", nil)
	}

	if *flagDebug {
		mysql.SetDebug(true)
	}

	if *flagPics {
		go pics.Run()
	}

	if *flagConsumers {
		queue.RunConsumers()
	}

	// Routes
	r := chi.NewRouter()

	r.Mount("/admin", adminRouter())

	r.Get("/apps", web.AppsHandler)
	r.Get("/apps/{id}", web.AppHandler)
	r.Get("/apps/{id}/{slug}", web.AppHandler)

	r.Get("/changes", web.ChangesHandler)
	r.Get("/changes/{id}", web.ChangeHandler)

	r.Get("/chat", web.ChatHandler)
	r.Get("/chat/{id}", web.ChatHandler)

	r.Get("/contact", web.ContactHandler)
	r.Post("/contact", web.PostContactHandler)

	r.Get("/deals", web.DealsHandler)
	r.Get("/deals/{id}", web.DealsHandler)

	r.Get("/experience", web.ExperienceHandler)
	r.Get("/experience/{id}", web.ExperienceHandler)

	r.Get("/login", web.LoginHandler)
	r.Get("/logout", web.LogoutHandler)
	r.Get("/login-callback", web.LoginCallbackHandler)

	r.Get("/packages", web.PackagesHandler)
	r.Get("/packages/{id}", web.PackageHandler)

	r.Post("/players", web.PlayerIDHandler)
	r.Get("/players", web.RanksHandler)
	r.Get("/players/{id:[a-z]+}", web.RanksHandler)
	r.Get("/players/{id:[0-9]+}", web.PlayerHandler)
	r.Get("/players/{id:[0-9]+}/{slug}", web.PlayerHandler)

	r.Get("/queues", web.QueuesHandler)
	r.Get("/queues/queues.json", web.QueuesJSONHandler)

	r.Get("/settings", web.SettingsHandler)
	r.Post("/settings", web.SaveSettingsHandler)

	r.Get("/browserconfig.xml", web.RootFileHandler)
	r.Get("/site.webmanifest", web.RootFileHandler)

	// Other
	r.Get("/", web.HomeHandler)
	r.Get("/commits", web.CommitsHandler)
	r.Get("/donate", web.DonateHandler)
	r.Get("/genres", web.GenresHandler)
	r.Get("/info", web.InfoHandler)
	r.Get("/news", web.NewsHandler)
	r.Get("/tags", web.TagsHandler)
	r.Get("/websocket", websockets.Handler)
	r.Get("/coop", web.CoopHandler)

	// 404
	r.NotFound(web.Error404Handler)

	// File server
	fileServer(r, "/assets")

	http.ListenAndServe(":8085", r)

	// Block for goroutines to run forever
	forever := make(chan bool)
	<-forever
}

func adminRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(basicauth.New("Steam", map[string][]string{
		os.Getenv("STEAM_ADMIN_USER"): {os.Getenv("STEAM_ADMIN_PASS")},
	}))
	r.Get("/", web.AdminHandler)
	r.Get("/{option}", web.AdminHandler)
	r.Post("/{option}", web.AdminHandler)
	return r
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func fileServer(r chi.Router, path string) {

	if strings.ContainsAny(path, "{}*") {
		logger.Info("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(http.Dir(filepath.Join(os.Getenv("STEAM_PATH"), "assets"))))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
