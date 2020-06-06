package pages

import (
	"html/template"
	"net/http"
	"strings"
	"sync"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
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
			err = queue.ProduceGroup(queue.GroupMessage{ID: id, UserAgent: &ua})
			log.Err(err)

			returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Sorry but we can not find this group"})
			return
		}

		log.Err(r, err)
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
				log.Err(err, r)
			}
		} else {
			t.setBackground(app, true, true)
		}
	}

	t.fill(w, r, group.GetName(), template.HTML(group.Headline))
	t.addAssetHighCharts()
	t.Canonical = group.GetPath()
	t.IncludeSocialJS = true

	// Update group
	func() {

		if !group.ShouldUpdate() {
			return
		}

		ua := r.UserAgent()
		err = queue.ProduceGroup(queue.GroupMessage{ID: group.ID, UserAgent: &ua})
		if err == nil {
			log.Info(log.LogNameTriggerUpdate, r, ua)
			t.addToast(Toast{Title: "Update", Message: "Group has been queued for an update"})
		}
		err = helpers.IgnoreErrors(err, queue.ErrIsBot, memcache.ErrInQueue)
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Fix links
	summary := group.Summary
	summary = strings.ReplaceAll(summary, "https://steamcommunity.com/linkfilter/?url=", "")

	//
	t.Group = group
	t.Summary = helpers.RenderHTMLAndBBCode(summary)
	t.Group.Error = strings.Replace(t.Group.Error, "Click here for information on how to report groups on Steam.", "", 1)

	returnTemplate(w, r, "group", t)
}

type groupTemplate struct {
	GlobalTemplate
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

	//
	query := datatable.NewDataTableQuery(r, true)

	//
	var wg sync.WaitGroup

	// Get players
	var players []mongo.Player
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		playerGroups, err := mongo.GetGroupPlayers(id, query.GetOffset64())
		if err != nil {
			log.Err(err, r)
			return
		}

		// Get players
		var playerIDs []int64
		for _, v := range playerGroups {
			playerIDs = append(playerIDs, v.PlayerID)
		}

		playersResp, err := mongo.GetPlayersByID(playerIDs, bson.M{"_id": 1, "persona_name": 1, "vanity_url": 1, "avatar": 1, "level": 1, "country_code": 1})
		if err != nil {
			log.Err(err, r)
			return
		}

		var playersMap = map[int64]mongo.Player{}
		for _, v := range playersResp {
			playersMap[v.ID] = v
		}

		// Build response in correct order
		for _, v := range playerGroups {
			if val, ok := playersMap[v.PlayerID]; ok {
				players = append(players, val)
			}
		}
	}(r)

	// Get total
	var total int64
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionPlayerGroups, bson.D{{"group_id", id}}, 60*60*6)
		log.Err(err, r)
	}(r)

	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, total, total, nil)
	for _, player := range players {
		response.AddRow([]interface{}{
			player.ID,              // 0
			player.GetName(),       // 1
			player.CommunityLink(), // 2
			player.GetAvatar(),     // 3
			player.GetFlag(),       // 4
			player.Level,           // 5
			player.CountryCode,     // 6
			player.GetAvatar2(),    // 7
			player.GetPath(),       // 8
		})
	}

	returnJSON(w, r, response)
}

func groupAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := helpers.IsValidGroupID(chi.URLParam(r, "id"))
	if err != nil {
		return
	}

	builder := influxql.NewBuilder()
	builder.AddSelect(`max("members_count")`, "max_members_count")
	builder.AddSelect(`max("members_in_chat")`, "max_members_in_chat")
	builder.AddSelect(`max("members_in_game")`, "max_members_in_game")
	// builder.AddSelect(`max("members_online")`, "max_members_online")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementGroups.String())
	builder.AddWhere("group_id", "=", id)
	// builder.AddWhere("time", ">", "now()-365d")
	builder.AddGroupByTime("1d")
	builder.SetFillNone()

	resp, err := influx.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var hc influx.HighChartsJSON

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

		hc = influx.InfluxResponseToHighCharts(resp.Results[0].Series[0])
	}

	returnJSON(w, r, hc)
}
