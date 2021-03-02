package mongo

import (
	"html/template"
	"sort"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Article struct {
	ID          int64     `bson:"_id"`
	Title       string    `bson:"title"`
	URL         string    `bson:"url"`
	IsExternal  bool      `bson:"is_external"`
	Author      string    `bson:"author"`
	Contents    string    `bson:"contents"`
	Date        time.Time `bson:"date"`
	FeedLabel   string    `bson:"feed_label"`
	FeedName    string    `bson:"feed_name"`
	FeedType    int8      `bson:"feed_type"`
	AppID       int       `bson:"app_id"`
	AppName     string    `bson:"app_name"`
	AppIcon     string    `bson:"app_icon"`
	ArticleIcon string    `bson:"icon"`
}

func (article Article) BSON() bson.D {

	return bson.D{
		{"_id", article.ID},
		{"title", article.Title},
		{"url", article.URL},
		{"is_external", article.IsExternal},
		{"author", article.Author},
		{"contents", article.Contents},
		{"date", article.Date},
		{"feed_label", article.FeedLabel},
		{"feed_name", article.FeedName},
		{"feed_type", article.FeedType},
		{"icon", article.ArticleIcon},

		{"app_id", article.AppID},
		{"app_name", article.AppName},
		{"app_icon", article.AppIcon},
	}
}

func (article Article) GetBody() template.HTML {
	return helpers.GetArticleBody(article.Contents)
}

func (article Article) GetDate() string {
	return article.Date.Format(helpers.Date)
}

func (article Article) GetAppIcon() string {
	return helpers.GetAppIcon(article.AppID, article.AppIcon)
}

func (article Article) GetArticleIcon() string {
	return helpers.GetArticleIcon(article.ArticleIcon, article.AppID, article.AppIcon)
}

func (article Article) GetAppPath() string {
	return helpers.GetAppPath(article.AppID, article.AppName)
}

func GetArticles(offset int64, limit int64, order bson.D, filter bson.D) (news []Article, err error) {

	return getArticles(offset, limit, filter, order, nil)
}

func getArticles(offset int64, limit int64, filter bson.D, order bson.D, projection bson.M) (news []Article, err error) {

	filter = append(filter, bson.E{Key: "feed_name", Value: bson.M{"$ne": "Gamemag.ru"}})

	cur, ctx, err := find(CollectionAppArticles, offset, limit, filter, order, projection, nil)
	if err != nil {
		return news, err
	}

	defer closeCursor(cur, ctx)

	for cur.Next(ctx) {

		var article Article
		err := cur.Decode(&article)
		if err != nil {
			log.ErrS(err, article.ID)
			continue
		}

		news = append(news, article)
	}

	return news, cur.Err()
}

func ReplaceArticles(articles []Article) (err error) {

	if len(articles) == 0 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	var writes []mongo.WriteModel
	for _, article := range articles {

		write := mongo.NewReplaceOneModel()
		write.SetFilter(bson.M{"_id": article.ID})
		write.SetReplacement(article.BSON())
		write.SetUpsert(true)

		writes = append(writes, write)
	}

	c := client.Database(config.C.MongoDatabase).Collection(CollectionAppArticles.String())

	_, err = c.BulkWrite(ctx, writes, options.BulkWrite())

	return err
}

type ArticleFeed struct {
	ID    string `bson:"_id" json:"id"`
	Name  string `bson:"name" json:"name"`
	Count int    `bson:"count" json:"count"`
}

func (af ArticleFeed) GetCount() string {
	return humanize.Comma(int64(af.Count))
}

func (af ArticleFeed) GetName() string {
	return helpers.GetAppArticleFeedName(af.ID, af.Name)
}

func GetAppArticlesGroupedByFeed(appID int) (feeds []ArticleFeed, err error) {

	err = memcache.GetSetInterface(memcache.ItemArticleFeedAggsMongo(appID), &feeds, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return feeds, err
		}

		pipeline := mongo.Pipeline{}

		if appID > 0 {
			pipeline = append(pipeline, bson.D{{
				Key:   "$match",
				Value: bson.M{"app_id": appID},
			}})
		}

		pipeline = append(pipeline,
			bson.D{{
				Key:   "$group",
				Value: bson.M{"_id": "$feed_name", "count": bson.M{"$sum": 1}, "name": bson.M{"$first": "$feed_label"}},
			}},
			bson.D{{
				Key:   "$match",
				Value: bson.M{"_id": bson.M{"$ne": "Gamemag.ru"}},
			}},
		)

		cur, err := client.Database(config.C.MongoDatabase, options.Database()).Collection(CollectionAppArticles.String()).Aggregate(ctx, pipeline)
		if err != nil {
			return feeds, err
		}

		defer closeCursor(cur, ctx)

		var feeds []ArticleFeed
		for cur.Next(ctx) {

			var feed ArticleFeed
			err := cur.Decode(&feed)
			if err != nil {
				log.ErrS(err, feed.ID)
			}

			feeds = append(feeds, feed)
		}

		sort.Slice(feeds, func(i, j int) bool {
			return feeds[i].Count > feeds[j].Count
		})

		return feeds, cur.Err()
	})

	return feeds, err
}
