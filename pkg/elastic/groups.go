package elastic

import (
	"encoding/json"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/olivere/elastic/v7"
)

type Group struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	URL          string `json:"url"`
	Abbreviation string `json:"abbreviation"`
	Headline     string `json:"headline"`
	Icon         string `json:"icon"`
	Members      int    `json:"members"`
	Trend        int64  `json:"trend"`
	Error        bool   `json:"error"`
}

func IndexGroup(group Group) error {
	return indexDocument(IndexGroups, group.ID, group)
}

func SearchGroups(limit int, offset int, search string, sorters []elastic.Sorter) (groups []Group, total int64, err error) {

	client, ctx, err := GetElastic()
	if err != nil {
		return groups, 0, err
	}

	searchService := client.Search().
		Index(IndexGroups).
		From(offset).
		Size(limit).
		TrackTotalHits(true).
		Highlight(elastic.NewHighlight().Field("name").PreTags("<mark>").PostTags("</mark>"))

	if search != "" {
		searchService.Query(elastic.NewBoolQuery().MinimumNumberShouldMatch(2).Should(
			elastic.NewMatchQuery("name", search).Fuzziness("1").Boost(3),
			elastic.NewMatchQuery("abbreviation", search).Fuzziness("1").Boost(2),
			elastic.NewMatchQuery("url", search).Fuzziness("1").Boost(2),
			elastic.NewFunctionScoreQuery().AddScoreFunc(elastic.NewFieldValueFactorFunction().Modifier("sqrt").Field("members").Factor(0.001)),
		))
	}

	if sorters != nil && len(sorters) > 0 {
		searchService.SortBy(sorters...)
	}

	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return groups, 0, err
	}

	for _, hit := range searchResult.Hits.Hits {

		var group Group
		err := json.Unmarshal(hit.Source, &group)
		if err != nil {
			log.Err(err)
			continue
		}

		if val, ok := hit.Highlight["name"]; ok {
			if len(val) > 0 {
				group.Name = val[0]
			}
		}

		groups = append(groups, group)
	}

	return groups, searchResult.TotalHits(), nil
}

//noinspection GoUnusedExportedFunction
func DeleteAndRebuildGroupsIndex() {

	var mapping = map[string]interface{}{
		"settings": settings,
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id":           fieldTypeKeyword,
				"name":         fieldTypeText,
				"url":          fieldTypeText,
				"abbreviation": fieldTypeText,
				"headline":     fieldTypeText,
				"icon":         fieldTypeDisabled,
				"members":      fieldTypeInteger,
				"trend":        fieldTypeInteger,
				"error":        fieldTypeBool,
			},
		},
	}

	err := rebuildIndex(IndexGroups, mapping)
	log.Err(err)
}
