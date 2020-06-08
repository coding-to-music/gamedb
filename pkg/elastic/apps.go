package elastic

import (
	"encoding/json"
	"strconv"
	"strings"

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
		Highlight(elastic.NewHighlight().Field("name").PreTags("<mark>").PostTags("</mark>")).
		Aggregation("type", elastic.NewTermsAggregation().Field("type").Size(10).OrderByCountDesc()).
		SortBy(sorters...).
		Query(appsSearchQuery(search))

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

func appsSearchQuery(search string) (ret *elastic.BoolQuery) {

	if strings.TrimSpace(search) == "" {
		return ret
	}

	return elastic.NewBoolQuery().
		Must(
			elastic.NewBoolQuery().MinimumNumberShouldMatch(1).Should(
				elastic.NewTermQuery("id", search).Boost(3),
				elastic.NewMatchQuery("name", search).Fuzziness("1").Boost(3),
				elastic.NewMatchQuery("aliases", search).Fuzziness("1").Boost(1),
			),
		).
		Should(
			elastic.NewTermQuery("name", search).Boost(10),
			elastic.NewFunctionScoreQuery().AddScoreFunc(elastic.NewFieldValueFactorFunction().Modifier("sqrt").Field("players").Factor(0.05)),
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
				"id":         fieldTypeKeyword,
				"name":       fieldTypeText,
				"aliases":    fieldTypeText,
				"players":    fieldTypeInteger,
				"icon":       fieldTypeDisabled,
				"followers":  fieldTypeInteger,
				"score":      fieldTypeHalfFloat,
				"prices":     map[string]interface{}{"type": "object", "properties": priceProperties},
				"tags":       fieldTypeKeyword,
				"genres":     fieldTypeKeyword,
				"categories": fieldTypeKeyword,
				"publishers": fieldTypeKeyword,
				"developers": fieldTypeKeyword,
				"type":       fieldTypeKeyword,
				"platforms":  fieldTypeKeyword,
			},
		},
	}

	rebuildIndex(IndexApps, mapping)
}
