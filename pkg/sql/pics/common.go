package pics

type Associations map[string]struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
