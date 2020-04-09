package queue

import (
	"regexp"
	"strconv"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	steamHelper "github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gocolly/colly"
	"go.mongodb.org/mongo-driver/bson"
)

type AppMorelikeMessage struct {
	ID int `json:"id"`
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

		var relatedAppIDs []int

		c := colly.NewCollector(
			colly.URLFilters(regexp.MustCompile(`store\.steampowered\.com/recommended/morelike/app/[0-9]+$`)),
			steamHelper.WithAgeCheckCookie,
		)

		c.OnHTML(".similar_grid_capsule", func(e *colly.HTMLElement) {
			i, err := strconv.Atoi(e.Attr("data-ds-appid"))
			if err == nil {
				relatedAppIDs = append(relatedAppIDs, i)
			}
		})

		err = c.Visit("https://store.steampowered.com/recommended/morelike/app/" + strconv.Itoa(payload.ID))
		if err != nil {
			steamHelper.LogSteamError(err)
			sendToRetryQueue(message)
			continue
		}

		if len(relatedAppIDs) == 0 {
			log.Warning("no similar apps", payload.ID)
			message.Ack(false)
			continue
		}

		_, err = mongo.UpdateOne(mongo.CollectionApps, bson.D{{"_id", payload.ID}}, bson.D{{"related_app_ids", relatedAppIDs}})
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			continue
		}

		err = memcache.Delete(memcache.MemcacheApp(payload.ID).Key)
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
}
