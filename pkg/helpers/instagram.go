package helpers

import (
	"net/http"

	"github.com/ahmdrz/goinsta/v2"
	"github.com/gamedb/gamedb/pkg/config"
)

var (
	instagram *goinsta.Instagram
)

func getInstagram() (*goinsta.Instagram, error) {

	if instagram == nil {

		client := goinsta.New(
			config.Config.InstagramUsername.Get(),
			config.Config.InstagramPassword.Get(),
		)

		err := client.Login()
		if err != nil {
			return client, err
		}

		instagram = client
	}

	return instagram, nil
}

func UploadInstagram(imageURL string, message string) (err error) {

	ig, err := getInstagram()
	if err != nil {
		return err
	}

	resp, err := http.Get(imageURL)
	if err != nil {
		return err
	}

	//noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	_, err = ig.UploadPhoto(resp.Body, message, 0, 0)
	return err
}

func UpdateBio(bio string) (err error) {

	ig, err := getInstagram()
	if err != nil {
		return err
	}

	return ig.Account.SetBiography(bio)
}
