package consumers

type PackageMessage struct {
	ID           int                    `json:"id"`
	ChangeNumber int                    `json:"change_number"`
	VDF          map[string]interface{} `json:"vdf"`
}
