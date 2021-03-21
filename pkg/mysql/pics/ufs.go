package pics

import (
	"strings"

	"github.com/Jleagle/unmarshal-go/ctypes"
)

type saveFiles map[string]*saveFile

type saveFile struct {
	Path      ctypes.String     `json:"path"`
	Pattern   string            `json:"pattern"`
	Platforms map[string]string `json:"platforms"`
	Recursive string            `json:"recursive"`
	Root      string            `json:"root"`
}

func (sf saveFile) GetPlatforms() (platforms string) {

	var ret []string
	for _, v := range sf.Platforms {
		ret = append(ret, v)
	}

	return strings.Join(ret, ", ")
}
