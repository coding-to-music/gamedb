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

func GetArticlesByAppIDs(appIDs []int, limit int64, afterDate time.Time) (news []Article, err error) {

	if len(appIDs) < 1 {
		return news, nil
	}

	appsFilter := bson.A{}
	for _, v := range appIDs {
		appsFilter = append(appsFilter, v)
	}

	filter := bson.D{
		{"app_id", bson.M{"$in": appsFilter}},
	}

	if !afterDate.IsZero() && afterDate.Unix() != 0 {
		filter = append(filter, bson.E{Key: "date", Value: bson.M{"$gte": afterDate}})
	}

	return getArticles(0, limit, filter, bson.D{{"date", -1}}, nil)
}

func GetArticles(offset int64, limit int64, order bson.D, filter bson.D) (news []Article, err error) {

	return getArticles(offset, limit, filter, order, nil)
}

func getArticles(offset int64, limit int64, filter bson.D, order bson.D, projection bson.M) (news []Article, err error) {

	filter = append(filter, bson.E{Key: "feed_name", Value: bson.M{"$ne": "Gamemag.ru"}})

	cur, ctx, err := Find(CollectionAppArticles, offset, limit, order, filter, projection, nil)
	if err != nil {
		return news, err
	}

	defer closeCursor(cur, ctx)

	for cur.Next(ctx) {

		var article Article
		err := cur.Decode(&article)
		if err != nil {
			log.ErrS(err, article.ID)
		} else {
			news = append(news, article)
		}
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

func GetAppArticlesGroupedByFeed() (feeds []ArticleFeed, err error) {

	err = memcache.GetSetInterface(memcache.ItemArticleFeedAggsMongo, &feeds, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return feeds, err
		}

		pipeline := mongo.Pipeline{
			{{
				Key:   "$group",
				Value: bson.M{"_id": "$feed_name", "count": bson.M{"$sum": 1}, "name": bson.M{"$first": "$feed_label"}},
			}},
			{{
				Key:   "$match",
				Value: bson.M{"_id": bson.M{"$ne": "Gamemag.ru"}},
			}},
		}

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
