package web

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/steam-authority/steam-authority/logger"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {

	t := http.Response{
		Body: ioutil.NopCloser(bytes.NewBufferString("OK")),
	}

	buff := bytes.NewBuffer(nil)
	t.Write(buff)
}

func HeaderHandler(w http.ResponseWriter, r *http.Request) {

	t := GlobalTemplate{}
	t.Fill(w, r, "Header")

	returnTemplate(w, r, "_header_esi", t)
}

func InfoHandler(w http.ResponseWriter, r *http.Request) {

	t := staticTemplate{}
	t.Fill(w, r, "Info")

	returnTemplate(w, r, "info", t)
}

func DonateHandler(w http.ResponseWriter, r *http.Request) {

	t := staticTemplate{}
	t.Fill(w, r, "Donate")

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
