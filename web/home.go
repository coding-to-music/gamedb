package web

import (
	"net/http"
	"sync"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	var wg sync.WaitGroup
	var err error

	var playersCount int
	wg.Add(1)
	go func() {

		playersCount, err = datastore.CountPlayers()
		if err != nil {
			logger.Error(err)
		}

		wg.Done()

	}()

	var pricesCount int
	wg.Add(1)
	go func() {

		pricesCount, err = datastore.CountPrices()
		if err != nil {
			logger.Error(err)
		}

		wg.Done()

	}()

	var ranksCount int
	wg.Add(1)
	go func() {

		ranksCount, err = datastore.CountRanks()
		if err != nil {
			logger.Error(err)
		}

		wg.Done()

	}()

	wg.Wait()

	template := homeTemplate{}
	template.Fill(r, "Home")

	template.PlayersCount = playersCount
	template.PricesCount = pricesCount
	template.RanksCount = ranksCount

	returnTemplate(w, r, "home", template)
}

type homeTemplate struct {
	GlobalTemplate
	PlayersCount int
	PricesCount  int
	RanksCount   int
}
