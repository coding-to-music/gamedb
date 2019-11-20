package instagram

import (
	"net/http"
)

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
