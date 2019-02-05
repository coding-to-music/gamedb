package queue2

import (
	"errors"
	"strconv"
	"time"

	"github.com/gamedb/website/helpers"
	"github.com/streadway/amqp"
)

func ProduceApps(IDs []int) (err error) {

	return produce(QueueGoApps, baseMessage{
		Message: appMessage{
			AppIDs: IDs,
			Time:   time.Now().Unix(),
		},
		FirstSeen:     time.Now(),
		Attempt:       1,
		OriginalQueue: QueueAppsGo,
		MaxAttempts:   2,
	})
}

type appMessage struct {
	AppIDs      []int `json:"IDs"`
	Time        int64 `json:"Time"`
	PICSAppInfo int
}

type AppQueue struct {
	baseQueue
}

func (q AppQueue) process(msg amqp.Delivery) {

	var err error
	var payload = baseMessage{
		Message: appMessage{},
	}

	err = helpers.Unmarshal(msg.Body, &payload)
	if err != nil {
		logError(err)
		payload.ack(msg)
		return
	}

	logInfo("Consuming app")
	logInfo(payload.Attempt)

	err = errors.New("test error")
	if err != nil {
		logError(err)
		payload.delay(msg)
		return
	}
}
