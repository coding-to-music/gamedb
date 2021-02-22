package pics

import (
	"strconv"

	"github.com/Jleagle/unmarshal-go/ctypes"
)

type PICSAppConfigLaunchItem struct {
	Order               ctypes.String `json:"order"` // Int but can be "main"
	Executable          string        `json:"executable"`
	Arguments           string        `json:"arguments"`
	Description         string        `json:"description"`
	Typex               string        `json:"type"`
	OSList              string        `json:"oslist"`
	OSArch              string        `json:"osarch"`
	OwnsDLCs            []string      `json:"ownsdlc"`
	BetaKey             string        `json:"betakey"`
	WorkingDir          string        `json:"workingdir"`
	VRMode              string        `json:"vrmode"`
	VACModuleFilename   string        `json:"vacmodulefilename"`
	DescriptionLocation string        `json:"description_loc"`
	Realm               string        `json:"realm"`
}

func (li PICSAppConfigLaunchItem) OrderInt() int {
	i, _ := strconv.Atoi(string(li.Order))
	return i
}
