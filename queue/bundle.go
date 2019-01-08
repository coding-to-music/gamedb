package queue

import (
	"strconv"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/websockets"
	"github.com/streadway/amqp"
)

type RabbitMessageBundle struct {
	BundleID int
}

func (d RabbitMessageBundle) getConsumeQueue() RabbitQueue {
	return QueueBundlesData
}

func (d RabbitMessageBundle) getProduceQueue() RabbitQueue {
	return QueueBundlesData
}

func (d RabbitMessageBundle) getRetryData() RabbitMessageDelay {
	return RabbitMessageDelay{}
}

func (d RabbitMessageBundle) process(msg amqp.Delivery) (requeue bool, err error) {

	// Get message payload
	message := RabbitMessageBundle{}

	err = helpers.Unmarshal(msg.Body, &message)
	if err != nil {
		return false, err
	}

	logInfo("Consuming bundle: " + strconv.Itoa(message.BundleID))

	// Load current bundle
	gorm, err := db.GetMySQLClient()
	if err != nil {
		return true, err
	}

	bundle := db.Bundle{}
	gorm = gorm.FirstOrInit(&bundle, db.Bundle{ID: message.BundleID})
	if gorm.Error != nil {
		return true, gorm.Error
	}

	err = updateBundle(&bundle)
	if err != nil && err != steam.ErrAppNotFound {
		return true, err
	}

	// Save new data
	gorm = gorm.Save(&bundle)
	if gorm.Error != nil {
		return true, gorm.Error
	}

	// Send websocket
	page, err := websockets.GetPage(websockets.PageBundle)
	if err != nil {
		return true, err
	} else if page.HasConnections() {
		page.Send(bundle.ID)
	}

	page, err = websockets.GetPage(websockets.PageBundles)
	if err != nil {
		return true, err
	} else if page.HasConnections() {
		page.Send(bundle.ID)
	}

	return false, nil
}

func updateBundle(bundle *db.Bundle) error {

	return nil
}
