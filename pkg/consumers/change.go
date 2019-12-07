package consumers

type ChangesMessage struct {
	AppIDs     map[int]int `json:"app_ids"`
	PackageIDs map[int]int `json:"package_ids"`
}
