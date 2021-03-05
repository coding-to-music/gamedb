package instagram

import (
	"sync"

	"github.com/ahmdrz/goinsta/v2"
	"github.com/gamedb/gamedb/pkg/config"
)

var (
	instagram *goinsta.Instagram
	lock      sync.Mutex
)

func Client() (*goinsta.Instagram, error) {

	lock.Lock()
	defer lock.Unlock()

	if instagram == nil {

		client := goinsta.New(config.C.InstagramUsername, config.C.InstagramPassword)
		err := client.Login()
		if err != nil {
			return client, err
		}

		instagram = client
	}

	return instagram, nil
}
