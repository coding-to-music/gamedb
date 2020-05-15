package instagram

import (
	"github.com/gamedb/gamedb/pkg/helpers"
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
