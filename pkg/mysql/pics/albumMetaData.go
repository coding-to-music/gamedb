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
		albums[v.DiscNumber] = append(albums[v.DiscNumber], v)
	}

	return albums
}

func (a AlbumMetaData) HasArtist() bool {
	for _, v := range a.MetaData.Artist {
		if v != "" {
			return true
		}
	}
	return false
}

func (a AlbumMetaData) HasComposer() bool {
	for _, v := range a.MetaData.Composer {
		if v != "" {
			return true
		}
	}
	return false
}

func (a AlbumMetaData) HasLabel() bool {
	for _, v := range a.MetaData.Label {
		if v != "" {
			return true
		}
	}
	return false
}

func (a AlbumMetaData) HasCredits() bool {
	for _, v := range a.MetaData.OtherCredits {
		if v != "" {
			return true
		}
	}
	return false
}

type AlbumTrack struct {
	OriginalName string `json:"originalname"`
	DiscNumber   string `json:"discnumber"`
	TrackNumber  string `json:"tracknumber"`
	Minutes      string `json:"m"`
	Seconds      string `json:"s"`
}
