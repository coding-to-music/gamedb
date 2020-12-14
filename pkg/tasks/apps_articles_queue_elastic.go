package tasks

import (
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
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

func (c AppsArticlesQueueElastic) Group() TaskGroup {
	return TaskGroupElastic
}

func (c AppsArticlesQueueElastic) Cron() TaskTime {
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
				Elastic: elasticsearch.Article{
					ID:          article.ID,
					Title:       article.Title,
					Author:      article.Author,
					Body:        article.Contents,
					Feed:        article.FeedName,
					FeedName:    article.FeedLabel,
					Time:        article.Date.Unix(),
					AppID:       article.AppID,
					AppName:     article.AppName,
					AppIcon:     article.AppIcon,
					ArticleIcon: helpers.FindArticleImage(article.Contents),
					// TitleMarked: "",
					// Score:       0,
				},
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
