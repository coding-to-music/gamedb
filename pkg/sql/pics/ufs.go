package pics

type saveFiles map[string]struct {
	Path      string `json:"path"`
	Pattern   string `json:"pattern"`
	Recursive string `json:"recursive"`
	Root      string `json:"root"`
}
