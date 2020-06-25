package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/websockets"
)

type WebsocketMessage struct {
	Pages   []websockets.WebsocketPage `json:"pages"`
	Message []byte                     `json:"message"`
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

			switch page {
			case websockets.PageApp, websockets.PageBundle, websockets.PagePackage:

				idPayload := IntPayload{}

				err = helpers.Unmarshal(payload.Message, &idPayload)
				log.Err(err)

				wsPage.Send(idPayload.ID)

			case websockets.PageGroup, websockets.PagePlayer:

				idPayload := StringPayload{}

				err = helpers.Unmarshal(payload.Message, &idPayload)
				log.Err(err)

				wsPage.Send(idPayload.String)

			case websockets.PageChatBot:

				cbPayload := ChatBotPayload{}

				err = helpers.Unmarshal(payload.Message, &cbPayload)
				log.Err(err)

				wsPage.Send(cbPayload)

			case websockets.PageAdmin:

				adminPayload := AdminPayload{}

				err = helpers.Unmarshal(payload.Message, &adminPayload)
				log.Err(err)

				wsPage.Send(adminPayload)

			case websockets.PageChanges:

				changePayload := ChangesPayload{}

				err = helpers.Unmarshal(payload.Message, &changePayload)
				log.Err(err)

				wsPage.Send(changePayload.Data)

			case websockets.PagePackages:

				idPayload := IntPayload{}

				err = helpers.Unmarshal(payload.Message, &idPayload)
				log.Err(err)

				pack, err := mongo.GetPackage(idPayload.ID)
				if err != nil {
					log.Err(err)
					continue
				}

				wsPage.Send(pack.OutputForJSON(steamapi.ProductCCUS))

			case websockets.PageBundles:

				idPayload := IntPayload{}

				err = helpers.Unmarshal(payload.Message, &idPayload)
				log.Err(err)

				bundle, err := mysql.GetBundle(idPayload.ID, nil)
				log.Err(err)
				if err == nil {
					wsPage.Send(bundle.OutputForJSON())
				}

			case websockets.PagePrices:

				idsPayload := StringsPayload{}

				err = helpers.Unmarshal(payload.Message, &idsPayload)
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

type IntPayload struct {
	ID int `json:"id"`
}

type StringPayload struct {
	String string `json:"id"`
}

type StringsPayload struct {
	IDs []string `json:"id"`
}

type ChangesPayload struct {
	Data [][]interface{} `json:"d"`
}

type AdminPayload struct {
	TaskID string `json:"task_id"`
	Action string `json:"action"`
	Time   int64  `json:"time"`
}

type ChatBotPayload struct {
	AuthorID     string `json:"author_id"`
	AuthorName   string `json:"author_name"`
	AuthorAvatar string `json:"author_avatar"`
	Message      string `json:"message"`
}

type ChatPayload struct {
	I            float32 `json:"i"`
	AuthorID     string  `json:"author_id"`
	AuthorUser   string  `json:"author_user"`
	AuthorAvatar string  `json:"author_avatar"`
	Content      string  `json:"content"`
	Channel      string  `json:"channel"`
	Time         string  `json:"timestamp"`
	Embeds       bool    `json:"embeds"`
}
