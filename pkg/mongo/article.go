package mongo

import (
	"html/template"
	"strconv"
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
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

func (article Article) GetIcon() string {

	if strings.HasPrefix(article.AppIcon, "http") || strings.HasPrefix(article.AppIcon, "/") {
		return strings.TrimPrefix(article.AppIcon, "https://gamedb.online")
	} else if article.AppIcon != "" {
		return helpers.AppIconBase + strconv.Itoa(article.AppID) + "/" + article.AppIcon + ".jpg"
	} else {
		return helpers.DefaultAppIcon
	}
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

func GetArticlesByApp(appID int, offset int64) (news []Article, err error) {

	return getArticles(offset, 100, bson.D{{"app_id", appID}}, bson.D{{"date", -1}}, nil)
}

func GetArticles(offset int64, limit int64, order bson.D) (news []Article, err error) {

	return getArticles(offset, limit, nil, order, nil)
}

func getArticles(offset int64, limit int64, filter bson.D, order bson.D, projection bson.M) (news []Article, err error) {

	cur, ctx, err := Find(CollectionAppArticles, offset, limit, order, filter, projection, nil)
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
		} else {
			news = append(news, article)
		}
	}

	return news, cur.Err()
}
