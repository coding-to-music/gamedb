package mongo

import (
	"strconv"
)

type PlayerGroup struct {
	PlayerID int64  `bson:"player_id"`
	GroupID  string `bson:"group_id"`
}

func (app PlayerGroup) BSON() (ret interface{}) {

	return M{
		"_id":       app.getKey(),
		"player_id": app.PlayerID,
		"group_id":  app.GroupID,
	}
}

func (app PlayerGroup) getKey() string {
	return strconv.FormatInt(app.PlayerID, 10) + "-" + app.GroupID
}
