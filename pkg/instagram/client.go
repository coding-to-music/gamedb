package instagram

import (
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
