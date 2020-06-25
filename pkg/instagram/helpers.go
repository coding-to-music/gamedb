package instagram

import (
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
)

func UploadInstagram(imageURL string, message string) (err error) {

	ig, err := getInstagram()
	if err != nil {
		return err
	}

	resp, err := helpers.GetWithTimeout(imageURL, 0)
	if err != nil {
		return err
	}
	defer func() {
		err := resp.Body.Close()
		log.Err(err)
	}()

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
