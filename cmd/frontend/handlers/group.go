package handlers

import (
	"html/template"
	"net/http"
	"strings"
	"sync"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

func GroupRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", groupHandler)
	r.Get("/members.json", groupAjaxHandler)
	r.Get("/table.json", groupTableAjaxHandler)
	r.Get("/{slug}", groupHandler)
	return r
}

func groupHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid group ID"})
		return
	}

	id, err := helpers.IsValidGroupID(id)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid group ID: " + id})
		return
	}

	// Get group
	group, err := mongo.GetGroup(id)
	if err != nil {

		if err == mongo.ErrNoDocuments {

			ua := r.UserAgent()
			err = consumers.ProduceGroup(consumers.GroupMessage{ID: id, UserAgent: &ua})
			err = helpers.IgnoreErrors(err, consumers.ErrInQueue, consumers.ErrIsBot)
			if err != nil {
				log.ErrS(err)
			}

			returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Sorry but we can not find this group"})
			return
		}

		log.ErrS(err)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the group"})
		return
	}

	t := groupTemplate{}

	// Get background app
	if group.Type == helpers.GroupTypeGame && group.AppID > 0 {

		var err error
		app, err := mongo.GetApp(group.AppID)
		if err != nil {
			err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
			if err != nil {
				log.ErrS(err)
			}
		} else {
			t.setBackground(app, true, true)
		}
	}

	t.fill(w, r, "group", group.GetName(), template.HTML(group.Headline))
	t.addAssetHighCharts()
	t.Canonical = group.GetPath()

	// Update group
	func() {

		if !group.ShouldUpdate() {
			return
		}

		ua := r.UserAgent()
		err = consumers.ProduceGroup(consumers.GroupMessage{ID: group.ID, UserAgent: &ua})
		if err == nil {
			log.Info("group queued", zap.String("ua", ua))
			t.addToast(Toast{Title: "Update", Message: "Group has been queued for an update", Success: true})
		}
		err = helpers.IgnoreErrors(err, consumers.ErrIsBot, consumers.ErrInQueue)
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Fix links
	summary := group.Summary
	summary = strings.ReplaceAll(summary, "https://steamcommunity.com/linkfilter/?url=", "")

	//
	t.Group = group
	t.Summary = helpers.RenderHTMLAndBBCode(summary)
	t.Group.Error = strings.Replace(t.Group.Error, "Click here for information on how to report groups on Steam.", "", 1)

	returnTemplate(w, r, t)
}

type groupTemplate struct {
	globalTemplate
	Group   mongo.Group
	Summary template.HTML
}

func groupTableAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		return
	}

	id, err := helpers.IsValidGroupID(id)
	if err != nil {
		return
	}

	// Get group
	group, err := mongo.GetGroup(id)
	if err != nil {
		err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
		if err != nil {
			log.ErrS(err)
		}
		return
	}

	//
	query := datatable.NewDataTableQuery(r, true)

	//
	var wg sync.WaitGroup

	// Get players
	var playerGroups []mongo.PlayerGroup
	wg.Add(1)
	go func() {

		defer wg.Done()

		var columns = map[string]string{
			"2": "player_level",
			"3": "player_games",
		}

		var err error
		playerGroups, err = mongo.GetGroupPlayers(id, query.GetOffset64(), query.GetOrderMongo(columns))
		if err != nil {
			log.ErrS(err)
			return
		}
	}()

	// Get total
	var total int64
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionPlayerGroups, bson.D{{Key: "group_id", Value: id}}, 60*60*6)
		if err != nil {
			log.ErrS(err)
		}
	}(r)

	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, int64(group.Members), total, nil)
	for _, playerGroup := range playerGroups {
		response.AddRow([]interface{}{
			playerGroup.PlayerID,                 // 0
			playerGroup.GetPlayerName(),          // 1
			playerGroup.GetPlayerCommunityLink(), // 2
			playerGroup.GetPlayerAvatar(),        // 3
			playerGroup.GetPlayerFlag(),          // 4
			playerGroup.PlayerLevel,              // 5
			playerGroup.PlayerCountry,            // 6
			playerGroup.GetPlayerAvatar2(),       // 7
			playerGroup.GetPlayerPath(),          // 8
			playerGroup.PlayerGames,              // 9
		})
	}

	returnJSON(w, r, response)
}

func groupAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := helpers.IsValidGroupID(chi.URLParam(r, "id"))
	if err != nil {
		return
	}

	var hc influx.HighChartsJSON

	callback := func() (interface{}, error) {

		builder := influxql.NewBuilder()
		builder.AddSelect(`MAX("members_count")`, "max_members_count")
		builder.AddSelect(`MAX("members_in_chat")`, "max_members_in_chat")
		builder.AddSelect(`MAX("members_in_game")`, "max_members_in_game")
		// builder.AddSelect(`max("members_online")`, "max_members_online")
		builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementGroups.String())
		builder.AddWhere("group_id", "=", id)
		// builder.AddWhere("time", ">", "now()-365d")
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

	item := memcache.ItemGroupFollowersChart(id)
	err = memcache.Client().GetSet(item.Key, item.Expiration, &hc, callback)
	if err != nil {
		log.ErrS(err)
	}

	returnJSON(w, r, hc)
}
