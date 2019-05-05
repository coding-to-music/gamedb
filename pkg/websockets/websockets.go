package websockets

import (
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
)

type WebsocketPage string

const (
	PageAdmin    WebsocketPage = "admin"
	PageApp      WebsocketPage = "app"
	PageBundle   WebsocketPage = "bundle"
	PageBundles  WebsocketPage = "bundles"
	PageChanges  WebsocketPage = "changes"
	PageChat     WebsocketPage = "chat"
	PageNews     WebsocketPage = "news"
	PagePackage  WebsocketPage = "package"
	PagePackages WebsocketPage = "packages"
	PagePrices   WebsocketPage = "prices"
	PagePlayers  WebsocketPage = "profile"
)

var (
	Pages = map[WebsocketPage]Page{}
)

func init() {

	pagesSlice := []WebsocketPage{
		PageChanges,
		PageChat,
		PageNews,
		PagePrices,
		PageAdmin,
		PageApp,
		PagePackage,
		PagePackages,
		PagePlayers,
		PageBundle,
		PageBundles,
	}
	for _, v := range pagesSlice {
		Pages[v] = Page{
			name:        v,
			connections: map[uuid.UUID]*websocket.Conn{},
		}
	}
}

func GetPage(page WebsocketPage) (ret Page) {

	if val, ok := Pages[page]; ok {
		return val
	}

	return ret
}

type Page struct {
	name        WebsocketPage
	connections map[uuid.UUID]*websocket.Conn
}

func (p Page) GetName() WebsocketPage {
	return p.name
}

func (p Page) CountConnections() int {
	return len(p.connections)
}

func (p *Page) AddConnection(conn *websocket.Conn) error {

	id := uuid.NewV4()

	p.connections[id] = conn

	return nil
}

func (p *Page) Send(data interface{}) {

	if p.CountConnections() > 0 {

		payload := WebsocketPayload{}
		payload.Page = p.name
		payload.Data = data

		for k, v := range p.connections {
			err := v.WriteJSON(payload)
			if err != nil {

				if strings.Contains(err.Error(), "broken pipe") {

					// Clean up old connections
					err := v.Close()
					log.Err(err)
					delete(p.connections, k)

				} else {
					log.Err(err)
				}
			}
		}
	}
}

type WebsocketPayload struct {
	Data  interface{}
	Page  WebsocketPage
	Error string
}

// Converts pubsub messages into websockets
func ListenToPubSub() {

	err := helpers.Subscribe(helpers.PubSubWebsockets, func(m *pubsub.Message) {

		log.Info("Incoming PubSub: " + string(m.Data))

		pubSubMsg := PubSubBasePayload{}
		err := helpers.Unmarshal(m.Data, &pubSubMsg)
		if err != nil {
			log.Critical(err)
			return
		}

		for _, page := range pubSubMsg.Pages {

			wsPage := GetPage(page)

			switch page {
			case PageApp, PageBundle, PageBundles, PagePackage, PagePackages:

				idPayload := PubSubIDPayload{}

				err = helpers.Unmarshal(m.Data, &idPayload)
				log.Err(err)

				wsPage.Send(idPayload.ID)

			case PagePlayers:

				idPayload := PubSubID64Payload{}

				err = helpers.Unmarshal(m.Data, &idPayload)
				log.Err(err)

				wsPage.Send(idPayload.ID)

			case PageChanges:

				changePayload := PubSubChangesPayload{}

				err = helpers.Unmarshal(m.Data, &changePayload)
				log.Err(err)

				wsPage.Send(changePayload.Data)

			default:
				log.Err("no handler for page: " + string(page))
			}
		}
	})
	log.Err(err)
}
