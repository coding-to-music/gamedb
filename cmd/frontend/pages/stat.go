package pages

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/gamedb/cmd/frontend/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func StatRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", statHandler)
	r.Get("/time.json", statAppsAjaxHandler)
	r.Get("/{slug}", statHandler)

	return r
}

func statHandler(w http.ResponseWriter, r *http.Request) {

	typex := statPathToConst(chi.URLParam(r, "type"))

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid App ID"})
		return
	}

	stat, err := mongo.GetStat(typex, id)
	if err == mongo.ErrNoDocuments {
		returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Unable to find " + typex.Title()})
		return
	} else if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: err.Error()})
		return
	}

	t := statTagsTemplate{}
	t.fill(w, r, stat.Name+" "+stat.Type.Title(), template.HTML(stat.Name+" "+stat.Type.Title()))
	t.addAssetHighCharts()

	t.Stat = stat

	returnTemplate(w, r, "stat", t)
}

type statTagsTemplate struct {
	globalTemplate
	Stat mongo.Stat
}

func statAppsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	typex := chi.URLParam(r, "type")
	id := chi.URLParam(r, "id")

	idx, err := strconv.Atoi(id)
	if err != nil {
		return
	}

	var hc influx.HighChartsJSON

	callback := func() (interface{}, error) {

		code := session.GetProductCC(r)

		stat := mongo.Stat{
			Type: statPathToConst(typex),
			ID:   idx,
		}

		builder := influxql.NewBuilder()
		builder.AddSelect(`max("apps_count")`, "max_apps_count")
		builder.AddSelect(`max("apps_percent")`, "max_apps_percent")
		builder.AddSelect(`max("mean_score")`, "max_mean_score")
		builder.AddSelect(`max("mean_players")`, "max_mean_players")
		builder.AddSelect(`max("mean_price_`+string(code)+`")`, "max_mean_price_"+string(code))
		builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementStats.String())
		builder.AddWhere("key", "=", stat.GetKey())
		builder.AddWhere("time", ">", "now()-365d")
		builder.AddGroupByTime("1d")
		builder.SetFillNone()

		resp, err := influx.InfluxQuery(builder)
		if err != nil {
			log.Err(err.Error(), zap.String("query", builder.String()))
			return hc, err
		}

		if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

			hc = influx.InfluxResponseToHighCharts(resp.Results[0].Series[0], true)
		}

		return hc, err
	}

	var item = memcache.MemcacheGroupFollowersChart(id)
	err = memcache.GetSetInterface(item.Key, item.Expiration, &hc, callback)
	if err != nil {
		log.ErrS(err)
	}

	returnJSON(w, r, hc)
}
