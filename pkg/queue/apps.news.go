package queue

import (
	"strings"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/steam"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type AppNewsMessage struct {
	AppID int `json:"id"`
}

func (m AppNewsMessage) Queue() rabbit.QueueName {
	return QueueAppsNews
}

func appNewsHandler(message *rabbit.Message) {

	payload := AppNewsMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	resp, err := steam.GetSteam().GetNews(payload.AppID, 10000)
	err = steam.AllowSteamCodes(err, 403)
	if err != nil {
		steam.LogSteamError(err)
		sendToRetryQueue(message)
		return
	}

	if len(resp.Items) == 0 {
		message.Ack()
		return
	}

	app, err := mongo.GetApp(payload.AppID, false)
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	var newsIDsMap = map[int64]bool{}
	for _, v := range app.NewsIDs {
		newsIDsMap[v] = true
	}

	var articles []mongo.Article
	var newsIDs []int64

	for _, v := range resp.Items {

		if strings.TrimSpace(v.Contents) == "" {
			continue
		}

		if _, ok := newsIDsMap[int64(v.GID)]; ok {
			continue
		}

		news := mongo.Article{}
		news.ID = int64(v.GID)
		news.Title = v.Title
		news.URL = v.URL
		news.IsExternal = v.IsExternalURL
		news.Author = v.Author
		news.Contents = v.Contents
		news.FeedLabel = v.Feedlabel
		news.Date = time.Unix(v.Date, 0)
		news.FeedName = v.Feedname
		news.FeedType = int8(v.FeedType)
		news.ArticleIcon = helpers.FindArticleImage(v.Contents)

		news.AppID = v.AppID
		news.AppName = app.Name
		news.AppIcon = app.Icon

		articles = append(articles, news)
		newsIDs = append(newsIDs, int64(v.GID))

		m := AppsArticlesSearchMessage{
			ID:          int64(v.GID),
			Title:       v.Title,
			Body:        v.Contents,
			Time:        v.Date,
			AppID:       app.ID,
			AppName:     app.Name,
			AppIcon:     app.Icon,
			ArticleIcon: news.ArticleIcon,
		}

		err = ProduceArticlesSearch(m)
		if err != nil {
			log.ErrS(err)
		}
	}

	err = mongo.ReplaceArticles(articles)
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	// Update app row
	newsIDs = helpers.UniqueInt64(newsIDs)

	_, err = mongo.UpdateOne(mongo.CollectionApps, bson.D{{"_id", app.ID}}, bson.D{{"news_ids", newsIDs}})
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	// Clear cache
	err = memcache.Delete(memcache.MemcacheApp(app.ID).Key)
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	// Update in Elastic
	err = ProduceAppSearch(nil, payload.AppID)
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	message.Ack()
}
