package elasticsearch

import (
	"encoding/json"
	"strconv"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/olivere/elastic/v7"
)

type Achievement struct {
	ID          string  `json:"id"` // Achievement key
	Name        string  `json:"name"`
	NameMarked  string  `json:"name_marked"`
	Icon        string  `json:"icon"`
	Description string  `json:"description"`
	Hidden      bool    `json:"hidden"`
	Completed   float64 `json:"completed"`
	AppID       int     `json:"app_id"`
	AppName     string  `json:"app_name"`
	AppOwners   int64   `json:"app_owners"`
	Score       float64 `json:"-"` // Not stored, just used on frontend
}

func (achievement Achievement) GetKey() string {
	return strconv.Itoa(achievement.AppID) + "-" + achievement.ID
}

func (achievement Achievement) GetAppName() string {
	return helpers.GetAppName(achievement.AppID, achievement.AppName)
}

func (achievement Achievement) GetIcon() string {
	return helpers.GetAchievementIcon(achievement.AppID, achievement.Icon)
}

func (achievement Achievement) GetCompleed() string {
	return helpers.FloatToString(achievement.Completed, 1)
}

func (achievement Achievement) GetAppPath() string {
	return helpers.GetAppPath(achievement.AppID, achievement.AppName) + "#achievements"
}

func IndexAchievement(achievement Achievement) error {
	return indexDocument(IndexAchievements, achievement.GetKey(), achievement)
}

func IndexAchievementBulk(achievements map[string]Achievement) error {

	// todo, add to global
	i := map[string]interface{}{}
	for k, v := range achievements {
		i[k] = v
	}

	return indexDocuments(IndexAchievements, i)
}

func SearchAppAchievements(offset int, search string, sorters []elastic.Sorter) (achievements []Achievement, total int64, err error) {

	client, ctx, err := GetElastic()
	if err != nil {
		return achievements, 0, err
	}

	searchService := client.Search().
		Index(IndexAchievements).
		From(offset).
		Size(100).
		TrackTotalHits(true).
		SortBy(sorters...)

	if search != "" {

		searchService.Query(elastic.NewBoolQuery().
			Must(
				elastic.NewBoolQuery().MinimumNumberShouldMatch(1).Should(
					elastic.NewMatchQuery("name", search).Boost(3),
					elastic.NewMatchQuery("description", search).Boost(2),
					elastic.NewMatchQuery("app_name", search).Boost(1),
					elastic.NewPrefixQuery("name", search).Boost(0.3),
					elastic.NewPrefixQuery("description", search).Boost(0.2),
					elastic.NewPrefixQuery("app_name", search).Boost(0.1),
				),
			),
		)

		searchService.Highlight(elastic.NewHighlight().Field("name").Field("description").Field("app_name").PreTags("<mark>").PostTags("</mark>"))
	}

	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return achievements, 0, err
	}

	for _, hit := range searchResult.Hits.Hits {

		var achievement Achievement
		err := json.Unmarshal(hit.Source, &achievement)
		if err != nil {
			log.ErrS(err)
			continue
		}

		if hit.Score != nil {
			achievement.Score = *hit.Score
		}

		achievement.NameMarked = achievement.Name
		if val, ok := hit.Highlight["name"]; ok {
			if len(val) > 0 {
				achievement.NameMarked = val[0]
			}
		}

		if val, ok := hit.Highlight["description"]; ok {
			if len(val) > 0 {
				achievement.Description = val[0]
			}
		}

		if val, ok := hit.Highlight["app_name"]; ok {
			if len(val) > 0 {
				achievement.AppName = val[0]
			}
		}

		achievements = append(achievements, achievement)
	}

	return achievements, searchResult.TotalHits(), err
}

//noinspection GoUnusedExportedFunction
func DeleteAndRebuildAchievementsIndex() {

	var mapping = map[string]interface{}{
		"settings": settings,
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": fieldTypeKeyword,
				"name": map[string]interface{}{
					"type":       "text",
					"normalizer": "gdb_lowercase",
				},
				"icon": fieldTypeDisabled,
				"description": map[string]interface{}{
					"type":       "text",
					"normalizer": "gdb_lowercase",
				},
				"hidden":    fieldTypeBool,
				"completed": fieldTypeFloat16,
				"app_id":    fieldTypeInt32,
				"app_name": map[string]interface{}{
					"type":       "text",
					"normalizer": "gdb_lowercase",
				},
				"app_owners": fieldTypeInt64,
			},
		},
	}

	rebuildIndex(IndexAchievements, mapping)
}
