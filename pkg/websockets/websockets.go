package websockets

import (
	"strings"
	"sync"

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
	PagePlayer   WebsocketPage = "profile"
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
		PagePlayer,
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
	mutex       sync.RWMutex
}

func (p Page) GetName() WebsocketPage {
	return p.name
}

func (p Page) CountConnections() int {
	return len(p.connections)
}

func (p *Page) AddConnection(conn *websocket.Conn) {

	id := uuid.NewV4()

	p.mutex.Lock()
	p.connections[id] = conn
	p.mutex.Unlock()
}

func (p *Page) Send(data interface{}) {

	if p.CountConnections() > 0 {

		var connsToDelete []uuid.UUID

		payload := WebsocketPayload{}
		payload.Page = p.name
		payload.Data = data

		p.mutex.RLock()
		for k, v := range p.connections {

			err := v.WriteJSON(payload)
			if err != nil {

				if strings.Contains(err.Error(), "broken pipe") {

					connsToDelete = append(connsToDelete, k)

				} else {

					log.Err(err)
				}
			}
		}
		p.mutex.RUnlock()

		p.mutex.Lock()
		for _, v := range connsToDelete {
			delete(p.connections, v)
		}
		p.mutex.Unlock()
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

		// log.Info("PubSub (" + humanize.Bytes(uint64(len(m.Data))) + "): " + string(m.Data))

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

			case PagePlayer:

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
