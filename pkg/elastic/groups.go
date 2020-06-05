package elastic

import (
	"encoding/json"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/olivere/elastic/v7"
)

type Group struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	URL          string  `json:"url"`
	Abbreviation string  `json:"abbreviation"`
	Headline     string  `json:"headline"`
	Icon         string  `json:"icon"`
	Members      int     `json:"members"`
	Trend        int64   `json:"trend"`
	Error        bool    `json:"error"`
	Score        float64 `json:"-"`
}

func IndexGroup(group Group) error {
	return indexDocument(IndexGroups, group.ID, group)
}

func SearchGroups(offset int, sorters []elastic.Sorter, search string, errors string) (groups []Group, total int64, err error) {

	client, ctx, err := GetElastic()
	if err != nil {
		return groups, 0, err
	}

	var query = elastic.NewBoolQuery()
	if search != "" {

		query.Must(elastic.NewBoolQuery().MinimumNumberShouldMatch(1).Should(
			elastic.NewTermQuery("id", search).Boost(3),
			elastic.NewMatchQuery("name", search).Fuzziness("1").Boost(3),
			elastic.NewMatchQuery("abbreviation", search).Fuzziness("1").Boost(2),
			elastic.NewMatchQuery("url", search).Fuzziness("1").Boost(2),
		))

		query.Should(
			elastic.NewTermQuery("name", search).Boost(10),
			elastic.NewFunctionScoreQuery().AddScoreFunc(elastic.NewFieldValueFactorFunction().Modifier("sqrt").Field("members").Factor(0.003)),
		)
	}

	if errors == "0" || errors == "1" {
		query.Filter(elastic.NewTermQuery("error", errors == "0"))
	}

	searchService := client.Search().
		Index(IndexGroups).
		From(offset).
		Size(100).
		TrackTotalHits(true).
		Highlight(elastic.NewHighlight().Field("name").PreTags("<mark>").PostTags("</mark>")).
		SortBy(sorters...).
		Query(query).
		SearchType("dfs_query_then_fetch") // Improves acuracy with multiple shards

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

		if hit.Score != nil {
			group.Score = *hit.Score
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
