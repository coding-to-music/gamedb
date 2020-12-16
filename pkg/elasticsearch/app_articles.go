package elasticsearch

import (
	"encoding/json"
	"html/template"
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/olivere/elastic/v7"
)

type Article struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	TitleMarked string  `json:"title_marked"`
	Author      string  `json:"author"`
	Body        string  `json:"body"`
	Feed        string  `json:"feed"`
	FeedName    string  `json:"feed_name"`
	AppID       int     `json:"app_id"`
	AppName     string  `json:"app_name"`
	AppIcon     string  `json:"app_icon"`
	Time        int64   `json:"time"`
	ArticleIcon string  `json:"icon"`
	Score       float64 `json:"-"`
}

func (article Article) GetBody() template.HTML {
	return helpers.GetArticleBody(article.Body)
}

func (article Article) GetArticleIcon() string {
	return helpers.GetArticleIcon(article.ArticleIcon, article.AppID, article.AppIcon)
}

func (article Article) GetAppIcon() string {
	return helpers.GetAppIcon(article.AppID, article.AppIcon)
}

func (article Article) GetAppPath() string {
	return helpers.GetAppPath(article.AppID, article.AppName) + "#news"
}

func (article Article) GetFeedName() string {
	return helpers.GetAppArticleFeedName(article.Feed, article.FeedName)
}

func (article Article) GetAppName() string {
	return helpers.GetAppName(article.AppID, article.AppName)
}

func (article Article) GetDate() string {
	return time.Unix(article.Time, 0).Format(helpers.DateYearTime)
}

func (article Article) OutputForJSON() []interface{} {

	var id = strconv.FormatInt(article.ID, 10)

	return []interface{}{
		id,                       // 0
		article.Title,            // 1
		article.GetBody(),        // 2
		article.AppID,            // 3
		article.GetArticleIcon(), // 4
		article.Time,             // 5
		article.Score,            // 6
		article.GetAppName(),     // 7
		article.GetAppPath(),     // 8
		article.GetDate(),        // 9
		article.TitleMarked,      // 10
		article.GetFeedName(),    // 11
	}
}

func IndexArticle(article Article) error {
	return indexDocument(IndexArticles, strconv.FormatInt(article.ID, 10), article)
}

func IndexArticlesBulk(articles map[string]Article) error {

	// todo, add to global
	i := map[string]interface{}{}
	for k, v := range articles {
		i[k] = v
	}

	return indexDocuments(IndexArticles, i)
}

func SearchArticles(offset int, sorters []elastic.Sorter, search string, filters []elastic.Query) (articles []Article, total int64, err error) {

	client, ctx, err := GetElastic()
	if err != nil {
		return articles, 0, err
	}

	searchService := client.Search().
		Index(IndexArticles).
		From(offset).
		Size(100).
		TrackTotalHits(true).
		SortBy(sorters...)

	b := elastic.NewBoolQuery()

	if len(filters) > 0 {
		b.Filter(filters...)
	}

	if search != "" {

		b.Must(
			elastic.NewBoolQuery().MinimumNumberShouldMatch(1).Should(
				elastic.NewMatchQuery("title", search).Boost(2),
				elastic.NewMatchQuery("app_name", search).Boost(1),
				elastic.NewMatchQuery("author", search).Boost(1),
				elastic.NewPrefixQuery("title", search).Boost(0.2),
				elastic.NewPrefixQuery("app_name", search).Boost(0.1),
			),
		).Should(
			elastic.NewFunctionScoreQuery().
				AddScoreFunc(elastic.NewGaussDecayFunction().FieldName("time").
					Origin(time.Now().Unix()). // Max
					Scale(1213743600). // Min - First news article - 2008-06-18
					Decay(0.1),
				),
		)

		searchService.Highlight(elastic.NewHighlight().Field("title").Field("app_name").PreTags("<mark>").PostTags("</mark>"))
	}

	searchService.Query(b)

	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return articles, 0, err
	}

	for _, hit := range searchResult.Hits.Hits {

		var article Article
		err := json.Unmarshal(hit.Source, &article)
		if err != nil {
			log.ErrS(err)
			continue
		}

		if hit.Score != nil {
			article.Score = *hit.Score
		}

		article.TitleMarked = article.Title
		if val, ok := hit.Highlight["title"]; ok {
			if len(val) > 0 {
				article.TitleMarked = val[0]
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

func AggregateArticleFeeds() (aggregations []helpers.TupleStringInt, err error) {

	err = memcache.GetSetInterface(memcache.MemcacheArticleFeedAggs, &aggregations, func() (interface{}, error) {

		client, ctx, err := GetElastic()
		if err != nil {
			return aggregations, err
		}

		searchService := client.Search().
			Index(IndexArticles).
			Aggregation("feed", elastic.NewTermsAggregation().Field("feed").Size(100).OrderByCountDesc())

		searchResult, err := searchService.Do(ctx)
		if err != nil {
			return aggregations, err
		}

		if a, ok := searchResult.Aggregations.Terms("feed"); ok {
			for _, feeds := range a.Buckets {
				aggregations = append(aggregations, helpers.TupleStringInt{
					Key:   feeds.Key.(string),
					Value: feeds.DocCount,
				})
			}
		}

		return aggregations, err
	})

	return aggregations, err
}

//noinspection GoUnusedExportedFunction
func DeleteAndRebuildArticlesIndex() {

	var mapping = map[string]interface{}{
		"settings": settings,
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": fieldTypeDisabled,
				"title": map[string]interface{}{
					"type":     "text",
					"analyzer": "gdb_lowercase_text",
				},
				"author": map[string]interface{}{
					"type":       "keyword",
					"normalizer": "gdb_lowercase_keyword",
				},
				"body":      fieldTypeDisabled,
				"feed":      fieldTypeKeyword,
				"feed_name": fieldTypeDisabled,
				"app_id":    fieldTypeKeyword,
				"app_name": map[string]interface{}{
					"type":     "text",
					"analyzer": "gdb_lowercase_text",
				},
				"app_icon": fieldTypeDisabled,
				"icon":     fieldTypeDisabled,
				"time":     fieldTypeInt64,
			},
		},
	}

	rebuildIndex(IndexArticles, mapping)
}
