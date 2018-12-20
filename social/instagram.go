package social

import (
	"net/http"

	"github.com/ahmdrz/goinsta"
	"github.com/gamedb/website/log"
	"github.com/spf13/viper"
)

var (
	ig *goinsta.Instagram
)

func InitIG() {

	ig = goinsta.New("gamedb.online", viper.GetString("INSTAGRAM_PASSWORD"))

	err := ig.Login()
	if err != nil {
		log.Err(err)
	}

	log.Info("Logged into Instagram: " + ig.Account.Username)

	resp, err := http.Get("https://vignette.wikia.nocookie.net/cswikia/images/0/01/De_cache-overview.jpg")
	if err != nil {
		log.Err(err)
	}
	//noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	_, err = ig.UploadPhoto(resp.Body, "", 0, 0)
	if err != nil {
		log.Err(err)
	}

	log.Info("IG uploaded")

}

func post() {

}
