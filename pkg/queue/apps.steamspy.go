package queue

import (
	"errors"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/powerslacker/ratelimit"
	"go.mongodb.org/mongo-driver/bson"
)

type AppSteamspyMessage struct {
	ID int `json:"id"`
}

// https://steamspy.com/api.php
var steamspyLimiter = ratelimit.New(4, ratelimit.WithoutSlack)

func appSteamspyHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := AppSteamspyMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		// Create request
		query := url.Values{}
		query.Set("request", "appdetails")
		query.Set("appid", strconv.Itoa(payload.ID))

		steamspyLimiter.Take()
		response, err := helpers.GetWithTimeout("https://steamspy.com/api.php?"+query.Encode(), 0)
		if err != nil {

			if strings.Contains(err.Error(), "Client.Timeout exceeded while awaiting headers") ||
				strings.Contains(err.Error(), "read: connection reset by peer") {
				log.Info(err, payload.ID)
			} else {
				log.Err(err, payload.ID)
			}

			sendToRetryQueueWithDelay(message, time.Second*10)
			continue
		}

		//noinspection GoDeferInLoop
		defer func() {
			err = response.Body.Close()
			if err != nil {
				log.Err(err, payload.ID)
			}
		}()

		if response.StatusCode != 200 {

			log.Info(errors.New("steamspy is down"), payload.ID)
			sendToRetryQueueWithDelay(message, time.Minute*30)
			continue
		}

		bytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			continue
		}

		if strings.Contains(string(bytes), "Connection failed") {

			log.Info(errors.New("steamspy is down"), payload.ID, bytes)
			sendToRetryQueueWithDelay(message, time.Minute*30)
			continue
		}

		// Unmarshal JSON
		resp := mongo.SteamSpyAppResponse{}
		err = helpers.Unmarshal(bytes, &resp)
		if err != nil {

			log.Info(err, payload.ID, bytes)
			sendToRetryQueueWithDelay(message, time.Minute*30)
			continue
		}

		ss := helpers.AppSteamSpy{}
		ss.SSAveragePlaytimeTwoWeeks = resp.Average2Weeks
		ss.SSAveragePlaytimeForever = resp.AverageForever
		ss.SSMedianPlaytimeTwoWeeks = resp.Median2Weeks
		ss.SSMedianPlaytimeForever = resp.MedianForever

		owners := resp.GetOwners()
		if len(owners) == 2 {
			ss.SSOwnersLow = owners[0]
			ss.SSOwnersHigh = owners[1]
		}

		_, err = mongo.UpdateOne(mongo.CollectionApps, bson.D{{"_id", payload.ID}}, bson.D{{"steam_spy", ss}})
		if err != nil {
			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			continue
		}

		err = memcache.Delete(memcache.MemcacheApp(payload.ID).Key)
		if err != nil {
			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
}
