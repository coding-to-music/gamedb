package queue

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type AppSteamspyMessage struct {
	ID int `json:"id"`
}

func appSteamspyHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := AppSteamspyMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		query := url.Values{}
		query.Set("request", "appdetails")
		query.Set("appid", strconv.Itoa(payload.ID))

		// Create request
		client := &http.Client{}
		client.Timeout = time.Second * 5

		ssURL := "https://steamspy.com/api.php?" + query.Encode()
		req, err := http.NewRequest("GET", ssURL, nil)
		if err != nil {
			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			continue
		}

		response, err := client.Do(req)
		if err != nil {

			if strings.Contains(err.Error(), "Client.Timeout exceeded while awaiting headers") {
				log.Info(err, payload.ID)
			} else {
				log.Err(err, payload.ID)
			}

			time.Sleep(time.Second * 5)
			sendToRetryQueue(message)
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
			log.Err(errors.New("steamspy is down (1): "+ssURL), payload.ID)
			sendToRetryQueue(message)
			continue
		}

		bytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			continue
		}

		if strings.Contains(string(bytes), "Connection failed") {
			log.Err(errors.New("steamspy is down (2): "+ssURL), payload.ID)
			sendToRetryQueue(message)
			continue
		}

		// Unmarshal JSON
		resp := mongo.SteamSpyAppResponse{}
		err = helpers.Unmarshal(bytes, &resp)
		if err != nil {
			log.Err(errors.New("steamspy is down (3): "+ssURL), payload.ID)
			sendToRetryQueue(message)
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
