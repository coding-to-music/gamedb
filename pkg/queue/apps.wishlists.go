package queue

import (
	"strconv"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	influx "github.com/influxdata/influxdb1-client"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type AppWishlistsMessage struct {
	AppID int `json:"id"`
}

func appWishlistsHandler(message *rabbit.Message) {

	payload := AppWishlistsMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	playerWishlists, err := mongo.GetPlayerWishlistAppsByApp(payload.AppID)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	var (
		wishlistCount   = len(playerWishlists)
		wishlistAverage float64
		wishlistPercent float64
		wishlistFirsts  float64
	)

	//
	var total int
	var count int
	for _, v := range playerWishlists {

		if v.Order == 1 {
			wishlistFirsts++
		}

		if v.Order > 0 {
			total += v.Order
			count++
		}
	}

	if count == 0 {
		wishlistAverage = 0
	} else {
		wishlistAverage = float64(total) / float64(count)
	}

	// Get percent of players
	wishlistPlayers, err := mongo.CountDocuments(mongo.CollectionPlayers, nil, 60*60)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	if wishlistPlayers == 0 {
		wishlistPercent = 0
	} else {
		wishlistPercent = float64(wishlistCount) / float64(wishlistPlayers)
	}

	// Save to influx
	var point = influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementApps),
		Tags: map[string]string{
			"app_id": strconv.Itoa(payload.AppID),
		},
		Fields: map[string]interface{}{
			"wishlist_count":        wishlistCount,
			"wishlist_avg_position": wishlistAverage,
			"wishlist_percent":      wishlistPercent,
			"wishlist_firsts":       wishlistFirsts,
		},
		Time:      time.Now(),
		Precision: "m",
	}

	_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, point)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	// Save to Mongo
	var update = bson.D{
		{"wishlist_count", wishlistCount},
		{"wishlist_avg_position", wishlistAverage},
		{"wishlist_percent", wishlistPercent},
		{"wishlist_firsts", wishlistFirsts},
	}

	_, err = mongo.UpdateOne(mongo.CollectionApps, bson.D{{"_id", payload.AppID}}, update)
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	// Clear app memcache
	err = memcache.Delete(memcache.ItemApp(payload.AppID).Key)
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	// Update in Elastic
	updateInElastic := map[string]interface{}{
		"wishlist_count": wishlistCount,
		"wishlist_avg":   wishlistAverage,
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
