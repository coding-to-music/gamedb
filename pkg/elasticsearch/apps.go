package elasticsearch

import (
	"encoding/json"
	"html/template"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/olivere/elastic/v7"
)

type App struct {
	AchievementsAvg     float64               `json:"achievements_avg"`
	AchievementsCount   int                   `json:"achievements_counts"`
	AchievementsIcons   []helpers.Tuple       `json:"achievements_icons"`
	Aliases             []string              `json:"aliases"`
	Background          string                `json:"background"`
	Categories          []int                 `json:"categories"`
	Developers          []int                 `json:"developers"`
	FollowersCount      int                   `json:"followers"`
	Genres              []int                 `json:"genres"`
	GroupID             string                `json:"group_id"`
	Icon                string                `json:"icon"`
	ID                  int                   `json:"id"`
	MicroTrailor        string                `json:"micro_trailor"`
	Movies              string                `json:"movies"`
	MoviesCount         int                   `json:"movies_count"`
	Name                string                `json:"name"`
	NameMarked          string                `json:"name_marked"` // Not in DB
	Platforms           []string              `json:"platforms"`
	PlayersCount        int                   `json:"players"` // Peak week
	Prices              helpers.ProductPrices `json:"prices"`
	Publishers          []int                 `json:"publishers"`
	ReleaseDateOriginal string                `json:"release_date_original"`
	ReleaseDate         int64                 `json:"release_date"`
	ReleaseDateRounded  int64                 `json:"release_date_rounded"`
	ReviewScore         float64               `json:"score"`
	ReviewsCount        int                   `json:"reviews_count"`
	Score               float64               `json:"-"` // Not in DB - Search score
	Screenshots         string                `json:"screenshots"`
	ScreenshotsCount    int                   `json:"screenshots_count"`
	Tags                []int                 `json:"tags"`
	Trend               float64               `json:"trend"`
	Type                string                `json:"type"`
	WishlistAvg         float64               `json:"wishlist_avg"`
	WishlistCount       int                   `json:"wishlist_count"`
}

func (app App) GetID() int {
	return app.ID
}

func (app App) GetPlayersPeakWeek() int {
	return app.PlayersCount
}

func (app App) GetName() string {
	return helpers.GetAppName(app.ID, app.Name)
}

func (app App) GetHeaderImage() string {
	return helpers.GetAppHeaderImage(app.ID)
}

func (app App) GetMarkedName() string {
	return helpers.GetAppName(app.ID, app.NameMarked)
}

func (app App) GetIcon() string {
	return helpers.GetAppIcon(app.ID, app.Icon)
}

func (app App) GetPath() string {
	return helpers.GetAppPath(app.ID, app.Name)
}

func (app App) GetPathAbsolute() string {
	return helpers.GetAppPathAbsolute(app.ID, app.Name)
}

func (app App) GetIconAbsolute() string {
	return helpers.GetAppIconAbsolute(app.ID, app.Name)
}

func (app App) GetType() string {
	return helpers.GetAppType(app.Type)
}

func (app App) GetPrices() (prices helpers.ProductPrices) {
	return app.Prices
}

// For an interface
func (app App) GetBackground() string {
	return app.Background
}

func (app App) GetReleaseDateNice() string {
	return time.Unix(app.ReleaseDate, 0).Format(helpers.DateYear) // No need to use helper
}

func (app App) GetReleaseDateNiceRounded() string {
	return time.Unix(app.ReleaseDateRounded, 0).Format(helpers.DateYear) // No need to use helper
}

func (app App) GetFollowers() string {
	return helpers.GetAppFollowers(app.GroupID, app.FollowersCount)
}

func (app App) GetReviewScore() string {
	return helpers.GetAppReviewScore(app.ReviewScore)
}

func (app App) GetCommunityLink() string {
	return helpers.GetAppCommunityLink(app.ID)
}

func (app App) GetStoreLink() string {
	return helpers.GetAppStoreLink(app.ID)
}

func (app App) GetMovies() (movies []helpers.AppVideo) {

	movies = []helpers.AppVideo{}

	if app.Movies == "" {
		return movies
	}

	err := json.Unmarshal([]byte(app.Movies), &movies)
	if err != nil {
		log.ErrS(err)
	}

	return movies
}

func (app App) GetScreenshots() (screenshots []helpers.AppImage) {

	screenshots = []helpers.AppImage{}

	if app.Screenshots == "" {
		return screenshots
	}

	err := json.Unmarshal([]byte(app.Screenshots), &screenshots)
	if err != nil {
		log.ErrS(err)
	}

	return screenshots
}

func (app App) GetPlayLink() template.URL {
	return helpers.GetAppPlayLink(app.ID)
}

func SearchAppsSimple(limit int, search string) (apps []App, err error) {

	apps, _, err = searchApps(limit, 0, search, false, false, nil, nil, false)
	return apps, err
}

func SearchAppsAdvanced(offset int, limit int, search string, sorters []elastic.Sorter, boolQuery *elastic.BoolQuery) (apps []App, total int64, err error) {

	return searchApps(limit, offset, search, true, true, sorters, boolQuery, false)
}

func SearchAppsRandom(filters []elastic.Query) (app App, count int64, err error) {

	boolQuery := elastic.NewBoolQuery().Filter(filters...)

	apps, count, err := searchApps(1, 0, "", true, true, nil, boolQuery, true)
	if err != nil {
		return app, count, err
	}

	if len(apps) > 0 {
		return apps[0], count, nil
	}

	return app, count, ErrNoResult

}

func GetMostExpensiveApp(code steamapi.ProductCC) (top float64, err error) {

	const aggName = "max_price"

	err = memcache.GetSetInterface(memcache.ItemAppTopPrice(code), &top, func() (interface{}, error) {

		client, ctx, err := GetElastic()
		if err != nil {
			return top, err
		}

		searchService := client.Search().
			Index(IndexApps).
			Size(0).
			Aggregation(aggName, elastic.NewMaxAggregation().Field("prices."+string(code)+".final"))

		searchResult, err := searchService.Do(ctx)
		if err != nil {
			return top, err
		}

		aggs, ok := searchResult.Aggregations.Max(aggName)
		if ok {
			return *aggs.Value, nil
		}
		return 0, err
	})

	if top > 0 {
		top = top / 100                    // Convert to cents
		top = top / 10                     // Trim expensive apps
		top = (math.Ceil(top / 100)) * 100 // Round to nearest 100
	} else {
		top = 100
	}

	return top, err
}

func GetUpcomingGames() (apps []App, err error) {

	err = memcache.GetSetInterface(memcache.ItemHomeUpcoming, &apps, func() (interface{}, error) {

		apps, _, err := SearchAppsAdvanced(
			0,
			10,
			"",
			[]elastic.Sorter{elastic.NewFieldSort("followers").Desc()},
			elastic.NewBoolQuery().Filter(
				elastic.NewRangeQuery("release_date").From(time.Now().Add(time.Hour*-12).Unix()),
				elastic.NewRangeQuery("release_date").To(time.Now().Add(time.Hour*24*14).Unix()),
			),
		)
		return apps, err
	})

	return apps, err
}

func searchApps(limit int, offset int, search string, totals bool, highlights bool, sorters []elastic.Sorter, boolQuery *elastic.BoolQuery, random bool) (apps []App, total int64, err error) {

	client, ctx, err := GetElastic()
	if err != nil {
		return apps, 0, err
	}

	searchService := client.Search().
		Index(IndexApps).
		From(offset).
		Size(limit).
		SortBy(sorters...)

	if boolQuery == nil {
		boolQuery = elastic.NewBoolQuery()
	}

	search = strings.TrimSpace(search)
	if search != "" {

		boolQuery.Must(
			elastic.NewBoolQuery().MinimumNumberShouldMatch(1).Should(
				elastic.NewTermQuery("id", search).Boost(5),
				elastic.NewMatchQuery("aliases", search),
			),
		).Should(
			elastic.NewFunctionScoreQuery().
				AddScoreFunc(elastic.NewFieldValueFactorFunction().Modifier("sqrt").Field("players").Factor(0.0001)),
			elastic.NewFunctionScoreQuery().
				AddScoreFunc(elastic.NewFieldValueFactorFunction().Modifier("sqrt").Field("followers").Factor(0.00001)),
		)

		if highlights {
			searchService.Highlight(elastic.NewHighlight().Field("name").PreTags("<mark>").PostTags("</mark>"))
		}
	}

	if random {
		searchService.Query(elastic.NewFunctionScoreQuery().BoostMode("replace").AddScoreFunc(elastic.NewRandomFunction()).Query(boolQuery))
	} else {
		searchService.Query(boolQuery)
	}

	if totals {
		searchService.TrackTotalHits(true)
	}

	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return apps, 0, err
	}

	for _, hit := range searchResult.Hits.Hits {

		var app = App{}

		err := json.Unmarshal(hit.Source, &app)
		if err != nil {
			log.ErrS(err)
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

	return apps, searchResult.TotalHits(), err
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
				"aliases":             fieldTypeText, // Used for searching
				"background":          fieldTypeDisabled,
				"categories":          fieldTypeKeyword,
				"developers":          fieldTypeKeyword,
				"followers":           fieldTypeInt32,
				"genres":              fieldTypeKeyword,
				"group_id":            fieldTypeDisabled,
				"icon":                fieldTypeDisabled,
				"id":                  fieldTypeKeyword,
				"micro_trailor":       fieldTypeDisabled,
				"movies":              fieldTypeDisabled,
				"movies_count":        fieldTypeInt32,
				"name":                fieldTypeKeyword, // Just used for sorting
				"platforms":           fieldTypeKeyword,
				"players":             fieldTypeInt32,
				"prices":              map[string]interface{}{"type": "object", "properties": priceProperties},
				"publishers":          fieldTypeKeyword,
				"release_date_original": map[string]interface{}{ // type:text allows search, type:keyword allows sorting
					"type":     "text",
					"analyzer": "gdb_lowercase_text",
					"fields": map[string]interface{}{
						"raw": fieldTypeKeyword,
					},
				},
				"release_date":         fieldTypeInt64,
				"release_date_rounded": fieldTypeInt64,
				"reviews_count":        fieldTypeInt32,
				"score":                fieldTypeFloat16,
				"screenshots":          fieldTypeDisabled,
				"screenshots_count":    fieldTypeInt32,
				"tags":                 fieldTypeKeyword,
				"type":                 fieldTypeKeyword,
				"trend":                fieldTypeKeyword,
				"wishlist_avg":         fieldTypeFloat32,
				"wishlist_count":       fieldTypeInt32,
			},
		},
	}

	rebuildIndex(IndexApps, mapping)
}
