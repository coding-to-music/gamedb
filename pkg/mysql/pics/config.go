package pics

type PICSAppConfigLaunchItem struct {
	Order               interface{} `json:"order"` // Int but can be "main"
	Executable          string      `json:"executable"`
	Arguments           string      `json:"arguments"`
	Description         string      `json:"description"`
	Typex               string      `json:"type"`
	OSList              string      `json:"oslist"`
	OSArch              string      `json:"osarch"`
	OwnsDLCs            []string    `json:"ownsdlc"`
	BetaKey             string      `json:"betakey"`
	WorkingDir          string      `json:"workingdir"`
	VRMode              string      `json:"vrmode"`
	VACModuleFilename   string      `json:"vacmodulefilename"`
	DescriptionLocation string      `json:"description_loc"`
}
