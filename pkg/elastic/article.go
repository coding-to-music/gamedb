package elastic

import (
	"encoding/json"
	"strconv"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/olivere/elastic/v7"
)

type Article struct {
	ID      int64   `json:"id"`
	Title   string  `json:"title"`
	Body    string  `json:"body"`
	AppID   int     `json:"app_id"`
	AppName string  `json:"app_name"`
	AppIcon string  `json:"app_icon"`
	Time    int64   `json:"time"`
	Score   float64 `json:"-"`
}

func (article Article) GetBody() string {
	return helpers.GetArticleBody(article.Body)
}

func IndexArticle(article Article) error {
	return indexDocument(IndexArticles, strconv.FormatInt(article.ID, 10), article)
}

func IndexArticlesBulk(articles map[string]Article) error {

	i := map[string]interface{}{}
	for k, v := range articles {
		i[k] = v
	}

	return indexDocuments(IndexArticles, i)
}

func SearchArticles(offset int, sorters []elastic.Sorter, search string) (articles []Article, total int64, err error) {

	client, ctx, err := GetElastic()
	if err != nil {
		return articles, 0, err
	}

	searchService := client.Search().
		Index(IndexArticles).
		From(offset).
		Size(100).
		TrackTotalHits(true).
		Highlight(elastic.NewHighlight().Field("title").Field("app_name").PreTags("<mark>").PostTags("</mark>")).
		SortBy(sorters...)

	if search != "" {
		searchService.Query(elastic.NewBoolQuery().
			Must(
				elastic.NewBoolQuery().MinimumNumberShouldMatch(1).Should(
					elastic.NewMatchQuery("title", search).Fuzziness("1").Boost(2),
					elastic.NewMatchQuery("app_name", search).Fuzziness("1").Boost(1),
				),
			).
			Should(
				elastic.NewTermQuery("title", search).Boost(2), // Exact match
			),
		)
	}

	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return articles, 0, err
	}

	for _, hit := range searchResult.Hits.Hits {

		var article Article
		err := json.Unmarshal(hit.Source, &article)
		if err != nil {
			log.Err(err)
			continue
		}

		if hit.Score != nil {
			article.Score = *hit.Score
		}

		if val, ok := hit.Highlight["title"]; ok {
			if len(val) > 0 {
				article.Title = val[0]
			}
		}

		if val, ok := hit.Highlight["app_name"]; ok {
			if len(val) > 0 {
				article.AppName = val[0]
			}
		}

		articles = append(articles, article)
	}

	return articles, searchResult.TotalHits(), nil
}

//noinspection GoUnusedExportedFunction
func DeleteAndRebuildArticlesIndex() {

	var mapping = map[string]interface{}{
		"settings": settings,
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id":       fieldTypeDisabled,
				"title":    fieldTypeText,
				"body":     fieldTypeDisabled,
				"app_id":   fieldTypeDisabled,
				"app_name": fieldTypeText,
				"app_icon": fieldTypeDisabled,
				"time":     fieldTypeLong,
			},
		},
	}

	err := rebuildIndex(IndexArticles, mapping)
	log.Err(err)
}
