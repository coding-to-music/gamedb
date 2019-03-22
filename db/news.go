package db

import (
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gamedb/website/helpers"
)

type News struct {
	ArticleID  int64     `datastore:"article_id,noindex" bson:"_id"`
	Title      string    `datastore:"title,noindex"`
	URL        string    `datastore:"url,noindex"`
	IsExternal bool      `datastore:"is_external,noindex"`
	Author     string    `datastore:"author,noindex"`
	Contents   string    `datastore:"contents,noindex"`
	Date       time.Time `datastore:"date" bson:"created_at"`
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

func (article News) GetIcon() string {

	if strings.HasPrefix(article.AppIcon, "http") {
		return article.AppIcon
	} else {
		return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(article.AppID) + "/" + article.AppIcon + ".jpg"
	}
}

// Data array for datatables
func (article News) OutputForJSON() (output []interface{}) {

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
