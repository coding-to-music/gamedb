package tasks

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/sql/pics"
	"go.mongodb.org/mongo-driver/bson"
)

type ScanProductKeys struct {
	BaseTask
}

func (c ScanProductKeys) ID() string {
	return "scan-product-keys"
}

func (c ScanProductKeys) Name() string {
	return "Scan Product Keys"
}

func (c ScanProductKeys) Cron() string {
	return CronTimeScanProductQueues
}

func (c ScanProductKeys) work() (err error) {

	var addedKeys []string

	// APPS
	var projection = bson.M{"common": 1, "extended": 1, "config": 1, "ufs": 1}
	var filter = bson.D{
		{"$or", bson.A{
			bson.M{"extended": bson.M{"$exists": true}},
			bson.M{"common": bson.M{"$exists": true}},
			bson.M{"ufs": bson.M{"$exists": true}},
			bson.M{"config": bson.M{"$exists": true}},
		}},
	}

	var productKeysMap = map[string]map[string]map[string]*sql.ProductKey{}

	apps, err := mongo.GetApps(0, 0, nil, filter, projection, nil)
	if err != nil {
		return err
	}

	for _, app := range apps {

		fields := map[string]pics.PICSKeyValues{
			sql.ProductKeyFieldExtended: app.Extended,
			sql.ProductKeyFieldCommon:   app.Common,
			sql.ProductKeyFieldUFS:      app.UFS,
			sql.ProductKeyFieldConfig:   app.Config,
		}

		for field, data := range fields {

			for key := range data {

				key = helpers.TruncateString(key, 256, "...")

				if _, ok := productKeysMap[sql.ProductKeyTypeApp]; !ok {
					productKeysMap[sql.ProductKeyTypeApp] = map[string]map[string]*sql.ProductKey{}
				}
				if _, ok := productKeysMap[sql.ProductKeyTypeApp][field]; !ok {
					productKeysMap[sql.ProductKeyTypeApp][field] = map[string]*sql.ProductKey{}
				}

				if _, ok := productKeysMap[sql.ProductKeyTypeApp][field][key]; ok {

					productKeysMap[sql.ProductKeyTypeApp][field][key].Count++

				} else {

					productKeysMap[sql.ProductKeyTypeApp][field][key] = &sql.ProductKey{
						Type:  sql.ProductKeyTypeApp,
						Field: field,
						Key:   key,
						Count: 1,
					}
				}
			}
		}
	}

	apps = nil // Memory

	for k, v := range productKeysMap {
		for kk, vv := range v {
			log.Info(k, kk)
			for kkk, vvv := range vv {
				err := vvv.Save()
				if err != nil {
					log.Err(err, k, kk, kkk)
				} else {
					addedKeys = append(addedKeys, k+"-"+kk+"-"+kk+"-")
				}
			}
		}
	}

	// PACKAGES
	projection = bson.M{"extended": 1}
	filter = bson.D{
		{"$or", bson.A{
			bson.M{"extended": bson.M{"$exists": true}},
		}},
	}

	productKeysMap = map[string]map[string]map[string]*sql.ProductKey{}

	packages, err := mongo.GetPackages(0, 0, nil, filter, projection, nil)
	if err != nil {
		return err
	}

	for _, pack := range packages {

		fields := map[string]pics.PICSKeyValues{
			sql.ProductKeyFieldExtended: pack.Extended,
		}

		for field, data := range fields {

			for key := range data {

				key = helpers.TruncateString(key, 256, "...")

				if _, ok := productKeysMap[sql.ProductKeyTypePackage]; !ok {
					productKeysMap[sql.ProductKeyTypePackage] = map[string]map[string]*sql.ProductKey{}
				}
				if _, ok := productKeysMap[sql.ProductKeyTypePackage][field]; !ok {
					productKeysMap[sql.ProductKeyTypePackage][field] = map[string]*sql.ProductKey{}
				}

				if _, ok := productKeysMap[sql.ProductKeyTypePackage][field][key]; ok {

					productKeysMap[sql.ProductKeyTypePackage][field][key].Count++

				} else {

					productKeysMap[sql.ProductKeyTypePackage][field][key] = &sql.ProductKey{
						Type:  sql.ProductKeyTypePackage,
						Field: field,
						Key:   key,
						Count: 1,
					}
				}
			}
		}
	}

	packages = nil // Memory

	for k, v := range productKeysMap {
		for kk, vv := range v {
			log.Info(k, kk)
			for kkk, vvv := range vv {
				err := vvv.Save()
				if err != nil {
					log.Err(err, k, kk, kkk)
				} else {
					addedKeys = append(addedKeys, k+"-"+kk+"-"+kk+"-")
				}
			}
		}
	}

	// Mark removed keys
	all, err := sql.GetProductKeys()
	if err != nil {
		return err
	}

	var deleted int
	for _, v := range all {

		key := v.Type + "-" + v.Field + "-" + v.Key

		var found bool
		for _, vv := range addedKeys {
			if vv == key {
				found = true
				break
			}
		}

		if !found {
			v.Count = 0
			err = v.Save()
			if err != nil {
				log.Err(err)
			} else {
				deleted++
			}
		}
	}

	log.Info("Removing " + strconv.Itoa(deleted) + " keys")

	return nil
}
