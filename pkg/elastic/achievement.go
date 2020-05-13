package elastic

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/olivere/elastic/v7"
)

type Achievement struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func SearchAchievements(limit int, query string) (achievements []Achievement, err error) {

	var filters []elastic.Query
	var musts []elastic.Query

	musts = append(musts, elastic.NewMatchQuery("name", query))

	// https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl-function-score-query.html#function-field-value-factor
	musts = append(musts, elastic.NewFunctionScoreQuery().AddScoreFunc(
		elastic.NewFieldValueFactorFunction().Field("players").Modifier("log1p")))

	client, ctx, err := GetElastic()
	if err != nil {
		log.Err(err)
		return
	}

	searchResult, err := client.Search().
		Index(IndexAchievements).
		Query(elastic.NewBoolQuery().Must(musts...).Filter(filters...)).
		From(0).
		Size(limit).
		Do(ctx)

	if err != nil {
		log.Err(err)
		return
	}

	for _, hit := range searchResult.Hits.Hits {

		var achievement Achievement
		err := json.Unmarshal(hit.Source, &achievement)
		if err != nil {
			log.Err(err)
		}

		achievements = append(achievements, achievement)
	}

	return achievements, err
}

func DeleteAndRebuildAchievementsIndex() {

	var mapping = map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "integer",
				},
				"name": map[string]interface{}{
					"type": "text",
				},
				"description": map[string]interface{}{
					"type": "text",
				},
			},
		},
	}

	client, ctx, err := GetElastic()
	if err != nil {
		log.Err(err)
		return
	}

	_, err = client.DeleteIndex(IndexAchievements).Do(ctx)
	if err != nil {
		log.Err(err)
		return
	}

	time.Sleep(time.Second)

	createIndex, err := client.CreateIndex(IndexAchievements).BodyJson(mapping).Do(ctx)
	if err != nil {
		log.Err(err)
		return
	}

	if !createIndex.Acknowledged {
		log.Warning(errors.New("not acknowledged"))
	}
}
