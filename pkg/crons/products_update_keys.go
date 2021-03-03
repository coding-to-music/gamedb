package crons

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/mysql/pics"
	"go.mongodb.org/mongo-driver/bson"
)

type ProductsUpdateKeys struct {
	BaseTask
}

func (c ProductsUpdateKeys) ID() string {
	return "scan-product-keys"
}

func (c ProductsUpdateKeys) Name() string {
	return "Scan Product Keys"
}

func (c ProductsUpdateKeys) Group() TaskGroup {
	return ""
}

func (c ProductsUpdateKeys) Cron() TaskTime {
	return CronTimeScanProductQueues
}

func (c ProductsUpdateKeys) work() (err error) {

	var addedKeys []string
	var limit int64 = 10_000

	// APPS
	var offset int64 = 0
	var projection = bson.M{"common": 1, "extended": 1, "config": 1, "ufs": 1}
	var filter = bson.D{
		{"$or", bson.A{
			bson.M{"extended": bson.M{"$exists": true}},
			bson.M{"common": bson.M{"$exists": true}},
			bson.M{"ufs": bson.M{"$exists": true}},
			bson.M{"config": bson.M{"$exists": true}},
		}},
	}

	var productKeysMap = map[string]map[string]map[string]*mysql.ProductKey{}

	for {

		apps, err := mongo.GetApps(offset, limit, nil, filter, projection)
		if err != nil {
			return err
		}

		for _, app := range apps {

			fields := map[string]pics.PICSKeyValues{
				mysql.ProductKeyFieldExtended: app.Extended,
				mysql.ProductKeyFieldCommon:   app.Common,
				mysql.ProductKeyFieldUFS:      app.UFS,
				mysql.ProductKeyFieldConfig:   app.Config,
			}

			for field, data := range fields {

				for key := range data {

					key = helpers.TruncateString(key, 256, "...")

					if _, ok := productKeysMap[mysql.ProductKeyTypeApp]; !ok {
						productKeysMap[mysql.ProductKeyTypeApp] = map[string]map[string]*mysql.ProductKey{}
					}
					if _, ok := productKeysMap[mysql.ProductKeyTypeApp][field]; !ok {
						productKeysMap[mysql.ProductKeyTypeApp][field] = map[string]*mysql.ProductKey{}
					}

					if _, ok := productKeysMap[mysql.ProductKeyTypeApp][field][key]; ok {

						productKeysMap[mysql.ProductKeyTypeApp][field][key].Count++

					} else {

						productKeysMap[mysql.ProductKeyTypeApp][field][key] = &mysql.ProductKey{
							Type:  mysql.ProductKeyTypeApp,
							Field: field,
							Key:   key,
							Count: 1,
						}
					}
				}
			}
		}

		if int64(len(apps)) != limit {
			break
		}

		offset += limit
	}

	for k, v := range productKeysMap {
		for kk, vv := range v {
			// log.InfoS(k, kk)
			for kkk, vvv := range vv {
				err := vvv.Save()
				if err != nil {
					log.ErrS(err, k, kk, kkk)
				} else {
					addedKeys = append(addedKeys, k+"-"+kk+"-"+kkk)
				}
			}
		}
	}

	// PACKAGES
	offset = 0
	projection = bson.M{"extended": 1}
	filter = bson.D{
		{"$or", bson.A{
			bson.M{"extended": bson.M{"$exists": true}},
		}},
	}

	productKeysMap = map[string]map[string]map[string]*mysql.ProductKey{}

	for {

		packages, err := mongo.GetPackages(offset, limit, nil, filter, projection)
		if err != nil {
			return err
		}

		for _, pack := range packages {

			fields := map[string]pics.PICSKeyValues{
				mysql.ProductKeyFieldExtended: pack.Extended,
			}

			for field, data := range fields {

				for key := range data {

					key = helpers.TruncateString(key, 256, "...")

					if _, ok := productKeysMap[mysql.ProductKeyTypePackage]; !ok {
						productKeysMap[mysql.ProductKeyTypePackage] = map[string]map[string]*mysql.ProductKey{}
					}
					if _, ok := productKeysMap[mysql.ProductKeyTypePackage][field]; !ok {
						productKeysMap[mysql.ProductKeyTypePackage][field] = map[string]*mysql.ProductKey{}
					}

					if _, ok := productKeysMap[mysql.ProductKeyTypePackage][field][key]; ok {

						productKeysMap[mysql.ProductKeyTypePackage][field][key].Count++

					} else {

						productKeysMap[mysql.ProductKeyTypePackage][field][key] = &mysql.ProductKey{
							Type:  mysql.ProductKeyTypePackage,
							Field: field,
							Key:   key,
							Count: 1,
						}
					}
				}
			}
		}

		if int64(len(packages)) != limit {
			break
		}

		offset += limit
	}

	for k, v := range productKeysMap {
		for kk, vv := range v {
			// log.InfoS(k, kk)
			for kkk, vvv := range vv {
				err := vvv.Save()
				if err != nil {
					log.ErrS(err, k, kk, kkk)
				} else {
					addedKeys = append(addedKeys, k+"-"+kk+"-"+kkk)
				}
			}
		}
	}

	// Mark removed keys
	all, err := mysql.GetProductKeys()
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
				log.ErrS(err)
			} else {
				deleted++
			}
		}
	}

	log.InfoS("Removing " + strconv.Itoa(deleted) + " keys")

	return nil
}
