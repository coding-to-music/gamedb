package pages

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gamedb/gamedb/cmd/webserver/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

func BadgeRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", badgeHandler)
	r.Get("/{slug}", badgeHandler)
	r.Get("/players.json", badgeAjaxHandler)
	return r
}

func badgeHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid badge ID"})
		return
	}

	idx, err := strconv.Atoi(id)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid badge ID"})
		return
	}

	val, ok := mongo.GlobalBadges[idx]
	if !ok {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid badge ID"})
		return
	}

	t := badgeTemplate{}
	t.fill(w, r, val.BadgeName, "")
	t.Badge = val
	t.Foil = r.URL.Query().Get("foil")
	t.IncludeSocialJS = true

	returnTemplate(w, r, "badge", t)
}

type badgeTemplate struct {
	GlobalTemplate
	Badge mongo.PlayerBadge
	Foil  string
}

func badgeAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid badge ID"})
		return
	}

	idx, err := strconv.Atoi(id)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid badge ID"})
		return
	}

	badge, ok := mongo.GlobalBadges[idx]
	if !ok {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid badge ID"})
		return
	}

	query := datatable.NewDataTableQuery(r, true)

	var wg sync.WaitGroup

	var filter = bson.D{}

	if badge.IsSpecial() {
		filter = append(filter, bson.E{Key: "app_id", Value: 0})
		filter = append(filter, bson.E{Key: "badge_id", Value: idx})
	} else {
		filter = append(filter, bson.E{Key: "app_id", Value: idx})
		filter = append(filter, bson.E{Key: "badge_id", Value: bson.M{"$gt": 0}})
		filter = append(filter, bson.E{Key: "badge_foil", Value: r.URL.Query().Get("foil") == "1"})
	}

	var badges []mongo.PlayerBadge
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		badges, err = mongo.GetBadgePlayers(query.GetOffset64(), filter)
		log.Err(err, r)
	}()

	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountDocuments(mongo.CollectionPlayerBadges, filter, 0)
		log.Err(err, r)
	}()

	wg.Wait()

	response := datatable.DataTablesResponse{}
	response.RecordsTotal = count
	response.RecordsFiltered = count
	response.Draw = query.Draw
	response.Limit(r)

	for k, player := range badges {
		response.AddRow([]interface{}{
			query.GetOffset() + k + 1, // 0
			player.PlayerName,         // 1
			player.GetPlayerIcon(),    // 2
			player.BadgeLevel,         // 3
			player.BadgeCompletionTime.Format("2006-01-02 15:04:05"), // 4
			player.GetPlayerPath(), // 5
		})
	}

	returnJSON(w, r, response.Output())
}
