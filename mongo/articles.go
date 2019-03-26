package mongo

import (
	"strconv"
	"strings"
	"time"

	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Article struct {
	ID         int64     `bson:"_id"`
	Title      string    ``
	URL        string    ``
	IsExternal bool      ``
	Author     string    ``
	Contents   string    ``
	Date       time.Time ``
	FeedLabel  string    ``
	FeedName   string    ``
	FeedType   int8      ``
	AppID      int       ``
	AppName    string    ``
	AppIcon    string    ``
}

func (article Article) BSON() (ret interface{}) {

	return bson.M{
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

func (article Article) GetIcon() string {

	if strings.HasPrefix(article.AppIcon, "http") {
		return article.AppIcon
	} else {
		return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(article.AppID) + "/" + article.AppIcon + ".jpg"
	}
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
		article.AppIcon,                       // 8
		path + "#news," + id,                  // 9
	}
}

func GetArticlesByAppIDs(appIDs []int) (news []Article, err error) {

	appsFilter := bson.A{}
	for _, v := range appIDs {
		appsFilter = append(appsFilter, v)
	}

	return getArticles(0, 3, bson.M{"app_id": bson.M{"$in": appsFilter}})
}

func GetArticlesByAppID(appID int, offset int64) (news []Article, err error) {

	return getArticles(offset, 100, bson.M{"app_id": appID})
}

func GetArticles(offset int64) (news []Article, err error) {

	return getArticles(offset, 100, bson.M{})
}

func getArticles(offset int64, limit int64, filter interface{}) (news []Article, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return news, err
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionEvents.String())

	cur, err := c.Find(ctx, filter, options.Find().SetLimit(limit).SetSkip(offset).SetSort(bson.M{"_id": -1}))
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
