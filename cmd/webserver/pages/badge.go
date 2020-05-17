package pages

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
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

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid badge ID"})
		return
	}

	app := mongo.App{}
	badge, ok := mongo.GlobalBadges[id]
	if !ok {

		app, err = mongo.GetApp(id)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Invalid badge ID"})
			} else {
				log.Err(err)
				returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "Something went wrong"})
			}
			return
		}

		// Get what we can from the app
		badge = mongo.PlayerBadge{
			AppID:     app.ID,
			AppName:   app.GetName(),
			BadgeID:   1,
			BadgeName: app.GetName(),
		}

		// Look for full badge row, may not exist if nobody has the badge
		appBadge, err := mongo.GetAppBadge(id)
		if err != nil {
			err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
			log.Err(err)
		} else {
			badge = appBadge
		}
	}

	badge.BadgeFoil = r.URL.Query().Get("foil") == "1"

	var playerLevel int
	var playerTime string
	var playerRank string

	if session.IsLoggedIn(r) {

		badge.PlayerID, err = session.GetPlayerIDFromSesion(r)
		if err != nil {
			log.Err(err, r)
			returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: err.Error()})
			return
		}

		var row = mongo.PlayerBadge{}
		err = mongo.FindOne(mongo.CollectionPlayerBadges, bson.D{{"_id", badge.GetKey()}}, nil, nil, &row)
		if err != nil && err != mongo.ErrNoDocuments {
			log.Err(err, r)
			returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: err.Error()})
			return
		}

		if err == nil {

			playerLevel = row.BadgeLevel
			playerTime = row.BadgeCompletionTime.Format(helpers.DateYearTime)

			var filter = bson.D{
				{Key: "badge_level", Value: bson.M{"$gte": row.BadgeLevel}},
				{Key: "badge_completion_time", Value: bson.M{"$lt": row.BadgeCompletionTime}},
			}

			if badge.IsSpecial() {
				filter = append(filter,
					bson.E{Key: "app_id", Value: 0},
					bson.E{Key: "badge_id", Value: id},
				)
			} else {
				filter = append(filter,
					bson.E{Key: "app_id", Value: id},
					bson.E{Key: "badge_id", Value: bson.M{"$gt": 0}},
					bson.E{Key: "badge_foil", Value: badge.BadgeFoil},
				)
			}

			count, err := mongo.CountDocuments(mongo.CollectionPlayerBadges, filter, 60*60*24*14)
			if err != nil {
				log.Err(err, r)
				returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: err.Error()})
				return
			}

			playerRank = helpers.OrdinalComma(int(count + 1))
		}
	}

	t := badgeTemplate{}
	t.setBackground(app, false, false)
	t.fill(w, r, badge.GetName()+" Badge", "Steam Badge Ladder / Leaderboard")
	t.IncludeSocialJS = true

	t.LoggedIn = session.IsLoggedIn(r)
	t.Badge = badge
	t.PlayerLevel = playerLevel
	t.PlayerTime = playerTime
	t.PlayerRank = playerRank

	returnTemplate(w, r, "badge", t)
}

type badgeTemplate struct {
	GlobalTemplate
	Badge       mongo.PlayerBadge
	PlayerLevel int
	PlayerRank  string
	PlayerTime  string
	LoggedIn    bool
}

func badgeAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid badge ID"})
		return
	}

	badge, ok := mongo.GlobalBadges[id]
	if !ok {
		badge = mongo.PlayerBadge{AppID: id, BadgeID: 1}
	}

	badge.BadgeFoil = r.URL.Query().Get("foil") == "1"

	var query = datatable.NewDataTableQuery(r, true)

	//
	var filter bson.D
	if badge.IsSpecial() {
		filter = bson.D{
			bson.E{Key: "app_id", Value: 0},
			bson.E{Key: "badge_id", Value: id},
		}
	} else {
		filter = bson.D{
			bson.E{Key: "app_id", Value: id},
			bson.E{Key: "badge_id", Value: bson.M{"$gt": 0}},
			bson.E{Key: "badge_foil", Value: badge.BadgeFoil},
		}
	}

	var wg sync.WaitGroup

	var badges []mongo.PlayerBadge
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		badges, err = mongo.GetBadgePlayers(query.GetOffset64(), filter)
		if err != nil {
			log.Err(err, r)
		}
	}()

	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountDocuments(mongo.CollectionPlayerBadges, filter, 0)
		if err != nil {
			log.Err(err, r)
		}
	}()

	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, count, count)
	for k, player := range badges {
		response.AddRow([]interface{}{
			query.GetOffset() + k + 1,                                 // 0
			helpers.GetPlayerName(player.PlayerID, player.PlayerName), // 1
			player.GetPlayerIcon(),                                    // 2
			player.BadgeLevel,                                         // 3
			player.BadgeCompletionTime.Format(helpers.DateSQL),        // 4
			player.GetPlayerPath(),                                    // 5
			player.GetPlayerCommunityLink(),                           // 6
		})
	}

	returnJSON(w, r, response)
}
