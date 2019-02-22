package web

import (
	"encoding/json"
	"net/http"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/go-chi/chi"
)

func queuesRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", queuesHandler)
	r.Get("/ajax.json", queuesJSONHandler)
	return r
}

func queuesHandler(w http.ResponseWriter, r *http.Request) {

	t := queuesTemplate{}
	t.Fill(w, r, "Queues", "When new items get added to the site, they go through a queue to not overload the servers.")
	t.addAssetHighCharts()

	err := returnTemplate(w, r, "queues", t)
	log.Err(err, r)
}

type queuesTemplate struct {
	GlobalTemplate
}

func queuesJSONHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	var item = helpers.MemcacheQueues
	var highcharts db.HighChartsJson

	err := helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &highcharts, func() (interface{}, error) {

		resp, err := db.InfluxQuery(`SELECT sum("messages") as "messages" FROM "Telegraf"."autogen"."rabbitmq_queue" WHERE time >= now() - 1h GROUP BY time(10s) fill(linear)`)
		if err != nil {
			log.Err(err, r)
			return highcharts, err
		}

		return db.InfluxResponseToHighCharts(resp), err
	})

	if err != nil {
		log.Err(err, r)
		return
	}

	b, err := json.Marshal(highcharts)
	if err != nil {
		log.Err(err, r)
		return
	}

	err = returnJSON(w, r, b)
	if err != nil {
		log.Err(err, r)
		return
	}
}
