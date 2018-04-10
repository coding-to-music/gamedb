package queue

import (
	"encoding/json"
	"strings"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/steam"
	"github.com/streadway/amqp"
)

func processPlayer(msg amqp.Delivery) (err error) {

	// Get message
	message := new(PlayerMessage)

	err = json.Unmarshal(msg.Body, message)
	if err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(msg.Body))
		}

		msg.Nack(false, false)
		return nil
	}

	// Update player
	player, err := datastore.GetPlayer(message.PlayerID)
	if err != nil {
		logger.Error(err)
	}

	errs := player.UpdateIfNeeded()
	if len(errs) > 0 {
		for _, v := range errs {

			logger.Error(err)

			// API is probably down
			if v.Error() == steam.ErrInvalidJson {
				msg.Nack(false, true)
				return nil
			}
		}
	}

	msg.Ack(false)
	return nil
}

type PlayerMessage struct {
	PlayerID int
}
