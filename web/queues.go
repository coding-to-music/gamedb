package web

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/dustin/go-humanize"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/queue"
)

func QueuesHandler(w http.ResponseWriter, r *http.Request) {

	template := queuesTemplate{}
	template.Fill(w, r, "Queues")

	returnTemplate(w, r, "queues", template)
	return
}

func QueuesJSONHandler(w http.ResponseWriter, r *http.Request) {

	queuesResp, err := queue.GetQeueus()
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Only expose what we need
	var queues []queuesQueue

	for _, v := range queuesResp {

		messages := v.Messages
		rate := v.MessageStats.AckDetails.Rate

		if rate > 0 && messages == 0 {
			messages = 1
		}

		queues = append(queues, queuesQueue{
			v.Name,
			humanize.Comma(int64(messages)),
			rate,
		})
	}

	// Sort by name, no datatable
	sort.Slice(queues, func(i int, j int) bool {
		return queues[i].Name > queues[i].Name
	})

	// Encode
	bytes, err := json.Marshal(queues)
	if err != nil {
		logger.Error(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

type queuesTemplate struct {
	GlobalTemplate
}

type queuesQueue struct {
	Name     string
	Messages string
	Rate     float64
}
