package queue

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/steam"
	"go.mongodb.org/mongo-driver/bson"
)

type FindGroupMessage struct {
	AppID int `json:"app_id"`
}

func (m FindGroupMessage) Queue() rabbit.QueueName {
	return QueueAppsFindGroup
}

//noinspection RegExpRedundantEscape
var regexpGroupID = regexp.MustCompile(`\(\s?\'(\d{18})\'\s?\)`)

func appsFindGroupHandler(message *rabbit.Message) {

	payload := FindGroupMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err, message.Message.Body)
		sendToFailQueue(message)
		return
	}

	body, _, err := helpers.GetWithTimeout("https://steamcommunity.com/app/"+strconv.Itoa(payload.AppID), 0)
	if err != nil {
		steam.LogSteamError(err, message.Message.Body)
		sendToRetryQueue(message)
		return
	}

	var groupID string
	ret := regexpGroupID.FindAllStringSubmatch(string(body), -1)
	for _, v := range ret {
		if len(v) == 2 && strings.HasPrefix(v[1], "103") {
			groupID = v[1]
		}
	}

	if groupID == "" {
		message.Ack(false)
		return
	}

	// Update app
	filter := bson.D{
		{"_id", payload.AppID},
		{"group_id", ""},
	}

	_, err = mongo.UpdateOne(mongo.CollectionApps, filter, bson.D{{"group_id", groupID}})
	if err != nil {
		log.Err(err, message.Message.Body)
		sendToRetryQueue(message)
		return
	}

	// Clear cache
	err = memcache.Delete(memcache.MemcacheApp(payload.AppID).Key)
	if err != nil {
		log.Err(err, message.Message.Body)
		sendToRetryQueue(message)
		return
	}

	// Update in Elastic
	err = ProduceAppSearch(nil, payload.AppID)
	if err != nil {
		log.Err(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	//
	message.Ack(false)
}
