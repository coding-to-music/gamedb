package tasks

type StatsCategories struct {
	BaseTask
}

func (c StatsCategories) ID() string {
	return "update-category-stats"
}

func (c StatsCategories) Name() string {
	return "Update categories"
}

func (c StatsCategories) Group() string {
	return ""
}

func (c StatsCategories) Cron() string {
	return CronTimeCategories
}

func (c StatsCategories) work() (err error) {
	return nil
}
