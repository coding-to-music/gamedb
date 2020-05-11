package main

import (
	"encoding/json"

	"github.com/gamedb/gamedb/pkg/config"
	elasticHelper "github.com/gamedb/gamedb/pkg/elastic"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/olivere/elastic/v7"
)

func main() {

	config.Init("test", helpers.GetIP())
	log.Initialise([]log.LogName{log.LogNameTest})
	// queue.Init(queue.AllProducerDefinitions)

	client, ctx, err := elasticHelper.GetElastic()
	if err != nil {
		log.Err(err)
		return
	}

	var filters = []elastic.Query{

	}

	var musts = []elastic.Query{
		elastic.NewMatchQuery("name", "main depot"),
	}

	// Search with a term query
	searchResult, err := client.Search().
		Index(elasticHelper.IndexApps).
		Query(elastic.NewBoolQuery().Must(musts...).Filter(filters...)).
		// Sort("id", true).
		From(0).
		Size(10).
		Do(ctx)

	if err != nil {
		log.Err(err)
		return
	}

	for _, hit := range searchResult.Hits.Hits {

		var result queue.AppsSearchMessage
		err := json.Unmarshal(hit.Source, &result)
		if err != nil {
			log.Err(err)
		}

		var i []interface{}
		if hit.Score != nil {
			i = append(i, *hit.Score)
		}
		i = append(i, result.Name)

		log.Info(i...)
	}
}
