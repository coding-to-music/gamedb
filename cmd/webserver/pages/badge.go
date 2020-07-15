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

	var app = mongo.App{}
	var playerBadge mongo.PlayerBadge

	// Get a player badge from an ID
	builtInBadge, ok := helpers.BuiltInSpecialBadges[id]
	if ok {
		playerBadge.AppID = builtInBadge.AppID
		playerBadge.BadgeID = builtInBadge.BadgeID
		playerBadge.BadgeIcon = builtInBadge.Icon
		playerBadge.AppName = builtInBadge.Name
	} else {

		builtInBadge, ok = helpers.BuiltInEventBadges[id]
		if ok {
			playerBadge.AppID = builtInBadge.AppID
			playerBadge.BadgeID = builtInBadge.BadgeID
			playerBadge.BadgeIcon = builtInBadge.Icon
			playerBadge.AppName = builtInBadge.Name
		} else {

			app, err = mongo.GetApp(id)
			if err != nil {
				if err == mongo.ErrNoDocuments || err == mongo.ErrInvalidAppID {
					returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Invalid badge ID"})
				} else {
					log.Err(err, r)
					returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "Something went wrong"})
				}
				return
			}

			// Get what we can from the app
			builtInBadge = helpers.BuiltInbadge{
				AppID:   app.ID,
				BadgeID: 1,
				Name:    app.GetName(),
				Icon:    app.GetIcon(),
			}

			// Look for full badge row, may not exist if nobody has the badge
			appBadge, err := mongo.GetAppBadge(id)
			if err != nil {
				err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
				log.Err(err, r)
			} else {
				playerBadge = appBadge
			}
		}
	}

	playerBadge.BadgeFoil = r.URL.Query().Get("foil") == "1"

	var playerLevel int
	var playerTime string
	var playerRank string

	if session.IsLoggedIn(r) {

		playerBadge.PlayerID, err = session.GetPlayerIDFromSesion(r)
		if err == nil && playerBadge.PlayerID > 0 {

			var row = mongo.PlayerBadge{}
			err = mongo.FindOne(mongo.CollectionPlayerBadges, bson.D{{"_id", playerBadge.GetKey()}}, nil, nil, &row)
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

				if playerBadge.IsSpecial() {
					filter = append(filter,
						bson.E{Key: "app_id", Value: 0},
						bson.E{Key: "badge_id", Value: id},
					)
				} else {
					filter = append(filter,
						bson.E{Key: "app_id", Value: id},
						bson.E{Key: "badge_id", Value: bson.M{"$gt": 0}},
						bson.E{Key: "badge_foil", Value: playerBadge.BadgeFoil},
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
	}

	t := badgeTemplate{}
	t.setBackground(app, false, false)
	t.fill(w, r, playerBadge.GetName()+" Badge", "Steam Badge Ladder / Leaderboard")
	t.IncludeSocialJS = true

	t.LoggedIn = session.IsLoggedIn(r)
	t.Badge = playerBadge
	t.PlayerLevel = playerLevel
	t.PlayerTime = playerTime
	t.PlayerRank = playerRank

	returnTemplate(w, r, "badge", t)
}

type badgeTemplate struct {
	globalTemplate
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

	var query = datatable.NewDataTableQuery(r, true)

	badge, ok := helpers.BuiltInSpecialBadges[id]
	if !ok {
		badge, ok = helpers.BuiltInEventBadges[id]
		if !ok {
			badge = helpers.BuiltInbadge{
				AppID:   id,
				BadgeID: 1,
			}
		}
	}

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
			bson.E{Key: "badge_foil", Value: r.URL.Query().Get("foil") == "1"},
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

	var response = datatable.NewDataTablesResponse(r, query, count, count, nil)
	for k, playerBadge := range badges {

		response.AddRow([]interface{}{
			query.GetOffset() + k + 1,                               // 0
			playerBadge.GetPlayerName(),                             // 1
			playerBadge.GetPlayerIcon(),                             // 2
			playerBadge.BadgeLevel,                                  // 3
			playerBadge.BadgeCompletionTime.Format(helpers.DateSQL), // 4
			playerBadge.GetPlayerPath(),                             // 5
			playerBadge.GetPlayerCommunityLink(),                    // 6
		})
	}

	returnJSON(w, r, response)
}
