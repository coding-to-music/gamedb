package helpers

import (
	"net/http"
	"time"

	"github.com/ahmdrz/goinsta/v2"
	"github.com/cenkalti/backoff"
	"github.com/gamedb/website/pkg/config"
	"github.com/gamedb/website/pkg/log"
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

func UploadInstagram(url string, message string) (err error) {

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	//noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	ig, err := getInstagram()
	if err != nil {
		return err
	}

	operation := func() (err error) {

		_, err = ig.UploadPhoto(resp.Body, message, 0, 0)
		return err
	}

	policy := backoff.NewExponentialBackOff()
	policy.InitialInterval = time.Second * 10

	return backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Info(err) })
}
