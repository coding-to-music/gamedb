package web

import (
	"io/ioutil"
	"net/http"

	"github.com/gamedb/website/logging"
	"github.com/spf13/viper"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func HeaderHandler(w http.ResponseWriter, r *http.Request) {

	t := GlobalTemplate{}
	t.Fill(w, r, "Header")

	returnTemplate(w, r, "_header_esi", t)
}

func InfoHandler(w http.ResponseWriter, r *http.Request) {

	t := GlobalTemplate{}
	t.Fill(w, r, "Info")

	returnTemplate(w, r, "info", t)
}

func DonateHandler(w http.ResponseWriter, r *http.Request) {

	t := GlobalTemplate{}
	t.Fill(w, r, "Donate")

	returnTemplate(w, r, "donate", t)
}

func Error404Handler(w http.ResponseWriter, r *http.Request) {

	returnErrorTemplate(w, r, 404, "page not found")
}

func RootFileHandler(w http.ResponseWriter, r *http.Request) {

	data, err := ioutil.ReadFile(viper.GetString("PATH") + r.URL.Path)
	if err != nil {
		logging.Error(err)
		returnErrorTemplate(w, r, 404, "Unable to read file.")
		return
	}

	w.Write(data)
}
