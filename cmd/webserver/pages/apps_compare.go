package pages

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

func gamesCompareRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", appsCompareHandler)
	r.Get("/search.json", compareSearchAjaxHandler)
	r.Get("/apps.json", compareAppsAjaxHandler)
	r.Get("/{id}", appsCompareHandler)
	r.Get("/{id}/players.json", appsComparePlayersAjaxHandler)
	r.Get("/{id}/players2.json", appsComparePlayers2AjaxHandler)
	r.Get("/{id}/members.json", appsCompareGroupsHandler)
	r.Get("/{id}/reviews.json", appsCompareScoresHandler)
	r.Get("/{id}/wishlists.json", appsCompareWishlistHandler)
	return r
}

func appsCompareHandler(w http.ResponseWriter, r *http.Request) {

	var idStrings = helpers.UniqueString(helpers.RegexInts.FindAllString(chi.URLParam(r, "id"), -1))

	var apps []mongo.App
	var names []string
	var namesMap = map[string]string{}
	var ids []string

	var groupIDs []string
	var groupNamesMap = map[string]string{}

	for _, appID := range idStrings {

		id, err := strconv.Atoi(appID)
		if err == nil && helpers.IsValidAppID(id) {

			app, err := mongo.GetApp(id)
			if err != nil {
				err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
				if err != nil {
					log.Err(err, r)
				}
				continue
			}

			apps = append(apps, app)
			names = append(names, app.GetName())
			namesMap[appID] = app.GetName()
			ids = append(ids, appID)
			groupNamesMap[app.GroupID] = app.GetName()

			if app.GroupID != "" {
				groupIDs = append(groupIDs, app.GroupID)
			}
		}
	}

	if len(apps) > 10 {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Too many apps"})
		return
	}

	// Template
	t := appsCompareTemplate{}
	t.fill(w, r, "Compare Games", template.HTML(strings.Join(names, " vs ")))
	t.addAssetHighCharts()
	t.Apps = apps
	t.IDs = strings.Join(ids, ",")
	t.GroupIDs = strings.Join(groupIDs, ",")

	b, err := json.Marshal(namesMap)
	if err != nil {
		log.Err(err, r)
	}
	t.AppNames = template.JS(b)

	b, err = json.Marshal(groupNamesMap)
	if err != nil {
		log.Err(err, r)
	}
	t.GroupNames = template.JS(b)

	// Make google JSON
	var j = appsCompareGoogleTemplate{}
	var d int64
	for _, v := range apps {
		if v.ReleaseDateUnix < d || d == 0 {
			d = v.ReleaseDateUnix
		}
	}
	if d == 0 {
		d = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	}
	appsChunks := mongo.ChunkApps(apps, 5)
	for _, appChunk := range appsChunks {

		j.ComparisonItem = []appsCompareGoogleItemTemplate{}

		for _, v := range appChunk {
			j.ComparisonItem = append(j.ComparisonItem, appsCompareGoogleItemTemplate{
				Keyword: v.GetName(),
				Time:    time.Unix(d, 0).AddDate(-1, 0, 0).Format(helpers.DateSQLDay) + " " + time.Now().Format(helpers.DateSQLDay),
			})
		}

		b, err = json.Marshal(j)
		if err != nil {
			log.Err(err, r)
		}

		t.GoogleJSON = append(t.GoogleJSON, template.JS(b))
	}

	returnTemplate(w, r, "apps_compare", t)
}

type appsCompareTemplate struct {
	globalTemplate
	Apps       []mongo.App
	IDs        string
	GroupIDs   string
	AppNames   template.JS
	GroupNames template.JS
	GoogleJSON []template.JS
}

type appsCompareGoogleTemplate struct {
	ComparisonItem []appsCompareGoogleItemTemplate `json:"comparisonItem"`
	Category       int                             `json:"category"`
	Property       string                          `json:"property"`
}

type appsCompareGoogleItemTemplate struct {
	Keyword string `json:"keyword"`
	Geo     string `json:"geo"`
	Time    string `json:"time"`
}

func compareSearchAjaxHandler(w http.ResponseWriter, r *http.Request) {

	var limit = 5

	query := datatable.NewDataTableQuery(r, true)
	search := query.GetSearchString("search")
	code := session.GetProductCC(r)
	ids := helpers.StringToSlice(query.GetSearchString("ids"), ",")
	response := datatable.NewDataTablesResponse(r, query, int64(limit), int64(limit), nil)

	if search != "" {

		apps, err := elasticsearch.SearchAppsSimple(limit, search)
		if err != nil {
			log.Err(err, r)
			return
		}

		for k, app := range apps {

			var offset = query.GetOffset() + k + 1
			var price = app.Prices.Get(code).GetFinal()
			var linkBool = helpers.SliceHasString(strconv.Itoa(app.ID), ids)
			var link = makeCompareActionLink(ids, strconv.Itoa(app.ID), linkBool)

			response.AddRow([]interface{}{
				offset,                 // 0
				app.ID,                 // 1
				app.GetName(),          // 2
				app.GetIcon(),          // 3
				app.GetPath(),          // 4
				app.GetCommunityLink(), // 5
				price,                  // 6,
				link,                   // 7
				linkBool,               // 8
				app.Score,              // 9 - Search Score
			})
		}
	}

	returnJSON(w, r, response)
}

func compareAppsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, true)
	code := session.GetProductCC(r)
	ids := helpers.StringToSlice(query.GetSearchString("ids"), ",")
	ids2 := helpers.StringSliceToIntSlice(ids)

	apps, err := mongo.GetAppsByID(ids2, bson.M{"_id": 1, "name": 1, "icon": 1, "prices": 1})
	if err != nil {
		log.Err(err, r)
		return
	}

	var appMap = map[int][]interface{}{}

	var response = datatable.NewDataTablesResponse(r, query, int64(len(apps)), int64(len(apps)), nil)
	for _, app := range apps {

		var price = app.GetPrices().Get(code).GetFinal()
		var linkBool = helpers.SliceHasString(strconv.Itoa(app.ID), ids)
		var link = makeCompareActionLink(ids, strconv.Itoa(app.ID), linkBool)

		appMap[app.ID] = []interface{}{
			nil,                    // 0
			app.ID,                 // 1
			app.GetName(),          // 2
			app.GetIcon(),          // 3
			app.GetPath(),          // 4
			app.GetCommunityLink(), // 5
			price,                  // 6
			link,                   // 7
			linkBool,               // 8
			0,                      // 9 - Search Score
		}
	}

	for _, v := range ids2 {
		if val, ok := appMap[v]; ok {
			response.AddRow(val)
		}
	}

	returnJSON(w, r, response)
}

func makeCompareActionLink(ids []string, id string, linkBool bool) string {

	var newIDs []string

	if linkBool {
		for _, v := range ids {
			if v != id {
				newIDs = append(newIDs, v)
			}
		}
	} else {
		newIDs = ids
		newIDs = append(newIDs, id)
	}

	return "/games/compare/" + strings.Join(newIDs, ",")
}

func appsComparePlayersAjaxHandler(w http.ResponseWriter, r *http.Request) {

	var ids = helpers.UniqueString(helpers.RegexInts.FindAllString(chi.URLParam(r, "id"), -1))

	if len(ids) < 1 || len(ids) > 10 {
		return
	}

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW()-7d")
	builder.AddWhereRaw(`"app_id" =~ /^(` + strings.Join(ids, "|") + `)$/`)
	builder.AddGroupByTime("10m")
	builder.AddGroupBy("app_id")
	builder.SetFillNumber(0)

	resp, err := influx.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var ret []influx.HighChartsJSONMulti
	if len(resp.Results) > 0 {
		for _, id := range ids {
			for _, v := range resp.Results[0].Series {
				if id == v.Tags["app_id"] {
					ret = append(ret, influx.HighChartsJSONMulti{
						Key:   v.Tags["app_id"],
						Value: influx.InfluxResponseToHighCharts(v, false),
					})
				}
			}
		}
	}

	returnJSON(w, r, ret)
}

func appsComparePlayers2AjaxHandler(w http.ResponseWriter, r *http.Request) {

	var ids = helpers.UniqueString(helpers.RegexInts.FindAllString(chi.URLParam(r, "id"), -1))

	if len(ids) < 1 || len(ids) > 10 {
		return
	}

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW()-1825d")
	builder.AddWhereRaw(`"app_id" =~ /^(` + strings.Join(ids, "|") + `)$/`)
	builder.AddGroupByTime("1d")
	builder.AddGroupBy("app_id")
	builder.SetFillNumber(0)

	resp, err := influx.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var ret []influx.HighChartsJSONMulti
	if len(resp.Results) > 0 {
		for _, id := range ids {
			for _, v := range resp.Results[0].Series {
				if id == v.Tags["app_id"] {
					ret = append(ret, influx.HighChartsJSONMulti{
						Key:   v.Tags["app_id"],
						Value: influx.InfluxResponseToHighCharts(v, true),
					})
				}
			}
		}
	}

	returnJSON(w, r, ret)
}

func appsCompareWishlistHandler(w http.ResponseWriter, r *http.Request) {

	var ids = helpers.UniqueString(helpers.RegexInts.FindAllString(chi.URLParam(r, "id"), -1))

	if len(ids) < 1 || len(ids) > 10 {
		return
	}

	builder := influxql.NewBuilder()
	builder.AddSelect("MEAN(wishlist_count)", "mean_wishlist_count")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW()-365d")
	builder.AddWhereRaw(`"app_id" =~ /^(` + strings.Join(ids, "|") + `)$/`)
	builder.AddGroupByTime("1d")
	builder.AddGroupBy("app_id")
	builder.SetFillNone()

	resp, err := influx.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var ret []influx.HighChartsJSONMulti
	if len(resp.Results) > 0 {
		for _, id := range ids {
			for _, v := range resp.Results[0].Series {
				if id == v.Tags["app_id"] {
					ret = append(ret, influx.HighChartsJSONMulti{
						Key:   v.Tags["app_id"],
						Value: influx.InfluxResponseToHighCharts(v, true),
					})
				}
			}
		}
	}

	returnJSON(w, r, ret)
}

func appsCompareScoresHandler(w http.ResponseWriter, r *http.Request) {

	var ids = helpers.UniqueString(helpers.RegexInts.FindAllString(chi.URLParam(r, "id"), -1))

	if len(ids) < 1 || len(ids) > 10 {
		return
	}

	builder := influxql.NewBuilder()
	builder.AddSelect("mean(reviews_score)", "mean_reviews_score")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW()-365d")
	builder.AddWhereRaw(`"app_id" =~ /^(` + strings.Join(ids, "|") + `)$/`)
	builder.AddGroupByTime("1d")
	builder.AddGroupBy("app_id")
	builder.SetFillNone()

	resp, err := influx.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var ret []influx.HighChartsJSONMulti
	if len(resp.Results) > 0 {
		for _, id := range ids {
			for _, v := range resp.Results[0].Series {
				if id == v.Tags["app_id"] {
					ret = append(ret, influx.HighChartsJSONMulti{
						Key:   v.Tags["app_id"],
						Value: influx.InfluxResponseToHighCharts(v, true),
					})
				}
			}
		}
	}

	returnJSON(w, r, ret)
}

func appsCompareGroupsHandler(w http.ResponseWriter, r *http.Request) {

	var ids []string
	var err error

	for _, v := range helpers.UniqueString(helpers.RegexInts.FindAllString(chi.URLParam(r, "id"), -1)) {

		v, err = helpers.IsValidGroupID(v) // Upgrade group IDs
		if err != nil {
			continue
		}

		ids = append(ids, v)
	}

	if len(ids) < 1 || len(ids) > 10 {
		return
	}

	builder := influxql.NewBuilder()
	builder.AddSelect(`max("members_count")`, "max_members_count")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementGroups.String())
	builder.AddWhereRaw(`"group_id" =~ /^(` + strings.Join(ids, "|") + `)$/`)
	builder.AddGroupByTime("1d")
	builder.AddGroupBy("group_id")
	builder.SetFillNone()

	resp, err := influx.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var ret []influx.HighChartsJSONMulti
	if len(resp.Results) > 0 {
		for _, id := range ids {
			for _, v := range resp.Results[0].Series {
				if id == v.Tags["group_id"] {
					ret = append(ret, influx.HighChartsJSONMulti{
						Key:   v.Tags["group_id"],
						Value: influx.InfluxResponseToHighCharts(v, true),
					})
				}
			}
		}
	}

	returnJSON(w, r, ret)
}
