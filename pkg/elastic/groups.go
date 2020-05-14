package elastic

import (
	"encoding/json"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/olivere/elastic/v7"
)

type Group struct {
	Name         string `json:"name"`
	URL          string `json:"url"`
	Abbreviation string `json:"abbreviation"`
	Headline     string `json:"headline"`
	Icon         string `json:"icon"`
	Members      int    `json:"members"`
	Trend        int64  `json:"trend"`
}

func SearchGroups(limit int, offset int, search string, sorter elastic.Sorter) (groups []Group, total int64, err error) {

	client, ctx, err := GetElastic()
	if err != nil {
		log.Err(err)
		return
	}

	searchService := client.Search().Index(IndexGroups)

	if search != "" {
		searchService.Query(elastic.NewMultiMatchQuery(search, "name^3", "url^2", "abbreviation^2", "headline^1").Type("best_fields"))
	}

	if sorter != nil {
		searchService.SortBy(sorter)
	}

	searchResult, err := client.Search().Index(IndexGroups).From(offset).Size(limit).Do(ctx)
	if err != nil {
		log.Err(err)
		return
	}

	for _, hit := range searchResult.Hits.Hits {

		var group Group
		err := json.Unmarshal(hit.Source, &group)
		if err != nil {
			log.Err(err)
		}

		groups = append(groups, group)
	}

	return groups, searchResult.TotalHits(), err
}

//noinspection GoUnusedExportedFunction
func DeleteAndRebuildGroupsIndex() {

	var mapping = map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "text",
				},
				"url": map[string]interface{}{
					"type": "text",
				},
				"abbreviation": map[string]interface{}{
					"type": "text",
				},
				"headline": map[string]interface{}{
					"type": "text",
				},
				"icon": map[string]interface{}{
					"enabled": false,
				},
				"members": map[string]interface{}{
					"type": "integer",
				},
				"trend": map[string]interface{}{
					"type": "integer",
				},
			},
		},
	}

	err := rebuildIndex(IndexGroups, mapping)
	log.Err(err)
}
