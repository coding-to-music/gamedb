package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/cmd/scaler/hosts"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/go-chi/chi"
)

func main() {

	err := config.Init(helpers.GetIP())
	log.InitZap(log.LogNameScaler)
	defer log.Flush()
	if err != nil {
		log.ErrS(err)
		return
	}

	// Web server
	r := chi.NewRouter()
	r.Get("/", listHandler)
	r.Get("/create", createHandler)
	r.Get("/cycle", cycleHandler)
	r.Get("/delete/{id}", deleteHandler)
	r.Get("/health-check", healthCheckHandler)

	s := &http.Server{
		Addr:              "0.0.0.0:4000",
		Handler:           r,
		ReadTimeout:       2 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	log.Info("Starting Scaler on " + "http://" + s.Addr)

	go func() {
		err = s.ListenAndServe()
		if err != nil {
			log.ErrS(err)
		}
	}()

	helpers.KeepAlive(
		mysql.Close,
		mongo.Close,
	)
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
