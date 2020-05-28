package pages

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
)

func appsCompareHandler(w http.ResponseWriter, r *http.Request) {

	var idStrings = strings.Split(chi.URLParam(r, "id"), ",")
	idStrings = helpers.UniqueString(idStrings)

	var apps []appsCompareAppTemplate
	var names []string
	var namesMap = map[string]string{}
	var ids []string

	var groupIDs []string
	var groupNamesMap = map[string]string{}

	var wg sync.WaitGroup

	for _, appID := range idStrings {

		wg.Add(1)
		go func(appID string) {

			defer wg.Done()

			id, err := strconv.Atoi(appID)
			if err == nil && helpers.IsValidAppID(id) {

				a, err := mongo.GetApp(id)
				if err != nil {
					err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
					log.Err(err)
					return
				}

				app := appsCompareAppTemplate{App: a}

				var wg2 sync.WaitGroup

				// Tags
				wg2.Add(1)
				go func() {

					defer wg2.Done()

					var err error
					app.Tags, err = GetAppTags(app.App)
					if err != nil {
						log.Err(err, r)
					}
				}()

				// Categories
				wg2.Add(1)
				go func() {

					defer wg2.Done()

					var err error
					app.Categories, err = GetAppCategories(app.App)
					if err != nil {
						log.Err(err, r)
					}
				}()

				// Genres
				wg2.Add(1)
				go func() {

					defer wg2.Done()

					var err error
					app.Genres, err = GetAppGenres(app.App)
					if err != nil {
						log.Err(err, r)
					}
				}()

				// Get Developers
				wg2.Add(1)
				go func() {

					defer wg2.Done()

					var err error
					app.Developers, err = GetDevelopers(app.App)
					if err != nil {
						log.Err(err, r)
					}
				}()

				// Get Publishers
				wg2.Add(1)
				go func() {

					defer wg2.Done()

					var err error
					app.Publishers, err = GetPublishers(app.App)
					if err != nil {
						log.Err(err, r)
					}
				}()

				// Wait
				wg2.Wait()

				apps = append(apps, app)
				names = append(names, app.App.GetName())
				namesMap[appID] = app.App.GetName()
				ids = append(ids, appID)

				groupIDs = append(groupIDs, a.GroupID)
				groupNamesMap[a.GroupID] = app.App.GetName()
			}
		}(appID)
	}

	wg.Wait()

	if len(apps) < 2 {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Not enough valid app IDs"})
		return
	}

	if len(apps) > 5 {
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
		if v.App.ReleaseDateUnix < d || d == 0 {
			d = v.App.ReleaseDateUnix
		}
	}
	if d == 0 {
		d = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	}
	for _, v := range apps {
		j.ComparisonItem = append(j.ComparisonItem, appsCompareGoogleItemTemplate{
			Keyword: v.App.GetName(),
			Time:    time.Unix(d, 0).AddDate(-1, 0, 0).Format(helpers.DateSQLDay) + " " + time.Now().Format(helpers.DateSQLDay),
		})
	}

	b, err = json.Marshal(j)
	if err != nil {
		log.Err(err, r)
	}

	t.GoogleJSON = template.JS(b)

	returnTemplate(w, r, "apps_compare", t)
}

type appsCompareTemplate struct {
	GlobalTemplate
	Apps       []appsCompareAppTemplate
	IDs        string
	GroupIDs   string
	AppNames   template.JS
	GroupNames template.JS
	GoogleJSON template.JS
}

type appsCompareAppTemplate struct {
	App        mongo.App
	Tags       []sql.Tag
	Categories []sql.Category
	Developers []sql.Developer
	Genres     []sql.Genre
	Publishers []sql.Publisher
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

func appsComparePlayersAjaxHandler(w http.ResponseWriter, r *http.Request) {

	ids := strings.Split(chi.URLParam(r, "id"), ",")

	if len(ids) < 1 || len(ids) > 5 {
		return
	}

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW()-7d")
	builder.AddWhereRaw(`"app_id" =~ /^(` + strings.Join(ids, "|") + `)$/`)
	builder.AddGroupByTime("10m")
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
						Value: influx.InfluxResponseToHighCharts(v, false),
					})
				}
			}
		}
	}

	returnJSON(w, r, ret)
}

func appsComparePlayer2sAjaxHandler(w http.ResponseWriter, r *http.Request) {

	ids := strings.Split(chi.URLParam(r, "id"), ",")

	if len(ids) < 1 || len(ids) > 5 {
		return
	}

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW()-1825d")
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
						Value: influx.InfluxResponseToHighCharts(v, false),
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

	for _, v := range strings.Split(chi.URLParam(r, "id"), ",") {

		v, err = helpers.IsValidGroupID(v)
		if err != nil {
			continue
		}

		ids = append(ids, v)
	}

	if len(ids) < 1 || len(ids) > 5 {
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
						Value: influx.InfluxResponseToHighCharts(v, false),
					})
				}
			}
		}
	}

	returnJSON(w, r, ret)
}
