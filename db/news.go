package db

import (
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gamedb/website/helpers"
)

type News struct {
	CreatedAt  time.Time `datastore:"created_at,noindex"`
	UpdatedAt  time.Time `datastore:"updated_at,noindex"`
	ArticleID  int64     `datastore:"article_id,noindex"`
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

func (article News) GetKey() (key *datastore.Key) {
	return datastore.NameKey(KindNews, strconv.FormatInt(article.ArticleID, 10), nil)
}

func (article News) GetTimestamp() (int64) {
	return article.Date.Unix()
}

func (article News) GetNiceDate() (string) {
	return article.Date.Format(helpers.DateYear)
}

func GetArticles(appID int, limit int) (articles []News, err error) {

	client, ctx, err := GetDSClient()
	if err != nil {
		return articles, err
	}

	q := datastore.NewQuery(KindNews).Order("-date").Limit(limit)

	if appID != 0 {
		q = q.Filter("app_id =", appID)
	}

	_, err = client.GetAll(ctx, q, &articles)
	if err != nil {
		return
	}

	return articles, err
}

func GetNewArticles(appID int) (news []*News, err error) {

	// Get latest article from database
	var latestTime int64

	latest, err := GetArticles(appID, 1)
	if err != nil {
		return news, err
	}

	if len(latest) > 0 {
		latestTime = latest[0].Date.Unix()
	}

	// Get app articles from Steam
	resp, _, err := helpers.GetSteam().GetNews(appID)
	if err != nil {
		return news, err
	}

	for _, v := range resp.Items {

		if v.Date > latestTime {

			article := new(News)
			article.ArticleID = v.GID
			article.Title = v.Title
			article.URL = v.URL
			article.IsExternal = v.IsExternalURL
			article.Author = v.Author
			article.Contents = v.Contents
			article.FeedLabel = v.Feedlabel
			article.Date = time.Unix(int64(v.Date), 0)
			article.FeedName = v.Feedname
			article.FeedType = int8(v.FeedType)
			article.AppID = v.AppID

			news = append(news, article)
		}
	}

	return news, err
}
