package queue

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/steam-authority/steam-authority/db"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logging"
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

func processPlayer(msg amqp.Delivery) (ack bool, requeue bool) {

	// Get message
	message := new(RabbitMessageProfile)

	err := helpers.Unmarshal(msg.Body, message)
	if err != nil {
		return false, false
	}

	// Update player
	player, err := db.GetPlayer(int64(message.PlayerID))
	if err != nil {
		if err != db.ErrNoSuchEntity {
			logging.Error(err)
			return false, true
		}
	}

	errs := player.Update("")
	if len(errs) > 0 {
		for _, v := range errs {
			logging.Error(v)
		}

		// API is probably down, todo
		//for _, v := range errs {
		//	if v.Error() == steam.ErrInvalidJson {
		//		return false, true
		//	}
		//}
	}

	return true, false
}
