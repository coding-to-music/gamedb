package datastore

import (
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/steam"
)

type Article struct {
	CreatedAt  time.Time `datastore:"created_at,noindex"`
	UpdatedAt  time.Time `datastore:"updated_at,noindex"`
	ArticleID  int       `datastore:"article_id,noindex"`
	AppID      int       `datastore:"app_id"`
	Title      string    `datastore:"title,noindex"`
	URL        string    `datastore:"url,noindex"`
	IsExternal bool      `datastore:"is_external,noindex"`
	Author     string    `datastore:"author,noindex"`
	Contents   string    `datastore:"contents,noindex"`
	Date       time.Time `datastore:"date"`
	FeedLabel  string    `datastore:"feed_label,noindex"`
	FeedName   string    `datastore:"feed_name,noindex"`
	FeedType   int8      `datastore:"feed_type,noindex"`
}

func (article Article) GetKey() (key *datastore.Key) {
	return datastore.NameKey(KindArticle, strconv.Itoa(article.ArticleID), nil)
}

func (article Article) GetTimestamp() (int64) {
	return article.Date.Unix()
}

func (article Article) GetNiceDate() (string) {
	return article.Date.Format(helpers.DateYear)
}

func (article *Article) Tidy() *Article {

	article.UpdatedAt = time.Now()
	if article.CreatedAt.IsZero() {
		article.CreatedAt = time.Now()
	}

	return article
}

func GetArticles(appID int, limit int) (articles []Article, err error) {

	client, ctx, err := getClient()
	if err != nil {
		return articles, err
	}

	q := datastore.NewQuery(KindArticle).Order("-date").Limit(limit)

	if appID != 0 {
		q = q.Filter("app_id =", appID)
	}

	client.GetAll(ctx, q, &articles)

	return articles, err
}

func GetNewArticles(appID int) (articles []*Article, err error) {

	// Get latest article from database
	var latestTime int64

	latest, err := GetArticles(appID, 1)
	if err != nil {
		logger.Error(err)
	}

	if len(latest) > 0 {
		latestTime = latest[0].Date.Unix()
	}

	// Get app articles from Steam
	resp, err := steam.GetNewsForApp(strconv.Itoa(appID))

	var articlePointers []*Article
	for _, v := range resp {

		if v.Date > latestTime {

			articleID, err := strconv.Atoi(v.GID)
			if err != nil {
				logger.Error(err)
			}

			article := new(Article)
			article.ArticleID = articleID
			article.Title = v.Title
			article.URL = v.URL
			article.IsExternal = v.IsExternalURL
			article.Author = v.Author
			article.Contents = v.Contents
			article.FeedLabel = v.Feedlabel
			article.Date = time.Unix(int64(v.Date), 0)
			article.FeedName = v.Feedname
			article.FeedType = int8(v.FeedType)
			article.AppID = v.Appid

			articlePointers = append(articlePointers, article)
		}
	}

	err = bulkAddArticles(articlePointers)
	if err != nil {
		return articles, err
	}

	return articles, nil
}

func bulkAddArticles(articles []*Article) (err error) {

	articlesLen := len(articles)
	if articlesLen == 0 {
		return nil
	}

	client, context, err := getClient()
	if err != nil {
		return err
	}

	keys := make([]*datastore.Key, 0, articlesLen)

	for _, v := range articles {
		keys = append(keys, v.GetKey())
	}

	_, err = client.PutMulti(context, keys, articles)
	if err != nil {
		return err
	}

	return nil
}
