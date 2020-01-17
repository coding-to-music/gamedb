package pics

import (
	"github.com/Jleagle/unmarshal-go/ctypes"
)

type AlbumMetaData struct {
	CDNAssets struct {
		AlbumCover string `json:"album_cover"`
	} `json:"cdn_assets"`
	MetaData struct {
		Artist       map[string]ctypes.String `json:"artist"`
		Composer     map[string]ctypes.String `json:"composer"`
		Label        map[string]ctypes.String `json:"label"`
		OtherCredits map[string]ctypes.String `json:"othercredits"`
	} `json:"metadata"`
	Tracks map[string]AlbumTrack `json:"tracks"`
}

func (a AlbumMetaData) Albums() (albums map[string][]AlbumTrack) {

	albums = map[string][]AlbumTrack{}

	for _, v := range a.Tracks {
		if _, ok := albums[v.DiscNumber]; ok {
			albums[v.DiscNumber] = append(albums[v.DiscNumber], v)
		} else {
			albums[v.DiscNumber] = []AlbumTrack{v}
		}
	}

	return albums
}

type AlbumTrack struct {
	OriginalName string `json:"originalname"`
	DiscNumber   string `json:"discnumber"`
	TrackNumber  string `json:"tracknumber"`
	Minutes      string `json:"m"`
	Seconds      string `json:"s"`
}
