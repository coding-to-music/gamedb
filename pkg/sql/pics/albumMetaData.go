package pics

import (
	"github.com/Jleagle/unmarshal-go/ctypes"
)

type AlbumMetaData struct {
	CDNAssets struct {
		AlbumCover string `json:"album_cover"`
	} `json:"cdn_assets"`
	MetaData struct {
		Artist       map[string]ctypes.CString `json:"artist"`
		Composer     map[string]ctypes.CString `json:"composer"`
		Label        map[string]ctypes.CString `json:"label"`
		OtherCredits map[string]ctypes.CString `json:"othercredits"`
	} `json:"metadata"`
	Tracks map[string]AlbumTrack `json:"tracks"`
}

type AlbumTrack struct {
	OriginalName string `json:"originalname"`
	DiscNumber   string `json:"discnumber"`
	TrackNumber  string `json:"tracknumber"`
	Minutes      string `json:"m"`
	Seconds      string `json:"s"`
}
