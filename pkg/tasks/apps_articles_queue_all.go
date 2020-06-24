package tasks

import (
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsArticlesQueueElastic struct {
	BaseTask
}

func (c AppsArticlesQueueElastic) ID() string {
	return "apps-articles-queue-elastic"
}

func (c AppsArticlesQueueElastic) Name() string {
	return "Queue all app articles to Elastic"
}

func (c AppsArticlesQueueElastic) Cron() string {
	return ""
}

func (c AppsArticlesQueueElastic) work() (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		articles, err := mongo.GetArticles(offset, limit, bson.D{{"_id", 1}}, nil)
		if err != nil {
			return err
		}

		for _, article := range articles {

			err = queue.ProduceArticlesSearch(queue.AppsArticlesSearchMessage{
				ID:      article.ID,
				Title:   article.Title,
				Body:    article.Contents,
				Time:    article.Date.Unix(),
				AppID:   article.AppID,
				AppName: article.AppName,
				AppIcon: article.AppIcon,
			})
			if err != nil {
				return err
			}
		}

		if int64(len(articles)) != limit {
			break
		}

		offset += limit
	}

	return nil
}
