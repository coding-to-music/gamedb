package tasks

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
)

type DevCodeRun struct {
	BaseTask
}

func (c DevCodeRun) ID() string {
	return "run-dev-code"
}

func (c DevCodeRun) Name() string {
	return "Run dev code"
}

func (c DevCodeRun) Cron() string {
	return ""
}

func (c DevCodeRun) work() (err error) {

	db, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err)
		return
	}

	db = db.Limit(1000)

	var offset int
	for {

		var packages []sql.Package

		db = db.Offset(offset).Find(&packages)
		if db.Error != nil {
			log.Err(db.Error)
			return
		}

		log.Info(offset)

		for _, v := range packages {
			err = v.SaveToMongo()
			log.Err(err)
		}

		if len(packages) < 1000 {
			break
		}

		offset += 1000
	}

	return nil
}
