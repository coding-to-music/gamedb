package queue

import (
	"encoding/json"
	"time"

	"github.com/Jleagle/go-helpers/logger"
	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/websockets"
	"github.com/streadway/amqp"
)

func processChange(msg amqp.Delivery) (err error) {

	// Get change
	change := new(datastore.Change)

	err = json.Unmarshal(msg.Body, change)
	if err != nil {
		msg.Nack(false, false)
		return nil
	}

	// Save change to DS
	_, err = datastore.SaveKind(change.GetKey(), change)
	if err != nil {
		logger.Error(err)
	}

	// Send websocket
	if websockets.HasConnections() {

		// Get apps for change
		var apps []changeAppWebsocketPayload
		appsResp, err := mysql.GetApps(change.Apps, []string{"id", "name"})
		if err != nil {
			logger.Error(err)
		}

		for _, v := range appsResp {
			apps = append(apps, changeAppWebsocketPayload{
				ID:   v.ID,
				Name: v.GetName(),
			})
		}

		// Get packages for change
		var packages []changePackageWebsocketPayload
		packagesResp, err := mysql.GetPackages(change.Packages, []string{"id", "name"})
		if err != nil {
			logger.Error(err)
		}

		for _, v := range packagesResp {
			packages = append(packages, changePackageWebsocketPayload{
				ID:   v.ID,
				Name: v.GetName(),
			})
		}

		payload := changeWebsocketPayload{
			ID:            change.ChangeID,
			CreatedAt:     change.CreatedAt.Unix(),
			CreatedAtNice: change.CreatedAt.Format(time.RFC822),
			Apps:          apps,
			Packages:      packages,
		}
		websockets.Send(websockets.CHANGES, payload)
	}

	msg.Ack(false)
	return nil
}

type changeWebsocketPayload struct {
	ID            int                             `json:"id"`
	CreatedAt     int64                           `json:"created_at"`
	CreatedAtNice string                          `json:"created_at_nice"`
	Apps          []changeAppWebsocketPayload     `json:"apps"`
	Packages      []changePackageWebsocketPayload `json:"packages"`
}

type changeAppWebsocketPayload struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type changePackageWebsocketPayload struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
