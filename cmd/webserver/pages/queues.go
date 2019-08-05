package pages

import (
	"net/http"
	"strings"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
)

func QueuesRouter() http.Handler {
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

	var item = helpers.MemcacheQueues
	var highcharts = map[string]helpers.HighChartsJson{}

	err := helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &highcharts, func() (interface{}, error) {

		fields := []string{
			// `"queue"='GameDB_CS_Apps'`,
			// `"queue"='GameDB_CS_Packages'`,
			// `"queue"='GameDB_CS_Profiles'`,
			`"queue"='GameDB_Go_Apps'`,
			`"queue"='GameDB_Go_Changes'`,
			`"queue"='GameDB_Go_Groups'`,
			`"queue"='GameDB_Go_Packages'`,
			`"queue"='GameDB_Go_Profiles'`,
		}

		builder := influxql.NewBuilder()
		builder.AddSelect(`sum("messages")`, "sum_messages")
		builder.SetFrom(helpers.InfluxTelegrafDB, helpers.InfluxRetentionPolicy14Day.String(), helpers.InfluxMeasurementRabbitQueue.String())
		builder.AddWhere("time", ">=", "now() - 1h")
		builder.AddWhereRaw("(" + strings.Join(fields, " OR ") + ")")
		builder.AddGroupByTime("10s")
		builder.AddGroupBy("queue")
		builder.SetFillNone()

		resp, err := helpers.InfluxQuery(builder.String())
		if err != nil {
			log.Err(builder.String(), r)
			return highcharts, err
		}

		ret := map[string]helpers.HighChartsJson{}
		if len(resp.Results) > 0 {
			for _, v := range resp.Results[0].Series {
				ret[strings.Replace(v.Tags["queue"], "GameDB_Go_", "", 1)] = helpers.InfluxResponseToHighCharts(v)
			}
		}

		return ret, err
	})

	if err != nil {
		log.Err(err, r)
		return
	}

	err = returnJSON(w, r, highcharts)
	log.Err(err, r)
}
