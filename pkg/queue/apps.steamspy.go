package queue

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/rate-limit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type AppSteamspyMessage struct {
	AppID int `json:"id"`
}

func (m AppSteamspyMessage) Queue() rabbit.QueueName {
	return QueueAppsSteamspy
}

// https://steamspy.com/api.php
var (
	steamspyLimiterGlobal = rate.New(time.Second * 2)
	steamspyLimiterApp    = rate.New(time.Hour * 2)
)

func appSteamspyHandler(message *rabbit.Message) {

	attempt := time.Duration(message.Attempt())

	payload := AppSteamspyMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	// Rate limiters
	err = steamspyLimiterGlobal.GetLimiter("global").Wait(context.TODO())
	if err != nil {
		log.ErrS(err)
		sendToRetryQueue(message)
		return
	}
	if !steamspyLimiterApp.GetLimiter(fmt.Sprint(payload.AppID)).Allow() {
		message.Ack()
		return
	}

	// Create request
	query := url.Values{}
	query.Set("request", "appdetails")
	query.Set("appid", strconv.Itoa(payload.AppID))

	u := "https://steamspy.com/api.php?" + query.Encode()

	body, statusCode, err := helpers.Get(u, time.Second*30, nil)
	if err != nil {

		if strings.Contains(err.Error(), "Client.Timeout exceeded while awaiting headers") ||
			strings.Contains(err.Error(), "read: connection reset by peer") ||
			strings.Contains(err.Error(), "connect: cannot assign requested address") ||
			strings.Contains(err.Error(), "server responded with error") {

			log.Info(err.Error(), zap.Int("app", payload.AppID))
		} else {
			log.Err(err.Error(), zap.Int("app", payload.AppID), zap.String("url", u))
		}

		sendToRetryQueueWithDelay(message, time.Minute*attempt)
		return
	}

	if statusCode != 200 {

		log.InfoS("steamspy is down", payload.AppID)
		sendToRetryQueueWithDelay(message, time.Minute*30*attempt)
		return
	}

	if strings.Contains(string(body), "Connection failed") {

		log.Info("steamspy is down", zap.Int("app", payload.AppID), zap.String("url", u), zap.String("body", string(body)))
		sendToRetryQueueWithDelay(message, time.Minute*30*attempt)
		return
	}

	// Unmarshal JSON
	resp := steamSpyAppResponse{}
	err = helpers.Unmarshal(body, &resp)
	if err != nil {

		log.InfoS(err, payload.AppID, helpers.TruncateString(string(body), 200, "..."))
		sendToRetryQueueWithDelay(message, time.Minute*30*attempt)
		return
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

	// Update app row
	filter := bson.D{{"_id", payload.AppID}}
	update := bson.D{{"steam_spy", ss}}

	_, err = mongo.UpdateOne(mongo.CollectionApps, filter, update)
	if err != nil {
		log.ErrS(err, payload.AppID, u)
		sendToRetryQueue(message)
		return
	}

	// Clear cache
	err = memcache.Delete(memcache.ItemApp(payload.AppID).Key)
	if err != nil {
		log.ErrS(err, payload.AppID, u)
		sendToRetryQueue(message)
		return
	}

	// No need to update in Elastic

	//
	message.Ack()
}

type steamSpyAppResponse struct {
	AppID          int    `json:"appid"`
	Name           string `json:"name"`
	Developer      string `json:"developer"`
	Publisher      string `json:"publisher"`
	Positive       int    `json:"positive"`
	Negative       int    `json:"negative"`
	Userscore      int    `json:"userscore"`
	Owners         string `json:"owners"`
	AverageForever int    `json:"average_forever"`
	Average2Weeks  int    `json:"average_2weeks"`
	MedianForever  int    `json:"median_forever"`
	Median2Weeks   int    `json:"median_2weeks"`
	Price          string `json:"price"`
	InitialPrice   string `json:"initialprice"`
	Discount       string `json:"discount"`
	Languages      string `json:"languages"`
	Genre          string `json:"genre"`
	CCU            int    `json:"ccu"`
	// ScoreRank      int    `json:"score_rank"` // Can be empty string
	// Tags           map[string]int `json:"tags"` // Can be an empty slice
}

func (a steamSpyAppResponse) GetOwners() (ret []int) {

	owners := strings.ReplaceAll(a.Owners, ",", "")
	owners = strings.ReplaceAll(owners, " ", "")
	ownersStrings := strings.Split(owners, "..")
	return helpers.StringSliceToIntSlice(ownersStrings)
}
