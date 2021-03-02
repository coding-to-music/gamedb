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
	"go.uber.org/zap"
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
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	body, _, err := helpers.Get("https://steamcommunity.com/app/"+strconv.Itoa(payload.AppID), 0, nil)
	if err != nil {
		steam.LogSteamError(err, zap.String("body", string(message.Message.Body)))
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
		message.Ack()
		return
	}

	// Update app
	filter := bson.D{{"_id", payload.AppID}}
	update := bson.D{{"group_id", groupID}}

	_, err = mongo.UpdateOne(mongo.CollectionApps, filter, update, nil)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	// Clear cache
	err = memcache.Delete(memcache.ItemApp(payload.AppID).Key)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	// Update in Elastic
	updateInElastic := map[string]interface{}{
		"group_id": groupID,
	}

	err = ProduceAppSearch(nil, payload.AppID, updateInElastic)
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	//
	message.Ack()
}
