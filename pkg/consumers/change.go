package consumers

type ChangesMessage struct {
	AppIDs     map[uint32]uint32 `json:"app_ids"`
	PackageIDs map[uint32]uint32 `json:"package_ids"`
}
