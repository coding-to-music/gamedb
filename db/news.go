package db

import (
	"net/http"
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
	Title      string    `datastore:"title,noindex"`
	URL        string    `datastore:"url,noindex"`
	IsExternal bool      `datastore:"is_external,noindex"`
	Author     string    `datastore:"author,noindex"`
	Contents   string    `datastore:"contents,noindex"`
	Date       time.Time `datastore:"date"`
	FeedLabel  string    `datastore:"feed_label,noindex"`
	FeedName   string    `datastore:"feed_name,noindex"`
	FeedType   int8      `datastore:"feed_type,noindex"`

	AppID   int    `datastore:"app_id"`
	AppName string `datastore:"app_name,noindex"`
	AppIcon string `datastore:"app_icon,noindex"`
}

func (article News) GetKey() (key *datastore.Key) {
	return datastore.NameKey(KindNews, strconv.FormatInt(article.ArticleID, 10), nil)
}

func (article News) GetBody() string {
	return helpers.BBCodeCompiler.Compile(article.Contents)
}

// Data array for datatables
func (article News) OutputForJSON(r *http.Request) (output []interface{}) {

	var id = strconv.FormatInt(article.ArticleID, 10)
	var path = GetAppPath(article.AppID, article.AppName)

	return []interface{}{
		id,                                    // 0
		article.Title,                         // 1
		article.Author,                        // 2
		article.Date.Unix(),                   // 3
		article.Date.Format(helpers.DateYear), // 4
		article.GetBody(),                     // 5
		article.AppID,                         // 6
		article.AppName,                       // 7
		article.AppIcon,                       // 8
		path + "#news," + id,                  // 9
	}
}

func CreateArticle(app App, resp steam.NewsArticle) (news News) {

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
	news.AppName = app.Name
	news.AppIcon = app.Icon

	return news
}
