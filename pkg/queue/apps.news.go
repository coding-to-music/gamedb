package queue

import (
	"strings"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	steamHelper "github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type AppNewsMessage struct {
	ID int `json:"id"`
}

func appNewsHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := AppNewsMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		resp, _, err := steamHelper.GetSteam().GetNews(payload.ID, 10000)
		err = steamHelper.AllowSteamCodes(err, 403)
		if err != nil {
			steamHelper.LogSteamError(err)
			sendToRetryQueue(message)
			continue
		}

		if len(resp.Items) == 0 {
			message.Ack(false)
			continue
		}

		app, err := mongo.GetApp(payload.ID, false)
		if err != nil {
			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			continue
		}

		var newsIDsMap = map[int64]bool{}
		for _, v := range app.NewsIDs {
			newsIDsMap[v] = true
		}

		var documents []mongo.Document
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
			news.AppName = app.GetName()
			news.AppIcon = app.GetIcon()

			documents = append(documents, news)
			newsIDs = append(newsIDs, int64(v.GID))

			err = ProduceArticlesSearch(AppsArticlesSearchMessage{
				ID:          int64(v.GID),
				Title:       v.Title,
				Body:        v.Contents,
				Time:        v.Date,
				AppID:       app.ID,
				AppName:     app.Name,
				AppIcon:     app.Icon,
				ArticleIcon: news.ArticleIcon,
			})
			log.Err(err)
		}

		_, err = mongo.InsertMany(mongo.CollectionAppArticles, documents)
		if err != nil {
			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			continue
		}

		newsIDs = helpers.UniqueInt64(newsIDs)

		_, err = mongo.UpdateOne(mongo.CollectionApps, bson.D{{"_id", app.ID}}, bson.D{{"news_ids", newsIDs}})
		if err != nil {
			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			continue
		}

		err = memcache.Delete(memcache.MemcacheApp(app.ID).Key)
		if err != nil {
			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
}
