package tasks

import (
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type BadgesUpdateRandom struct {
	BaseTask
}

func (c BadgesUpdateRandom) ID() string {
	return "badges-update-summaries"
}

func (c BadgesUpdateRandom) Name() string {
	return "Update all badge summaries"
}

func (c BadgesUpdateRandom) Cron() string {
	return CronTimeSetBadgeCache
}

func (c BadgesUpdateRandom) work() (err error) {

	for k := range helpers.BuiltInSpecialBadges {

		err = mongo.UpdateBadgeSummary(k)
		if err != nil {
			log.Err(err)
			continue
		}
	}

	for k := range helpers.BuiltInEventBadges {

		err = mongo.UpdateBadgeSummary(k)
		if err != nil {
			log.Err(err)
			continue
		}
	}

	apps, err := mongo.PopularApps()
	if err != nil {
		return err
	}

	for _, v := range apps {

		err = mongo.UpdateBadgeSummary(v.ID)
		if err != nil {
			log.Err(err)
			continue
		}
	}

	return nil
}
