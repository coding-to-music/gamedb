package pics

import (
	"github.com/Jleagle/unmarshal-go/ctypes"
)

type saveFiles map[string]struct {
	Path      ctypes.CString    `json:"path"`
	Pattern   string            `json:"pattern"`
	Platforms map[string]string `json:"platforms"`
	Recursive string            `json:"recursive"`
	Root      string            `json:"root"`
}
