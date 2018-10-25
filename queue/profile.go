package queue

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/logging"
	"github.com/streadway/amqp"
)

type RabbitMessageProfile struct {
	Time     time.Time
	PlayerID int64
}

func (d *RabbitMessageProfile) Fill(playerID string) {

	d.Time = time.Now()

	playerIDInt, err := strconv.ParseInt(playerID, 10, 64)
	if err != nil {
		d.PlayerID = playerIDInt
		logging.Error(err)
	}
}

func (d *RabbitMessageProfile) ToBytes() []byte {
	bytes, err := json.Marshal(d)
	logging.Error(err)
	return bytes
}

func (d RabbitMessageProfile) getQueueName() string {
	return QueueProfilesData
}

func (d RabbitMessageProfile) getRetryData() RabbitMessageDelay {
	return RabbitMessageDelay{}
}

func (d RabbitMessageProfile) process(msg amqp.Delivery) (ack bool, requeue bool, err error) {

	// Get message
	message := new(RabbitMessageProfile)

	err = helpers.Unmarshal(msg.Body, message)
	if err != nil {
		return false, false, err
	}

	// Update player
	player, err := db.GetPlayer(int64(message.PlayerID))
	if err != nil {
		if err != db.ErrNoSuchEntity {
			logging.Error(err)
			return false, true, err
		}
	}

	errs := player.Update(db.PlayerUpdateManual, "")
	if len(errs) > 0 {
		for _, v := range errs {
			logging.Error(v)
		}

		return true, true, errs[0]
	}

	return true, false, nil
}
