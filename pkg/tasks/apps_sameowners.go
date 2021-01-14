package tasks

type AppsSameOwners struct {
	BaseTask
}

func (c AppsSameOwners) ID() string {
	return "apps-sameowners"
}

func (c AppsSameOwners) Name() string {
	return "Queue a game to scan same owners"
}

func (c AppsSameOwners) Group() TaskGroup {
	return TaskGroupApps
}

func (c AppsSameOwners) Cron() TaskTime {
	return ""
}

func (c AppsSameOwners) work() (err error) {

	return nil
}
