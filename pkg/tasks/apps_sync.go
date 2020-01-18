package tasks

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
)

type AppsSync struct {
	BaseTask
}

func (c AppsSync) ID() string {
	return "apps-sync"
}

func (c AppsSync) Name() string {
	return "Sync apps from SQL to Mongo"
}

func (c AppsSync) Cron() string {
	return ""
}

const batchSize = 1000

func (c AppsSync) work() (err error) {

	db, err := sql.GetMySQLClient()
	if err != nil {
		return err
	}

	db = db.Limit(batchSize)

	var offset int
	for {

		log.Info(offset)

		var apps []sql.App

		db = db.Offset(offset).Find(&apps)
		if db.Error != nil {
			return db.Error
		}

		for _, v := range apps {
			err = v.SaveToMongo()
			log.Err(err)
		}

		if len(apps) < batchSize {
			break
		}

		offset += batchSize
	}

	return nil
}
