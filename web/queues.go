package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/go-chi/chi"
)

func queuesRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", queuesHandler)
	r.Get("/queues.json", queuesAjaxHandler)
	return r
}

func queuesHandler(w http.ResponseWriter, r *http.Request) {

	t := queuesTemplate{}
	t.fill(w, r, "Queues", "When new items get added to the site, they go through a queue to not overload the servers.")
	t.addAssetHighCharts()

	err := returnTemplate(w, r, "queues", t)
	log.Err(err, r)
}

type queuesTemplate struct {
	GlobalTemplate
}

func queuesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	var item = helpers.MemcacheQueues
	var highcharts = map[string]db.HighChartsJson{}

	err := helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &highcharts, func() (interface{}, error) {

		builder := influxql.NewBuilder()
		builder.AddSelect(`sum("messages")`, "sum_messages")
		builder.SetFrom("Telegraf", "14d", "rabbitmq_queue")
		builder.AddWhere("time", ">=", "now() - 1h")
		builder.AddWhereRaw(`("queue"='GameDB_Go_Apps' OR "queue"='GameDB_Go_Packages' OR "queue"='GameDB_Go_Profiles')`)
		builder.AddGroupByTime("10s")
		builder.AddGroupBy("queue")
		builder.SetFillNone()

		resp, err := db.InfluxQuery(builder.String())
		if err != nil {
			log.Err(builder.String())
			return highcharts, err
		}

		ret := map[string]db.HighChartsJson{}
		if len(resp.Results) > 0 {
			for _, v := range resp.Results[0].Series {
				ret[strings.Replace(v.Tags["queue"], "GameDB_Go_", "", 1)] = db.InfluxResponseToHighCharts(v)
			}
		}

		return ret, err
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
