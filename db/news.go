package db

import (
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/steam-go/steam"
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

func GetAppArticles(appID int) (articles []News, err error) {

	client, ctx, err := GetDSClient()
	if err != nil {
		return articles, err
	}

	q := datastore.NewQuery(KindNews).Order("-date").Limit(1000)
	q = q.Filter("app_id =", appID)

	_, err = client.GetAll(ctx, q, &articles)
	if err != nil {
		return articles, err
	}

	return articles, err
}

func GetArticles() (articles []News, err error) {

	client, ctx, err := GetDSClient()
	if err != nil {
		return articles, err
	}

	q := datastore.NewQuery(KindNews).Order("-date").Limit(100)

	_, err = client.GetAll(ctx, q, &articles)
	if err != nil {
		return articles, err
	}

	return articles, err
}

func CreateArticle(resp steam.NewsArticle) (news News) {

	news.ArticleID = resp.GID
	news.Title = resp.Title
	news.URL = resp.URL
	news.IsExternal = resp.IsExternalURL
	news.Author = resp.Author
	news.Contents = resp.Contents
	news.FeedLabel = resp.Feedlabel
	news.Date = time.Unix(int64(resp.Date), 0)
	news.FeedName = resp.Feedname
	news.FeedType = int8(resp.FeedType)
	news.AppID = resp.AppID

	return news
}
