package main

import (
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/99designs/basicauth-go"
	"github.com/Jleagle/go-helpers/logger"
	"github.com/go-chi/chi"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/pics"
	"github.com/steam-authority/steam-authority/queue"
	"github.com/steam-authority/steam-authority/websockets"
)

func main() {

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", os.Getenv("STEAM_GOOGLE_APPLICATION_CREDENTIALS"))

	// Flags
	flagDebug := flag.Bool("debug", false, "Debug")
	flagPics := flag.Bool("pics", false, "Pics")
	flagConsumers := flag.Bool("consumers", false, "Consumers")

	flag.Parse()

	if *flagDebug {
		mysql.SetDebug(true)
	}

	if *flagPics {
		go pics.Run()
	}

	if *flagConsumers {
		queue.RunConsumers()
	}

	// Scripts
	arguments := os.Args[1:]
	if len(arguments) > 0 {

		switch arguments[0] {
		case "update-tags":
			logger.Info("Tags")
			os.Exit(0)
		case "update-genres":
			logger.Info("Genres")
			os.Exit(0)
		case "update-ranks":
			logger.Info("Ranks")
			os.Exit(0)
		}
	}

	logger.SetRollbarKey(os.Getenv("STEAM_ROLLBAR_PRIVATE"))

	r := chi.NewRouter()

	// Admin
	r.Mount("/admin", adminRouter())

	// Apps
	r.Get("/apps", appsHandler)
	r.Get("/apps/{id}", appHandler)
	r.Get("/apps/{id}/{slug}", appHandler)

	// Changes
	r.Get("/changes", changesHandler)
	r.Get("/changes/{id}", changeHandler)

	// Chat
	r.Get("/chat", chatHandler)
	r.Get("/chat/{id}", chatHandler)

	// Contact
	r.Get("/contact", contactHandler)
	r.Post("/contact", postContactHandler)

	// Deals
	r.Get("/deals", dealsHandler)
	r.Get("/deals/{id}", dealsHandler)

	// Experience
	r.Get("/experience", experienceHandler)
	r.Get("/experience/{id}", experienceHandler)

	// Login
	r.Get("/login", loginHandler)
	r.Get("/logout", logoutHandler)
	r.Get("/login-callback", loginCallbackHandler)

	// Packages
	r.Get("/packages", packagesHandler)
	r.Get("/packages/{id}", packageHandler)

	// Players
	r.Post("/players", playerIDHandler)
	r.Get("/players", playersHandler)
	r.Get("/players/{id:[a-z]+}", playersHandler)
	r.Get("/players/{id:[0-9]+}", playerHandler)
	r.Get("/players/{id:[0-9]+}/{slug}", playerHandler)

	// Settings
	r.Get("/settings", settingsHandler)
	r.Post("/settings", saveSettingsHandler)

	// Other
	r.Get("/", homeHandler)
	r.Get("/changelog", changelogHandler)
	r.Get("/genres", genresHandler)
	r.Get("/news", newsHandler)
	r.Get("/tags", tagsHandler)
	r.Get("/websocket", websockets.Handler)

	// Static pages
	r.Get("/credits", creditsHandler)
	r.Get("/donate", donateHandler)
	r.Get("/faqs", faqsHandler)

	//
	workDir, _ := os.Getwd()
	filesDir := filepath.Join(workDir, "assets")
	fileServer(r, "/assets", http.Dir(filesDir))

	http.ListenAndServe(":8085", r)

	// Block for goroutines
	forever := make(chan bool)
	<-forever
}

func adminRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(basicauth.New("Steam", map[string][]string{
		os.Getenv("STEAM_AUTH_USER"): {os.Getenv("STEAM_AUTH_PASS")},
	}))
	r.Get("/rerank", adminReRankHandler)
	r.Get("/fill-apps", adminUpdateAllAppsHandler)
	return r
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func fileServer(r chi.Router, path string, root http.FileSystem) {

	if strings.ContainsAny(path, "{}*") {
		logger.Info("FileServer does not permit URL parameters.")
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
