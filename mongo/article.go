package mongo

import (
	"strconv"
	"strings"
	"time"

	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Article struct {
	ID         int64     `bson:"_id"`
	Title      string    `bson:"title"`
	URL        string    `bson:"url"`
	IsExternal bool      `bson:"is_external"`
	Author     string    `bson:"author"`
	Contents   string    `bson:"contents"`
	Date       time.Time `bson:"date"`
	FeedLabel  string    `bson:"feed_label"`
	FeedName   string    `bson:"feed_name"`
	FeedType   int8      `bson:"feed_type"`
	AppID      int       `bson:"app_id"`
	AppName    string    `bson:"app_name"`
	AppIcon    string    `bson:"app_icon"`
}

func (article Article) BSON() (ret interface{}) {

	return M{
		"_id":         article.ID,
		"title":       article.Title,
		"url":         article.URL,
		"is_external": article.IsExternal,
		"author":      article.Author,
		"contents":    article.Contents,
		"date":        article.Date,
		"feed_label":  article.FeedLabel,
		"feed_name":   article.FeedName,
		"feed_type":   article.FeedType,

		"app_id":   article.AppID,
		"app_name": article.AppName,
		"app_icon": article.AppIcon,
	}
}

func (article Article) GetBody() string {
	return helpers.BBCodeCompiler.Compile(article.Contents)
}

func (article Article) GetDate() string {
	return article.Date.Format(helpers.Date)
}

func (article Article) GetIcon() string {

	if strings.HasPrefix(article.AppIcon, "http") || strings.HasPrefix(article.AppIcon, "/") {
		return article.AppIcon
	} else if article.AppIcon != "" {
		return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(article.AppID) + "/" + article.AppIcon + ".jpg"
	} else {
		return helpers.DefaultAppIcon
	}
}

func (article Article) GetAppPath() string {

	return helpers.GetAppPath(article.AppID, article.AppName)
}

func (article Article) OutputForJSON() (output []interface{}) {

	var id = strconv.FormatInt(article.ID, 10)
	var path = helpers.GetAppPath(article.AppID, article.AppName)

	return []interface{}{
		id,                                    // 0
		article.Title,                         // 1
		article.Author,                        // 2
		article.Date.Unix(),                   // 3
		article.Date.Format(helpers.DateYear), // 4
		article.GetBody(),                     // 5
		article.AppID,                         // 6
		article.AppName,                       // 7
		article.GetIcon(),                     // 8
		path + "#news," + id,                  // 9
	}
}

func GetArticlesByApps(appIDs []int, afterDate time.Time) (news []Article, err error) {

	if len(appIDs) < 1 {
		return news, nil
	}

	appsFilter := A{}
	for _, v := range appIDs {
		appsFilter = append(appsFilter, v)
	}

	return getArticles(0, 0, M{"app_id": M{"$in": appsFilter}, "date": M{"$gte": afterDate}})
}

func GetArticlesByApp(appID int, offset int64) (news []Article, err error) {

	return getArticles(offset, 100, M{"app_id": appID})
}

func GetArticles(offset int64) (news []Article, err error) {

	return getArticles(offset, 100, nil)
}

func getArticles(offset int64, limit int64, filter interface{}) (news []Article, err error) {

	if filter == nil {
		filter = M{}
	}

	client, ctx, err := getMongo()
	if err != nil {
		return news, err
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionAppArticles.String())

	cur, err := c.Find(ctx, filter, options.Find().SetLimit(limit).SetSkip(offset).SetSort(M{"_id": -1}))
	if err != nil {
		return news, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var article Article
		err := cur.Decode(&article)
		if err != nil {
			log.Err(err, article.ID)
		}
		news = append(news, article)
	}

	return news, cur.Err()
}
