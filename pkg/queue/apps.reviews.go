package queue

import (
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	steamHelper "github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	influx "github.com/influxdata/influxdb1-client"
	"go.mongodb.org/mongo-driver/bson"
)

type AppReviewsMessage struct {
	AppID int `json:"id"`
}

func appReviewsHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := AppReviewsMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		resp, _, err := steamHelper.GetSteamUnlimited().GetReviews(payload.AppID)
		err = steamHelper.AllowSteamCodes(err)
		if err != nil {
			steamHelper.LogSteamError(err)
			sendToRetryQueue(message)
			continue
		}

		//
		reviews := helpers.AppReviewSummary{}
		reviews.Positive = resp.QuerySummary.TotalPositive
		reviews.Negative = resp.QuerySummary.TotalNegative

		// Make slice of playerIDs
		var playersSlice []int64
		for _, v := range resp.Reviews {
			playersSlice = append(playersSlice, int64(v.Author.SteamID))
		}

		// Get players
		players, err := mongo.GetPlayersByID(playersSlice, bson.M{"_id": 1, "persona_name": 1})
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			continue
		}

		// Make map of players
		var playersMap = map[int64]mongo.Player{}
		for _, player := range players {
			playersMap[player.ID] = player
		}

		// Make template slice
		for _, v := range resp.Reviews {

			var player mongo.Player
			if val, ok := playersMap[int64(v.Author.SteamID)]; ok {
				player = val
			} else {
				player.ID = int64(v.Author.SteamID)

				err = ProducePlayer(PlayerMessage{ID: int64(v.Author.SteamID), SkipGroupUpdate: true})
				err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
				log.Err(err)
			}

			// Remove extra new lines
			v.Review = helpers.RegexMultipleNewLines.ReplaceAllString(v.Review, "\n\n")

			reviews.Reviews = append(reviews.Reviews, helpers.AppReview{
				Review:     helpers.BBCodeCompiler.Compile(v.Review),
				PlayerPath: player.GetPath(),
				PlayerName: player.GetName(),
				Created:    time.Unix(v.TimestampCreated, 0).Format(helpers.DateYear),
				VotesGood:  v.VotesUp,
				VotesFunny: v.VotesFunny,
				Vote:       v.VotedUp,
			})
		}

		// Set score
		var score float64
		if reviews.GetTotal() > 0 {
			// https://planspace.org/2014/08/17/how-to-sort-by-average-rating/
			var a = 1
			var b = 2
			score = (float64(reviews.Positive+a) / float64(reviews.GetTotal()+b)) * 100
		}

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {

			defer wg.Done()

			// Sort by upvotes
			sort.Slice(reviews.Reviews, func(i, j int) bool {
				return reviews.Reviews[i].VotesGood > reviews.Reviews[j].VotesGood
			})

			var update = bson.D{
				{"reviews_score", score},
				{"reviews", reviews},
				{"reviews_count", reviews.GetTotal()},
			}

			_, err = mongo.UpdateOne(mongo.CollectionApps, bson.D{{"_id", payload.AppID}}, update)
			if err != nil {
				log.Err(err, payload.AppID)
				sendToRetryQueue(message)
				return
			}
		}()

		wg.Add(1)
		go func() {

			defer wg.Done()

			err = memcache.Delete(memcache.MemcacheApp(payload.AppID).Key)
			if err != nil {
				log.Err(err, payload.AppID)
				sendToRetryQueue(message)
				return
			}
		}()

		wg.Add(1)
		go func() {

			defer wg.Done()

			_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, influx.Point{
				Measurement: string(influxHelper.InfluxMeasurementApps),
				Tags: map[string]string{
					"app_id": strconv.Itoa(payload.AppID),
				},
				Fields: map[string]interface{}{
					"reviews_score":    score,
					"reviews_positive": reviews.Positive,
					"reviews_negative": reviews.Negative,
				},
				Time:      time.Now(),
				Precision: "m",
			})
			if err != nil {
				log.Err(err, payload.AppID)
				sendToRetryQueue(message)
				return
			}
		}()

		wg.Wait()
		if message.ActionTaken {
			continue
		}

		message.Ack(false)
	}
}
