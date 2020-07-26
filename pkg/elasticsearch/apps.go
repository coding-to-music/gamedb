package elasticsearch

import (
	"encoding/json"
	"strconv"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/olivere/elastic/v7"
)

type App struct {
	AchievementsCount int                   `json:"achievements_counts"`
	AchievementsAvg   float64               `json:"achievements_avg"`
	AchievementsIcons []helpers.Tuple       `json:"achievements_icons"`
	Aliases           []string              `json:"aliases"`
	Categories        []int                 `json:"categories"`
	Developers        []int                 `json:"developers"`
	FollowersCount    int                   `json:"followers"`
	Genres            []int                 `json:"genres"`
	Icon              string                `json:"icon"`
	ID                int                   `json:"id"`
	Name              string                `json:"name"`
	NameMarked        string                `json:"name_marked"` // Not in DB
	Platforms         []string              `json:"platforms"`
	PlayersCount      int                   `json:"players"`
	Prices            helpers.ProductPrices `json:"prices"`
	Publishers        []int                 `json:"publishers"`
	ReleaseDate       int64                 `json:"release_date"`
	ReviewScore       float64               `json:"score"`
	Score             float64               `json:"-"` // Not in DB - Search score
	Tags              []int                 `json:"tags"`
	Type              string                `json:"type"`
	Trend             float64               `json:"trend"`
	WishlistAvg       float64               `json:"wishlist_avg"`
	WishlistCount     int                   `json:"wishlist_count"`
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

func SearchApps(limit int, offset int, search string, totals bool, highlights bool, aggregation bool) (apps []App, aggregations map[string]map[string]int64, total int64, err error) {

	client, ctx, err := GetElastic()
	if err != nil {
		return apps, aggregations, 0, err
	}

	searchService := client.Search().
		Index(IndexApps).
		From(offset).
		Size(limit)

	if search != "" {

		var search2 = helpers.RegexNonAlphaNumeric.ReplaceAllString(search, "")

		searchService.Query(elastic.NewBoolQuery().
			Must(
				elastic.NewBoolQuery().MinimumNumberShouldMatch(1).Should(
					elastic.NewTermQuery("id", search2).Boost(5),
					elastic.NewMatchQuery("name", search).Boost(1),
					elastic.NewTermQuery("aliases", search2).Boost(1),
					elastic.NewPrefixQuery("name", search).Boost(0.2),
				),
			).
			Should(
				elastic.NewFunctionScoreQuery().
					AddScoreFunc(elastic.NewFieldValueFactorFunction().Modifier("sqrt").Field("players").Factor(0.005)),
			),
		)

		if highlights {
			searchService.Highlight(elastic.NewHighlight().Field("name").PreTags("<mark>").PostTags("</mark>"))
		}
	}

	if aggregation {
		searchService.Aggregation("type", elastic.NewTermsAggregation().Field("type").Size(10).OrderByCountDesc())
	}

	if totals {
		searchService.TrackTotalHits(true)
	}

	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return apps, aggregations, 0, err
	}

	if aggregation {
		aggregations = make(map[string]map[string]int64, len(searchResult.Aggregations))
		for k := range searchResult.Aggregations {
			a, ok := searchResult.Aggregations.Terms(k)
			if ok {
				aggregations[k] = make(map[string]int64, len(a.Buckets))
				for _, v := range a.Buckets {
					aggregations[k][*v.KeyAsString] = v.DocCount
				}
			}
		}
	}

	for _, hit := range searchResult.Hits.Hits {

		var app = App{}

		err := json.Unmarshal(hit.Source, &app)
		if err != nil {
			log.Err(err)
			continue
		}

		if hit.Score != nil {
			app.Score = *hit.Score
		}

		if highlights {

			app.NameMarked = app.Name
			if val, ok := hit.Highlight["name"]; ok {
				if len(val) > 0 {
					app.NameMarked = val[0]
				}
			}
		}

		apps = append(apps, app)
	}

	return apps, aggregations, searchResult.TotalHits(), err
}

func IndexApp(a App) error {
	return indexDocument(IndexApps, strconv.Itoa(a.ID), a)
}

//noinspection GoUnusedExportedFunction
func DeleteAndRebuildAppsIndex() {

	var priceProperties = map[string]interface{}{}
	for _, v := range steamapi.ProductCCs {
		priceProperties[string(v)] = map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"currency":         fieldTypeKeyword,
				"discount_percent": fieldTypeInt32,
				"final":            fieldTypeInt32,
				"individual":       fieldTypeInt32,
				"initial":          fieldTypeInt32,
				"free":             fieldTypeBool,
			},
		}
	}

	var mapping = map[string]interface{}{
		"settings": settings,
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"achievements_counts": fieldTypeInt32,
				"achievements_avg":    fieldTypeFloat16,
				"achievements_icons":  fieldTypeDisabled,
				"aliases":             fieldTypeText,
				"categories":          fieldTypeKeyword,
				"developers":          fieldTypeKeyword,
				"followers":           fieldTypeInt32,
				"genres":              fieldTypeKeyword,
				"icon":                fieldTypeDisabled,
				"id":                  fieldTypeKeyword,
				"name":                fieldTypeText,
				"platforms":           fieldTypeKeyword,
				"players":             fieldTypeInt32,
				"prices":              map[string]interface{}{"type": "object", "properties": priceProperties},
				"publishers":          fieldTypeKeyword,
				"release_date":        fieldTypeInt64,
				"score":               fieldTypeFloat16,
				"tags":                fieldTypeKeyword,
				"type":                fieldTypeFloat32,
				"trend":               fieldTypeKeyword,
				"wishlist_avg":        fieldTypeFloat32,
				"wishlist_count":      fieldTypeInt32,
			},
		},
	}

	rebuildIndex(IndexApps, mapping)
}
