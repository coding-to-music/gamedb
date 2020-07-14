package elastic_search

import (
	"encoding/json"
	"strconv"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/olivere/elastic/v7"
)

const (
	GlobalTypeAchievement = "achievement"
	GlobalTypeApp         = "app"
	GlobalTypeArticle     = "article"
	GlobalTypeGroup       = "group"
	GlobalTypePackage     = "package"
	GlobalTypePlayer      = "player"
)

type Global struct {
	ID         string  `json:"id"` // String as can contain mixed types
	Name       string  `json:"name"`
	NameMarked string  `json:"name_marked"`
	Icon       string  `json:"icon"`
	AppID      int     `json:"app_id"`
	Type       string  `json:"type"`
	Score      float64 `json:"-"`
}

func (global Global) GetName() string {

	switch global.Type {
	case GlobalTypeApp:

		i, _ := strconv.Atoi(global.ID)
		return helpers.GetAppName(i, global.Name)

	case GlobalTypePlayer:

		i64, _ := strconv.ParseInt(global.ID, 10, 64)
		return helpers.GetPlayerName(i64, global.Name)

	case GlobalTypeGroup:

		return helpers.GetGroupName(global.ID, global.Name)

	default:
		return global.Name
	}
}

func (global Global) GetIcon() string {

	switch global.Type {
	case GlobalTypeApp:

		i, _ := strconv.Atoi(global.ID)
		return helpers.GetAppIcon(i, global.Icon)

	case GlobalTypePlayer:

		return helpers.GetPlayerAvatar(global.Icon)

	case GlobalTypeGroup:

		return helpers.GetGroupIcon(global.Icon)

	default:
		return global.Name
	}
}

func (global Global) GetPath() string {

	switch global.Type {
	case GlobalTypeApp:

		i, _ := strconv.Atoi(global.ID)
		return helpers.GetAppPath(i, global.Name)

	case GlobalTypePlayer:

		i64, _ := strconv.ParseInt(global.ID, 10, 64)
		return helpers.GetPlayerPath(i64, global.Name)

	case GlobalTypeGroup:

		return helpers.GetGroupPath(global.ID, global.Name)

	default:
		return global.Name
	}
}

func SearchGlobal(limit int, offset int, search string, sorters []elastic.Sorter) (apps []App, total int64, err error) {

	client, ctx, err := GetElastic()
	if err != nil {
		return apps, 0, err
	}

	searchService := client.Search().
		Index(IndexGlobal).
		From(offset).
		Aggregation("type", elastic.NewTermsAggregation().Field("type").Size(10).OrderByCountDesc()).
		Highlight(elastic.NewHighlight().Field("name").PreTags("<mark>").PostTags("</mark>")).
		TrackTotalHits(true).
		Size(limit)

	if search != "" {

		var search2 = helpers.RegexNonAlphaNumeric.ReplaceAllString(search, "")

		searchService.Query(elastic.NewBoolQuery().
			Must(
				elastic.NewBoolQuery().MinimumNumberShouldMatch(1).Should(
					elastic.NewTermQuery("id", search2).Boost(5),
					elastic.NewMatchQuery("name", search).Boost(1),
					elastic.NewPrefixQuery("name", search).Boost(0.2),
				),
			).
			Should(
				elastic.NewTermQuery("type", GlobalTypePlayer).Boost(1.2),
				elastic.NewTermQuery("type", GlobalTypeApp).Boost(1.1),
			),
		)
	}

	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return apps, 0, err
	}

	if len(sorters) > 0 {
		searchService.SortBy(sorters...)
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

		app.NameMarked = app.Name
		if val, ok := hit.Highlight["name"]; ok {
			if len(val) > 0 {
				app.NameMarked = val[0]
			}
		}

		apps = append(apps, app)
	}

	return apps, searchResult.TotalHits(), err
}

func IndexGlobalItem(global Global) error {
	return indexDocument(IndexGlobal, global.Type+"-"+global.ID, global)
}

//noinspection GoUnusedExportedFunction
func DeleteAndRebuildGlobalIndex() {

	var mapping = map[string]interface{}{
		"settings": settings,
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id":     fieldTypeKeyword,
				"name":   fieldTypeText,
				"icon":   fieldTypeDisabled,
				"app_id": fieldTypeKeyword,
				"type":   fieldTypeKeyword,
			},
		},
	}

	rebuildIndex(IndexGlobal, mapping)
}
