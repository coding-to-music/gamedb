package queue

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/logging"
	"github.com/streadway/amqp"
)

type RabbitMessageProfile struct {
	Time       time.Time
	PlayerID   int64
	UserAgent  string
	RemoteAddr string
	UpdateType db.UpdateType
}

func (d *RabbitMessageProfile) Fill(r *http.Request, playerID int64, updateType db.UpdateType) {

	d.Time = time.Now()
	d.PlayerID = playerID
	d.UserAgent = r.Header.Get("User-Agent")
	d.RemoteAddr = r.RemoteAddr
	d.UpdateType = updateType
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

	r := new(http.Request)
	r.Header.Set("User-Agent", d.UserAgent)
	r.RemoteAddr = d.RemoteAddr

	err = player.Update(r, db.PlayerUpdateManual)
	if err != nil {
		logging.Error(err)
		return false, true, err
	}

	return true, false, nil
}
