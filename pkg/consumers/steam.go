package consumers

type SteamMessage struct {
	AppIDs     []int `json:"app_ids"`
	PackageIDs []int `json:"package_ids"`
}
