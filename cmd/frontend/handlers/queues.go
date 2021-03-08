package handlers

import (
	"net/http"
	"strings"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/go-chi/chi/v5"
)

var queuePageCharts = []string{
	string(consumers.QueuePlayers),
	string(consumers.QueueGroups),
	string(consumers.QueueApps),
	string(consumers.QueuePackages),
	string(consumers.QueueBundles),
	string(consumers.QueueChanges),
	string(consumers.QueueDelay),
}

func QueuesRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", queuesHandler)
	r.Get("/queues.json", queuesAjaxHandler)

	return r
}

func queuesHandler(w http.ResponseWriter, r *http.Request) {

	t := queuesTemplate{}
	t.fill(w, r, "queues", "Queues", "When new items get added to the site, they go through a queue to not overload the servers.")
	t.addAssetHighCharts()
	t.Charts = queuePageCharts

	returnTemplate(w, r, t)
}

type queuesTemplate struct {
	globalTemplate
	Charts []string
}

func queuesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	var highcharts = map[string]influx.HighChartsJSON{}

	callback := func() (interface{}, error) {

		builder := influxql.NewBuilder()
		builder.AddSelect(`max("messages")`, "max_messages")
		builder.SetFrom(influx.InfluxTelegrafDB, influx.InfluxRetentionPolicy14Day.String(), influx.InfluxMeasurementRabbitQueue.String())
		builder.AddWhere("time", ">=", "now() - 1h")
		builder.AddWhereRaw(`"queue" =~ /^(` + strings.Join(queuePageCharts, "|") + `)/`) // just get the main prefixes
		builder.AddGroupByTime("1m")
		builder.AddGroupBy("queue")
		builder.SetFillNone()

		resp, err := influx.InfluxQuery(builder)
		if err != nil {
			log.ErrS(builder.String())
			return highcharts, err
		}

		ret := map[string]influx.HighChartsJSON{}
		if len(resp.Results) > 0 {
			for _, v := range resp.Results[0].Series {
				ret[strings.Replace(v.Tags["queue"], "GameDB_Go_", "", 1)] = influx.InfluxResponseToHighCharts(v, false)
			}
		}

		return ret, err
	}

	item := memcache.ItemQueues
	err := memcache.Client().GetSet(item.Key, item.Expiration, &highcharts, callback)
	if err != nil {
		log.ErrS(err)
		return
	}

	returnJSON(w, r, highcharts)
}
