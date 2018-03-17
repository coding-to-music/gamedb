package queue

import (
	"encoding/json"

	"github.com/Jleagle/go-helpers/logger"
	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/steam"
	"github.com/streadway/amqp"
)

func processPlayer(msg amqp.Delivery) (err error) {

	// Get message
	message := new(PlayerMessage)

	err = json.Unmarshal(msg.Body, message)
	if err != nil {
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
			if v.Error() == steam.ErrorInvalidJson {
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
