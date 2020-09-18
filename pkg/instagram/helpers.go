package instagram

import (
	"bytes"

	"github.com/gamedb/gamedb/pkg/helpers"
)

func UploadInstagram(imageURL string, message string) (err error) {

	ig, err := getInstagram()
	if err != nil {
		return err
	}

	body, _, err := helpers.Get(imageURL, 0, nil)
	if err != nil {
		return err
	}

	_, err = ig.UploadPhoto(bytes.NewReader(body), message, 0, 0)
	return err
}

func UpdateBio(bio string) (err error) {

	ig, err := getInstagram()
	if err != nil {
		return err
	}

	return ig.Account.SetBiography(bio)
}
