package elastic

import (
	"encoding/json"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/olivere/elastic/v7"
)

type Achievement struct {
	ID          string  `json:"id"` // Achievement key
	Name        string  `json:"name"`
	Icon        string  `json:"icon"`
	Description string  `json:"description"`
	Hidden      bool    `json:"hidden"`
	Completed   float64 `json:"completed"`
	AppID       int     `json:"app_id"`
	AppName     string  `json:"app_name"`
	Score       float64 `json:"score"` // Not stored, just used on frontend
}

func IndexAchievementBulk(achievements map[string]Achievement) error {

	i := map[string]interface{}{}
	for k, v := range achievements {
		i[k] = v
	}

	return indexDocuments(IndexAchievements, i)
}

func SearchAchievements(limit int, offset int, search string) (achievements []Achievement, total int64, err error) {

	client, ctx, err := GetElastic()
	if err != nil {
		log.Err(err)
		return
	}

	var query elastic.Query
	if search != "" {
		query = elastic.NewMultiMatchQuery(search, "name^3", "description^2", "app_name^1").Type("best_fields")
	}

	searchResult, err := client.Search().Index(IndexAchievements).Query(query).From(offset).Size(limit).Do(ctx)
	if err != nil {
		log.Err(err)
		return
	}

	for _, hit := range searchResult.Hits.Hits {

		var achievement Achievement
		err := json.Unmarshal(hit.Source, &achievement)
		if err != nil {
			log.Err(err)
		}

		if hit.Score != nil {
			achievement.Score = *hit.Score
		}

		achievements = append(achievements, achievement)
	}

	return achievements, searchResult.TotalHits(), err
}

//noinspection GoUnusedExportedFunction
func DeleteAndRebuildAchievementsIndex() {

	var mapping = map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "keyword",
				},
				"name": map[string]interface{}{
					"type": "text",
				},
				"icon": map[string]interface{}{
					"enabled": false,
				},
				"description": map[string]interface{}{
					"type": "text",
				},
				"hidden": map[string]interface{}{
					"type": "boolean",
				},
				"completed": map[string]interface{}{
					"type": "half_float",
				},
				"app_id": map[string]interface{}{
					"type": "integer",
				},
				"app_name": map[string]interface{}{
					"type": "text",
				},
			},
		},
	}

	err := rebuildIndex(IndexAchievements, mapping)
	log.Err(err)
}
