package elastic

import (
	"encoding/json"
	"strconv"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/olivere/elastic/v7"
)

type Player struct {
	ID          int64  `json:"id"`
	PersonaName string `json:"name"`
	VanityURL   string `json:"url"`
	Flag        string `json:"flag"`
}

func IndexPlayer(player Player) error {
	return indexDocument(IndexPlayers, strconv.FormatInt(player.ID, 10), player)
}

func SearchPlayers(limit int, query string) (players []Player, err error) {

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
		Index(IndexPlayers).
		Query(elastic.NewBoolQuery().Must(musts...).Filter(filters...)).
		From(0).
		Size(limit).
		Do(ctx)

	if err != nil {
		log.Err(err)
		return
	}

	for _, hit := range searchResult.Hits.Hits {

		var player Player
		err := json.Unmarshal(hit.Source, &player)
		if err != nil {
			log.Err(err)
		}

		players = append(players, player)
	}

	return players, err
}

//noinspection GoUnusedExportedFunction
func DeleteAndRebuildPlayersIndex() {

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
				"aliases": map[string]interface{}{
					"type": "text",
				},
				"players": map[string]interface{}{
					"type": "integer",
				},
				// "icon": map[string]interface{}{
				// 	"enabled": false,
				// },
				// "followers": map[string]interface{}{
				// 	"type": "integer",
				// },
				// "score": map[string]interface{}{
				// 	"type": "half_float",
				// },
				// "prices": map[string]interface{}{
				// 	"type":       "object",
				// 	"properties": priceProperties,
				// },
				// "tags": map[string]interface{}{
				// 	"type": "integer",
				// },
				// "genres": map[string]interface{}{
				// 	"type": "integer",
				// },
				// "categories": map[string]interface{}{
				// 	"type": "integer",
				// },
				// "publishers": map[string]interface{}{
				// 	"type": "integer",
				// },
				// "developers": map[string]interface{}{
				// 	"type": "integer",
				// },
				// "type": map[string]interface{}{
				// 	"type": "keyword",
				// },
				// "platforms": map[string]interface{}{
				// 	"type": "keyword",
				// },
			},
		},
	}

	err := rebuildIndex(IndexPlayers, mapping)
	log.Err(err)
}
