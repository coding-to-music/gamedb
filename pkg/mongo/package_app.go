package mongo

import (
	"strconv"
)

type PackageApp struct {
	PackageID int `bson:"package_id"`
	AppID     int `bson:"app_id"`
}

func (app PackageApp) BSON() (ret interface{}) {

	return M{
		"_id":        app.getKey(),
		"package_id": app.PackageID,
		"app_id":     app.AppID,
	}
}

func (app PackageApp) getKey() string {
	return strconv.Itoa(app.PackageID) + "-" + strconv.Itoa(app.AppID)
}
