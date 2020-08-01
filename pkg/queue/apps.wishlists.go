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
)

type AppWishlistsMessage struct {
	ID int `json:"id"`
}

func appWishlistsHandler(message *rabbit.Message) {

	payload := AppWishlistsMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err, message.Message.Body)
		sendToFailQueue(message)
		return
	}

	playerWishlists, err := mongo.GetPlayerWishlistAppsByApp(payload.ID)
	if err != nil {
		log.Err(err, message.Message.Body)
		sendToRetryQueue(message)
		return
	}

	var wishlistCount = len(playerWishlists)
	var wishlistAverage float64
	var wishlistPercent float64

	//
	var total int
	var count int
	for _, v := range playerWishlists {
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
		log.Err(err, message.Message.Body)
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
			"app_id": strconv.Itoa(payload.ID),
		},
		Fields: map[string]interface{}{
			"wishlist_count":        wishlistCount,
			"wishlist_avg_position": wishlistAverage,
			"wishlist_percent":      wishlistPercent,
		},
		Time:      time.Now(),
		Precision: "m",
	}

	_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, point)
	if err != nil {
		log.Err(err, message.Message.Body)
		sendToRetryQueue(message)
		return
	}

	// Save to Mongo
	var update = bson.D{
		{"wishlist_count", wishlistCount},
		{"wishlist_avg_position", wishlistAverage},
		{"wishlist_percent", wishlistPercent},
	}

	_, err = mongo.UpdateOne(mongo.CollectionApps, bson.D{{"_id", payload.ID}}, update)
	if err != nil {
		log.Err(err, payload.ID)
		sendToRetryQueue(message)
		return
	}

	// Clear app memcache
	err = memcache.Delete(memcache.MemcacheApp(payload.ID).Key)
	if err != nil {
		log.Err(err, payload.ID)
		sendToRetryQueue(message)
		return
	}

	//
	message.Ack(false)
}
