package consumers

import (
	"sort"
	"strconv"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/steam"
	influx "github.com/influxdata/influxdb1-client"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type AppReviewsMessage struct {
	AppID int `json:"id"`
}

func (m AppReviewsMessage) Queue() rabbit.QueueName {
	return QueueAppsReviews
}

func appReviewsHandler(message *rabbit.Message) {

	payload := AppReviewsMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	respAll, err := steam.GetSteamUnlimited().GetReviews(payload.AppID, "all")
	err = steam.AllowSteamCodes(err)
	if err != nil {
		steam.LogSteamError(err)
		sendToRetryQueue(message)
		return
	}

	respEnglish, err := steam.GetSteamUnlimited().GetReviews(payload.AppID, steamapi.LanguageEnglish)
	err = steam.AllowSteamCodes(err)
	if err != nil {
		steam.LogSteamError(err)
		sendToRetryQueue(message)
		return
	}

	//
	reviews := helpers.AppReviewSummary{}
	reviews.Positive = respAll.QuerySummary.TotalPositive
	reviews.Negative = respAll.QuerySummary.TotalNegative

	// Make slice of playerIDs
	var playersSlice []int64
	for _, v := range respAll.Reviews {
		playersSlice = append(playersSlice, int64(v.Author.SteamID))
	}

	// Get players
	players, err := mongo.GetPlayersByID(playersSlice, bson.M{"_id": 1, "persona_name": 1})
	if err != nil {
		log.ErrS(err)
		sendToRetryQueue(message)
		return
	}

	// Make map of players
	var playersMap = map[int64]mongo.Player{}
	for _, player := range players {
		playersMap[player.ID] = player
	}

	// Make template slice
	var missingPlayers bool

	for _, review := range respEnglish.Reviews {

		var player mongo.Player
		if val, ok := playersMap[int64(review.Author.SteamID)]; ok {
			player = val
		} else {

			player.ID = int64(review.Author.SteamID)

			err = ProducePlayer(PlayerMessage{ID: int64(review.Author.SteamID), SkipExistingPlayer: true}, "queue-reviews")
			err = helpers.IgnoreErrors(err, ErrInQueue)
			if err != nil {
				log.ErrS(err)
			}

			missingPlayers = true
		}

		// Remove extra new lines
		review.Review = helpers.RegexNewLines.ReplaceAllString(review.Review, "\n\n")

		reviews.Reviews = append(reviews.Reviews, helpers.AppReview{
			Review:     review.Review,
			PlayerPath: player.GetPath(),
			PlayerName: player.GetName(),
			Created:    time.Unix(review.TimestampCreated, 0).Format(helpers.DateYear),
			VotesGood:  review.VotesUp,
			VotesFunny: review.VotesFunny,
			Vote:       review.VotedUp,
		})
	}

	// Retry later to get queued players
	if missingPlayers {
		// Disable for now
		// sendToRetryQueueWithDelay(message, time.Minute)
		// return
	}

	// Set score
	var score float64
	if reviews.GetTotal() > 0 {
		// https://planspace.org/2014/08/17/how-to-sort-by-average-rating/
		var a = 1
		var b = 2
		score = (float64(reviews.Positive+a) / float64(reviews.GetTotal()+b)) * 100
	}

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
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	err = memcache.Client().Delete(memcache.ItemApp(payload.AppID).Key)
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	// Update in Elastic
	updateInElastic := map[string]interface{}{
		"reviews_count": reviews.GetTotal(),
	}

	err = ProduceAppSearch(nil, payload.AppID, updateInElastic)
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	// Add to Influx
	point := influx.Point{
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
	}

	_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, point)
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	message.Ack()
}
