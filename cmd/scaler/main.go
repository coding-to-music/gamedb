package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/cmd/scaler/hosts"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
)

var version string
var commits string

func main() {

	err := config.Init(version, commits, helpers.GetIP())
	log.InitZap(log.LogNameScaler)
	defer log.Flush()
	if err != nil {
		log.FatalS(err)
		return
	}

	// Web server
	r := chi.NewRouter()
	r.Get("/", listHandler)
	r.Get("/create", createHandler)
	r.Get("/cycle", cycleHandler)
	r.Get("/delete/{id}", deleteHandler)
	r.Get("/health-check", healthCheckHandler)

	fmt.Println("Starting scaler on :4000")

	err = http.ListenAndServe(":4000", r)
	if err != nil {
		log.FatalS(err)
	}
}

func listHandler(w http.ResponseWriter, _ *http.Request) {

	funcs := template.FuncMap{
		"join":  func(a []string) string { return strings.Join(a, ", ") },
		"comma": func(a int) string { return humanize.Comma(int64(a)) },
	}

	t, err := template.New("t").Funcs(funcs).ParseFiles("list.gohtml")
	if err != nil {
		fmt.Println(err)
		return
	}

	host := hosts.GetHost()

	// Get template data
	data := HomeTemplate{}
	data.Consumers, err = host.ListConsumers()
	if err != nil {
		fmt.Println(err)
	}

	//
	err = t.ExecuteTemplate(w, "list", data)
	if err != nil {
		fmt.Println(err)
	}
}

type HomeTemplate struct {
	Consumers []hosts.Consumer
}

func createHandler(w http.ResponseWriter, r *http.Request) {

	_, err := hosts.GetHost().CreateConsumer()
	if err != nil {
		log.ErrS(err)
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		return
	}

	idx, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println(err)
		return
	}

	host := hosts.GetHost()
	err = host.DeleteConsumer(idx)
	if err != nil {
		fmt.Println(err)
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func cycleHandler(w http.ResponseWriter, r *http.Request) {

	host := hosts.GetHost()

	consumers, err := host.ListConsumers()
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, v := range consumers {

		if helpers.SliceHasString(hosts.ConsumerTag, v.Tags) {

			err = host.DeleteConsumer(v.ID)
			if err != nil {
				fmt.Println(err)
			}

			createHandler(w, r)
		}
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func healthCheckHandler(w http.ResponseWriter, _ *http.Request) {

	_, err := w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.ErrS(err)
	}
}
