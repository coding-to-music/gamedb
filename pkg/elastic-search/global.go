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
	GlobalTypePlayer      = "player"
)

type Global struct {
	ID         string `json:"id"` // String as can contain mixed types
	Name       string `json:"name"`
	NameMarked string `json:"name_marked"`
	Icon       string `json:"icon"`
	// AppID      int     `json:"app_id"`
	Type  string  `json:"type"`
	Score float64 `json:"-"`
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

func (global Global) GetNameMarked() string {

	switch global.Type {
	case GlobalTypeApp:

		i, _ := strconv.Atoi(global.ID)
		return helpers.GetAppName(i, global.NameMarked)

	case GlobalTypePlayer:

		i64, _ := strconv.ParseInt(global.ID, 10, 64)
		return helpers.GetPlayerName(i64, global.NameMarked)

	case GlobalTypeGroup:

		return helpers.GetGroupName(global.ID, global.NameMarked)

	default:
		return global.NameMarked
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

func SearchGlobal(limit int, offset int, search string) (items []Global, aggregations map[string]map[string]int64, total int64, err error) {

	client, ctx, err := GetElastic()
	if err != nil {
		return items, aggregations, 0, err
	}

	searchService := client.Search().
		Index(IndexGlobal).
		From(offset).
		// Aggregation("type", elastic.NewTermsAggregation().Field("type").Size(10).OrderByCountDesc()).
		TrackTotalHits(true).
		Size(limit)

	if search != "" {

		var search2 = helpers.RegexNonAlphaNumeric.ReplaceAllString(search, "")

		searchService.Query(
			elastic.NewBoolQuery().MinimumNumberShouldMatch(1).Should(
				elastic.NewTermQuery("id", search2).Boost(5),
				elastic.NewMatchQuery("name", search).Boost(1),
				elastic.NewPrefixQuery("name", search).Boost(0.2),
			),
		)

		searchService.Highlight(elastic.NewHighlight().Field("name").PreTags("<mark>").PostTags("</mark>"))
	}

	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return items, aggregations, 0, err
	}

	for _, hit := range searchResult.Hits.Hits {

		var item = Global{}

		err := json.Unmarshal(hit.Source, &item)
		if err != nil {
			log.Err(err)
			continue
		}

		if hit.Score != nil {
			item.Score = *hit.Score
		}

		item.NameMarked = item.Name
		if val, ok := hit.Highlight["name"]; ok {
			if len(val) > 0 {
				item.NameMarked = val[0]
			}
		}

		items = append(items, item)
	}

	return items, aggregations, searchResult.TotalHits(), err
}

func IndexGlobalItem(item Global) error {
	if item.Name == "" {
		return nil
	}
	return indexDocument(IndexGlobal, item.Type+"-"+item.ID, item)
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
				"app_id": fieldTypeInteger,
				"type":   fieldTypeKeyword,
			},
		},
	}

	rebuildIndex(IndexGlobal, mapping)
}
