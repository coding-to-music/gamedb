package cache

import "github.com/steam-authority/steam-authority/datastore"

var (
	playersCount = 0
)

func GetPlayersCount() (count int, err error) {

	if playersCount == 0 {

		count, err := datastore.CountPlayers()
		if err != nil {
			return 0, err
		}

		playersCount = count
	}

	return playersCount, nil
}
