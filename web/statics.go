package web

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/steam-authority/steam-authority/logger"
)

func HeaderHandler(w http.ResponseWriter, r *http.Request) {

	// Load templates needed
	folder := os.Getenv("STEAM_PATH")
	t, err := template.New("t").Funcs(getTemplateFuncMap()).ParseFiles(folder + "/templates/esi_header.html")
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 404, err.Error())
		return
	}

	// Template
	tp := GlobalTemplate{}
	tp.Fill(r, "Header")

	// Write a respone
	err = t.ExecuteTemplate(w, "esi_header", tp)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, "Something has gone wrong, the error has been logged!")
		return
	}
}

func InfoHandler(w http.ResponseWriter, r *http.Request) {

	t := staticTemplate{}
	t.Fill(r, "Info")

	returnTemplate(w, r, "info", t)
}

func DonateHandler(w http.ResponseWriter, r *http.Request) {

	t := staticTemplate{}
	t.Fill(r, "Donate")

	returnTemplate(w, r, "donate", t)
}

func Error404Handler(w http.ResponseWriter, r *http.Request) {

	returnErrorTemplate(w, r, 404, "page not found")
}

func RootFileHandler(w http.ResponseWriter, r *http.Request) {

	data, err := ioutil.ReadFile(os.Getenv("STEAM_PATH") + r.URL.Path)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 404, "Unable to read file.")
		return
	}

	w.Write(data)

}

type staticTemplate struct {
	GlobalTemplate
}
