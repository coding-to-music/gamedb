package queue

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/steam"
	"github.com/streadway/amqp"
)

func processPlayer(msg amqp.Delivery) {

	// Get message
	message := new(PlayerMessage)

	err := json.Unmarshal(msg.Body, message)
	if err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(msg.Body))
		}

		msg.Nack(false, false)
		return
	}

	// Update player
	player, err := datastore.GetPlayer(message.PlayerID)
	if err != nil {
		logger.Error(err)
	}

	errs := player.UpdateIfNeeded()
	if len(errs) > 0 {
		for _, v := range errs {
			logger.Error(v)
		}

		// API is probably down
		for _, v := range errs {
			if v.Error() == steam.ErrInvalidJson {
				time.Sleep(time.Second * 10)
				msg.Nack(false, true)
				return
			}
		}
	}

	msg.Ack(false)
	return
}

type PlayerMessage struct {
	Time     time.Time
	PlayerID int
}
