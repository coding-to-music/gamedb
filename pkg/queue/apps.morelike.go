package queue

import (
	"regexp"
	"strconv"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	steamHelper "github.com/gamedb/gamedb/pkg/steam"
	"github.com/gocolly/colly"
	"go.mongodb.org/mongo-driver/bson"
)

type AppMorelikeMessage struct {
	AppID int `json:"id"`
}

func (m AppMorelikeMessage) Queue() rabbit.QueueName {
	return QueueAppsMorelike
}

func appMorelikeHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := AppMorelikeMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		c := colly.NewCollector(
			colly.URLFilters(regexp.MustCompile(`store\.steampowered\.com/recommended/morelike/app/[0-9]+$`)),
			steamHelper.WithAgeCheckCookie,
			steamHelper.WithTimeout,
		)

		var relatedAppIDs []int

		c.OnHTML(".similar_grid_capsule", func(e *colly.HTMLElement) {
			i, err := strconv.Atoi(e.Attr("data-ds-appid"))
			if err == nil {
				relatedAppIDs = append(relatedAppIDs, i)
			}
		})

		err = c.Visit("https://store.steampowered.com/recommended/morelike/app/" + strconv.Itoa(payload.AppID))
		if err != nil {
			steamHelper.LogSteamError(err)
			sendToRetryQueue(message)
			continue
		}

		if len(relatedAppIDs) == 0 {
			message.Ack(false)
			continue
		}

		_, err = mongo.UpdateOne(mongo.CollectionApps, bson.D{{"_id", payload.AppID}}, bson.D{{"related_app_ids", relatedAppIDs}})
		if err != nil {
			log.Err(err, payload.AppID)
			sendToRetryQueue(message)
			continue
		}

		err = memcache.Delete(memcache.MemcacheApp(payload.AppID).Key)
		if err != nil {
			log.Err(err, payload.AppID)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
}
