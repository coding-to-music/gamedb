package pics

type AlbumMetaData struct {
	CDNAssets struct {
		AlbumCover string `json:"album_cover"`
	} `json:"cdn_assets"`
	Tracks map[string]AlbumTrack `json:"tracks"`
}

type AlbumTrack struct {
	OriginalName string `json:"originalname"`
	DiscNumber   string `json:"discnumber"`
	TrackNumber  string `json:"tracknumber"`
	Minutes      string `json:"m"`
	Seconds      string `json:"s"`
}
