package queue

import (
	"encoding/json"
	"strings"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/websockets"
	"github.com/streadway/amqp"
)

func processChange(msg amqp.Delivery) (ack bool, requeue bool) {

	// Get change
	change := new(datastore.Change)

	err := json.Unmarshal(msg.Body, change)
	if err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(msg.Body))
		}

		return false, false
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
			CreatedAtUnix: change.CreatedAt.Unix(),
			CreatedAtNice: change.CreatedAt.Format(helpers.DateYearTime),
			Apps:          apps,
			Packages:      packages,
		}
		websockets.Send(websockets.CHANGES, payload)
	}

	return true, false
}

type changeWebsocketPayload struct {
	ID            int                             `json:"id"`
	CreatedAtUnix int64                           `json:"created_at"`
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
