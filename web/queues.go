package web

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/Jleagle/rabbit-go/rabbit"
	"github.com/cenkalti/backoff"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/go-chi/chi"
)

func queuesRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", queuesHandler)
	r.Get("/ajax.json", queuesJSONHandler)
	return r
}

func queuesHandler(w http.ResponseWriter, r *http.Request) {

	t := queuesTemplate{}
	t.Fill(w, r, "Queues", "When new items get added to the site, they go through a queue to not overload the servers.")
	t.addAssetHighCharts()

	err := returnTemplate(w, r, "queues", t)
	log.Err(err, r)
}

type queuesTemplate struct {
	GlobalTemplate
}

func queuesJSONHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	var item = helpers.MemcacheQueues
	var jsonString string

	err := helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &jsonString, func() (interface{}, error) {

		payload := rabbit.Payload{}
		payload.LengthsAge = 3600
		payload.LengthsIncr = 60
		payload.MsgRatesAge = 3600
		payload.MsgRatesIncr = 60

		overview := rabbit.Overview{}

		// Retrying as this call can fail
		operation := func() (err error) {
			overview, err = helpers.GetRabbit().GetOverview(payload)
			return err
		}

		policy := backoff.NewExponentialBackOff()
		policy.InitialInterval = time.Second / 2
		policy.MaxElapsedTime = time.Second * 5

		err := backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Info(err) })
		if err != nil {
			return "", err
		}

		var response = queueJSONResponse{}

		// Defaults so the response get marshaled correctly
		response.Items = make([][]int64, 0)
		response.Rates = make([][]int64, 0)

		// Items
		items := overview.QueueTotals.MessagesDetails.Samples
		if len(items) > 0 {

			sort.Slice(items, func(i, j int) bool {
				return items[i].Timestamp < items[j].Timestamp
			})

			for _, v := range items {
				response.Items = append(response.Items, []int64{v.Timestamp, int64(v.Sample)})
			}
		}

		// Rates
		rates := overview.MessageStats.AckDetails.Samples
		if len(rates) > 0 {

			sort.Slice(rates, func(i, j int) bool {
				return rates[i].Timestamp < rates[j].Timestamp
			})

			var last = rates[0].Sample
			for _, v := range rates {
				response.Rates = append(response.Rates, []int64{v.Timestamp, int64(v.Sample - last)})
				last = v.Sample
			}
		}

		bytes, err := json.Marshal(response)
		return string(bytes), err
	})

	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the queues.", Error: err})
		return
	}

	err = returnJSON(w, r, []byte(jsonString))
	log.Err(err, r)
}

type queueJSONResponse struct {
	Items [][]int64 `json:"items"`
	Rates [][]int64 `json:"rate"`
}
