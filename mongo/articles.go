package mongo

import (
	"time"

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
	Date       time.Time `bson:"created_at"`
	FeedLabel  string    ``
	FeedName   string    ``
	FeedType   int8      ``
	AppID      int       ``
	AppName    string    ``
	AppIcon    string    ``
}

func (article Article) Key() interface{} {
	return article.ID
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
		"app_id":      article.AppID,
		"app_name":    article.AppName,
		"app_icon":    article.AppIcon,
	}
}

func GetArticles(appID int, offset int64) (news []Article, err error) {

	client, ctx, err := GetMongo()
	if err != nil {
		return news, err
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionEvents)

	cur, err := c.Find(ctx, bson.M{"app_id": appID}, options.Find().SetLimit(100).SetSkip(offset).SetSort(bson.M{"created_at": -1}))
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
		log.Err(err, article.ID)
		news = append(news, article)
	}

	return news, cur.Err()
}
