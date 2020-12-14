package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.uber.org/zap"
)

type AppsArticlesSearchMessage struct {
	Elastic elasticsearch.Article `json:"elastic"`
}

func appsArticlesSearchHandler(message *rabbit.Message) {

	payload := AppsArticlesSearchMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	if payload.Elastic.ArticleIcon == "" {
		payload.Elastic.ArticleIcon = payload.Elastic.GetArticleIcon()
	}

	err = elasticsearch.IndexArticle(payload.Elastic)
	if err != nil {
		log.ErrS(err)
		sendToRetryQueue(message)
		return
	}

	//
	message.Ack()
}
