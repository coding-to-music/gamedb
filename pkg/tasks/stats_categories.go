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

func (c StatsCategories) Cron() string {
	return "0 4"
}

func (c StatsCategories) work() {

}
