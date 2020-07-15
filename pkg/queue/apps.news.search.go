package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
)

type AppsArticlesSearchMessage struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	Time        int64  `json:"time"`
	AppID       int    `json:"app_id"`
	AppName     string `json:"app_name"`
	AppIcon     string `json:"app_icon"`
	ArticleIcon string `json:"icon"`
}

func appsArticlesSearchHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := AppsArticlesSearchMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		article := elasticsearch.Article{}
		article.ID = payload.ID
		article.Title = payload.Title
		article.Body = payload.Body
		article.Time = payload.Time
		article.AppID = payload.AppID
		article.AppName = payload.AppName
		article.AppIcon = payload.AppIcon
		article.ArticleIcon = payload.ArticleIcon

		err = elasticsearch.IndexArticle(article)
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
}
