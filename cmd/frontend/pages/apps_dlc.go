package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

func appsDLCRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", dlcHandler)
	return r
}

func dlcHandler(w http.ResponseWriter, r *http.Request) {

	playerID := session.GetPlayerIDFromSesion(r)
	projection := bson.M{"app_id": 1, "app_name": 1, "app_dlc_count": 1, "app_icon": 1}

	playerApps, err := mongo.GetPlayerAppsByPlayer(playerID, 0, 100, bson.D{{"app_name", 1}}, projection, nil)
	if err != nil {
		log.ErrS(err)
		return
	}

	// Get all DLC for page
	var appIDs []int
	for _, v := range playerApps {
		appIDs = append(appIDs, v.AppID)
	}

	dlcs, err := mongo.GetDLCForApps(appIDs, 0, 0, nil, nil)
	if err != nil {
		log.ErrS(err)
		return
	}

	// Get owned DLC
	var dlcAppIDs bson.A
	for _, v := range dlcs {
		dlcAppIDs = append(dlcAppIDs, v.AppID)
	}

	filter := bson.D{{"app_id", bson.M{"$in": dlcAppIDs}}}

	ownedDLC, err := mongo.GetPlayerAppsByPlayer(playerID, 0, 0, nil, projection, filter)
	if err != nil {
		log.ErrS(err)
		return
	}

	log.InfoS(len(dlcs))
	log.InfoS(len(ownedDLC))

	t := appsDLCTemplate{}
	t.fill(w, r, "dlc", "DLC", "")
	t.PlayerApps = playerApps

	returnTemplate(w, r, t)
}

type appsDLCTemplate struct {
	globalTemplate
	PlayerApps []mongo.PlayerApp
}
