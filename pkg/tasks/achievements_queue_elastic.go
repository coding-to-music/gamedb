package tasks

type AchievementsQueueElastic struct {
	BaseTask
}

func (c AchievementsQueueElastic) ID() string {
	return "achievements-queue-elastic"
}

func (c AchievementsQueueElastic) Name() string {
	return "Queue all achievements to Elastic (Todo!)"
}

func (c AchievementsQueueElastic) Cron() string {
	return ""
}

func (c AchievementsQueueElastic) work() (err error) {

	return nil
}
