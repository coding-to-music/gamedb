package mongo

import (
	"strconv"
)

type PlayerWishlistApp struct {
	PlayerID int64 `bson:"player_id"`
	AppID    int   `bson:"app_id"`
}

func (app PlayerWishlistApp) BSON() (ret interface{}) {

	return M{
		"_id":       app.getKey(),
		"player_id": app.PlayerID,
		"app_id":    app.AppID,
	}
}

func (app PlayerWishlistApp) getKey() string {
	return strconv.FormatInt(app.PlayerID, 10) + "-" + strconv.Itoa(app.AppID)
}
