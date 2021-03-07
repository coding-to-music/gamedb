package pics

import (
	"strconv"
)

type PICSAppConfigLaunchItem struct {
	Order               interface{} `json:"order"` // Int but can be "main", Mongo does not support ctypes package
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
	Realm               string      `json:"realm"`
}

func (li PICSAppConfigLaunchItem) OrderInt() int {

	switch val := li.Order.(type) {
	case int:
		return val
	case string:
		s, _ := strconv.Atoi(val)
		return s
	default:
		return 0
	}
}
