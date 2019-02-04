package queue2

import (
	"errors"
	"time"

	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/streadway/amqp"
)

func ProduceApps(IDs []int) (err error) {

	return produce(QueueAppsGo, baseMessage{
		Message: appMessage{
			AppIDs: IDs,
			Time:   time.Now().Unix(),
		},
		FirstSeen: time.Now(),
		Attempt:   1,
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
		log.Err(err)
		payload.ack(msg)
		return
	}

	logInfo("Consuming app")
	logInfo(payload.Attempt)

	err = errors.New("test error")
	if err != nil {
		log.Err(err)
		payload.ackRetry(msg, q.name)
		return
	}
}
