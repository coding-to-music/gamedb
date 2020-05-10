package main

import (
	"encoding/json"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/search"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/olivere/elastic/v7"
)

func main() {

	config.Init("test", helpers.GetIP())
	log.Initialise([]log.LogName{log.LogNameTest})
	queue.Init(queue.AllProducerDefinitions)

	err := queue.ProduceSearch(queue.SearchMessage{
		ID:      578080,
		Name:    "PLAYERUNKNOWN'S BATTLEGROUNDS",
		Aliases: []string{"pubg"},
		Type:    search.SearchTypeApp,
	})

	log.Err(err)

	client, ctx, err := search.GetElastic()
	if err != nil {
		log.Err(err)
		return
	}

	// Search with a term query
	searchResult, err := client.Search().
		Index(search.IndexName).
		Query(elastic.NewMatchQuery("Name", "Dota 2")).
		// Sort("ID", true).
		From(0).
		Size(100).
		Do(ctx)

	if err != nil {
		log.Err(err)
		return
	}

	var results []search.SearchResult

	log.Info(searchResult)

	for _, hit := range searchResult.Hits.Hits {

		// log.Info(string(hit.Source))

		// Deserialize hit.Source into a Tweet (could also be just a map[string]interface{}).
		var result search.SearchResult
		err := json.Unmarshal(hit.Source, &result)
		if err != nil {
			log.Err(err)
		}

		results = append(results, result)
	}

	log.Info("Queued")
}
