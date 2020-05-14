package elastic

import (
	"encoding/json"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/olivere/elastic/v7"
)

type App struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Players int      `json:"players"`
	Aliases []string `json:"aliases"`
	// Icon       string                `json:"icon"`
	// Followers  int                   `json:"followers"`
	// Score      float64               `json:"score"`
	// Prices     helpers.ProductPrices `json:"prices"`
	// Tags       []int                 `json:"tags"`
	// Genres     []int                 `json:"genres"`
	// Categories []int                 `json:"categories"`
	// Publishers []int                 `json:"publishers"`
	// Developers []int                 `json:"developers"`
	// Type       string                `json:"type"`
	// Platforms  []string              `json:"platforms"`
}

func SearchApps(limit int, query string) (apps []App, err error) {

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
		Index(IndexApps).
		Query(elastic.NewBoolQuery().Must(musts...).Filter(filters...)).
		From(0).
		Size(limit).
		Do(ctx)

	if err != nil {
		log.Err(err)
		return
	}

	for _, hit := range searchResult.Hits.Hits {

		var app App
		err := json.Unmarshal(hit.Source, &app)
		if err != nil {
			log.Err(err)
		}

		apps = append(apps, app)
	}

	return apps, err
}

//noinspection GoUnusedExportedFunction
func DeleteAndRebuildAppsIndex() {

	var priceProperties = map[string]interface{}{}
	for _, v := range steamapi.ProductCCs {
		priceProperties[string(v)] = map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"currency":         map[string]interface{}{"type": "keyword"},
				"discount_percent": map[string]interface{}{"type": "integer"},
				"final":            map[string]interface{}{"type": "integer"},
				"individual":       map[string]interface{}{"type": "integer"},
				"initial":          map[string]interface{}{"type": "integer"},
			},
		}
	}

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

	err := rebuildIndex(IndexApps, mapping)
	log.Err(err)
}
