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

func (app *App) fill(hit *elastic.SearchHit) error {

	err := json.Unmarshal(hit.Source, &app)
	if err != nil {
		return err
	}

	if hit.Score != nil {
		app.Score = *hit.Score
	}

	if val, ok := hit.Highlight["name"]; ok {
		if len(val) > 0 {
			app.Name = val[0]
		}
	}

	return nil
}

func IndexApp(app App) error {
	return indexDocument(IndexApps, strconv.Itoa(app.ID), app)
}

func SearchApps(limit int, offset int, search string, sorters []elastic.Sorter, aggregation bool) (apps []App, total int64, err error) {

	client, ctx, err := GetElastic()
	if err != nil {
		return apps, 0, err
	}

	searchService := client.Search().
		Index(IndexApps).
		From(offset).
		Size(limit).
		TrackTotalHits(true)

	if aggregation {
		searchService.Aggregation("type", elastic.NewTermsAggregation().Field("type").Size(10).OrderByCountDesc())
	}

	if search != "" {
		searchService.Query(appsSearchQuery(search))
	}

	if len(sorters) > 0 {
		searchService.SortBy(sorters...)
	}

	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return apps, 0, err
	}

	for _, hit := range searchResult.Hits.Hits {

		var app = App{}
		err = app.fill(hit)
		if err != nil {
			log.Err(err)
			continue
		}

		apps = append(apps, app)
	}

	return apps, searchResult.TotalHits(), err
}

func SearchAppsMini(limit int, search string) (apps []App, total int64, err error) {

	client, ctx, err := GetElastic()
	if err != nil {
		return apps, 0, err
	}

	searchService := client.Search().
		Index(IndexApps).
		Size(limit)

	if search != "" {
		searchService.Query(appsSearchQuery(search))
	}

	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return apps, 0, err
	}

	for _, hit := range searchResult.Hits.Hits {

		var app = App{}
		err = app.fill(hit)
		if err != nil {
			log.Err(err)
			continue
		}

		apps = append(apps, app)
	}

	return apps, searchResult.TotalHits(), err
}

func appsSearchQuery(search string) *elastic.BoolQuery {

	return elastic.NewBoolQuery().
		Must(
			elastic.NewMatchQuery("name", search).Fuzziness("1"),
		).
		Should(
			elastic.NewTermQuery("name", search).Boost(10), // Boost if exact match
			elastic.NewFunctionScoreQuery().AddScoreFunc(elastic.NewFieldValueFactorFunction().Modifier("sqrt").Field("players").Factor(0.0008)),
		)
}

//noinspection GoUnusedExportedFunction
func DeleteAndRebuildAppsIndex() {

	var priceProperties = map[string]interface{}{}
	for _, v := range steamapi.ProductCCs {
		priceProperties[string(v)] = map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"currency":         fieldTypeKeyword,
				"discount_percent": fieldTypeInteger,
				"final":            fieldTypeInteger,
				"individual":       fieldTypeInteger,
				"initial":          fieldTypeInteger,
			},
		}
	}

	var mapping = map[string]interface{}{
		"settings": settings,
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id":         fieldTypeInteger,
				"name":       fieldTypeText,
				"aliases":    fieldTypeText,
				"players":    fieldTypeInteger,
				"icon":       fieldTypeDisabled,
				"followers":  fieldTypeInteger,
				"score":      fieldTypeHalfFloat,
				"prices":     map[string]interface{}{"type": "object", "properties": priceProperties},
				"tags":       fieldTypeInteger,
				"genres":     fieldTypeInteger,
				"categories": fieldTypeInteger,
				"publishers": fieldTypeInteger,
				"developers": fieldTypeInteger,
				"type":       fieldTypeKeyword,
				"platforms":  fieldTypeKeyword,
			},
		},
	}

	err := rebuildIndex(IndexApps, mapping)
	log.Err(err)
}
