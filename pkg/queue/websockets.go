package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
)

type WebsocketMessage struct {
	Pages   []websockets.WebsocketPage `json:"pages"`
	Message string                     `json:"message"`
}

func websocketHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := WebsocketMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		for _, page := range payload.Pages {

			wsPage := websockets.GetPage(page)
			if wsPage == nil {
				continue
			}

			if wsPage.CountConnections() == 0 {
				continue
			}

			var data = []byte(payload.Message)

			switch page {
			case websockets.PageApp, websockets.PageBundle, websockets.PagePackage:

				idPayload := websockets.IntPayload{}

				err = helpers.Unmarshal(data, &idPayload)
				log.Err(err)

				wsPage.Send(idPayload.ID)

			case websockets.PageGroup, websockets.PagePlayer:

				idPayload := websockets.StringPayload{}

				err = helpers.Unmarshal(data, &idPayload)
				log.Err(err)

				wsPage.Send(idPayload.String)

			case websockets.PageChatBot:

				cbPayload := websockets.ChatBotPayload{}

				err = helpers.Unmarshal(data, &cbPayload)
				log.Err(err)

				wsPage.Send(cbPayload)

			case websockets.PageAdmin:

				adminPayload := websockets.AdminPayload{}

				err = helpers.Unmarshal(data, &adminPayload)
				log.Err(err)

				wsPage.Send(adminPayload)

			case websockets.PageChanges:

				changePayload := websockets.ChangesPayload{}

				err = helpers.Unmarshal(data, &changePayload)
				log.Err(err)

				wsPage.Send(changePayload.Data)

			case websockets.PagePackages:

				idPayload := websockets.IntPayload{}

				err = helpers.Unmarshal(data, &idPayload)
				log.Err(err)

				pack, err := mongo.GetPackage(idPayload.ID)
				if err != nil {
					log.Err(err)
					continue
				}

				wsPage.Send(pack.OutputForJSON(steamapi.ProductCCUS))

			case websockets.PageBundles:

				idPayload := websockets.IntPayload{}

				err = helpers.Unmarshal(data, &idPayload)
				log.Err(err)

				bundle, err := sql.GetBundle(idPayload.ID, nil)
				log.Err(err)
				if err == nil {
					wsPage.Send(bundle.OutputForJSON())
				}

			case websockets.PagePrices:

				idsPayload := websockets.StringsPayload{}

				err = helpers.Unmarshal(data, &idsPayload)
				log.Err(err)

				prices, err := mongo.GetPricesByID(idsPayload.IDs)
				log.Err(err)
				if err == nil {
					for _, v := range prices {
						wsPage.Send(v.OutputForJSON())
					}
				}

			default:
				log.Err("no handler for page: " + string(page))
			}
		}

		//
		message.Ack(false)
	}
}
