package elastic

import (
	"encoding/json"
	"strconv"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/olivere/elastic/v7"
)

type App struct {
	ID          int                   `json:"id"`
	Name        string                `json:"name"`
	Players     int                   `json:"players"`
	Aliases     []string              `json:"aliases"`
	Icon        string                `json:"icon"`
	Followers   int                   `json:"followers"`
	ReviewScore float64               `json:"score"`
	Prices      helpers.ProductPrices `json:"prices"`
	Tags        []int                 `json:"tags"`
	Genres      []int                 `json:"genres"`
	Categories  []int                 `json:"categories"`
	Publishers  []int                 `json:"publishers"`
	Developers  []int                 `json:"developers"`
	Type        string                `json:"type"`
	Platforms   []string              `json:"platforms"`
	Score       float64               `json:"-"`
}

func (app App) GetName() string {
	return helpers.GetAppName(app.ID, app.Name)
}

func (app App) GetIcon() string {
	return helpers.GetAppIcon(app.ID, app.Icon)
}

func (app App) GetPath() string {
	return helpers.GetAppPath(app.ID, app.Name)
}

func (app App) GetCommunityLink() string {
	return helpers.GetAppCommunityLink(app.ID)
}

func IndexApp(app App) error {
	return indexDocument(IndexApps, strconv.Itoa(app.ID), app)
}

func SearchApps(limit int, offset int, search string, sorters []elastic.Sorter) (apps []App, total int64, err error) {

	client, ctx, err := GetElastic()
	if err != nil {
		return apps, 0, err
	}

	searchService := client.Search().
		Index(IndexApps).
		From(offset).
		Size(limit).
		TrackTotalHits(true).
		Aggregation("type", elastic.NewTermsAggregation().Field("type").Size(10).OrderByCountDesc())

	if search != "" {

		var filters []elastic.Query
		var musts []elastic.Query

		musts = append(musts, elastic.NewMatchQuery("name", search))

		// https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl-function-score-query.html#function-field-value-factor
		musts = append(musts, elastic.NewFunctionScoreQuery().AddScoreFunc(
			elastic.NewFieldValueFactorFunction().Field("players").Modifier("log1p")))

		searchService.Query(elastic.NewBoolQuery().Must(musts...).Filter(filters...))
	}

	if len(sorters) > 0 {
		searchService.SortBy(sorters...)
	}

	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return apps, 0, err
	}

	for _, hit := range searchResult.Hits.Hits {

		var app App
		err := json.Unmarshal(hit.Source, &app)
		if err != nil {
			log.Err(err)
		}

		if hit.Score != nil {
			app.Score = *hit.Score
		}

		if val, ok := hit.Highlight["name"]; ok {
			if len(val) > 0 {
				app.Name = val[0]
			}
		}

		apps = append(apps, app)
	}

	return apps, searchResult.TotalHits(), err
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
		"settings": settings,
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id":      fieldTypeInteger,
				"name":    fieldTypeText,
				"aliases": fieldTypeText,
				"players": fieldTypeInteger,
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
