package websockets

import (
	"errors"
	"strings"

	"github.com/gamedb/website/pkg/log"
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
	PageProfile  WebsocketPage = "profile"
)

var (
	ErrInvalidPage = errors.New("invalid page")
	pages          = map[WebsocketPage]Page{}
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
		PageProfile,
		PageBundle,
		PageBundles,
	}
	for _, v := range pagesSlice {
		pages[v] = Page{
			name:        v,
			connections: map[uuid.UUID]*websocket.Conn{},
		}
	}
}

func GetPage(page WebsocketPage) (ret Page) {

	if val, ok := pages[page]; ok {
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

func (p Page) HasConnections() bool {
	return len(p.connections) > 0
}

func (p *Page) AddConnection(conn *websocket.Conn) error {

	id := uuid.NewV4()

	p.connections[id] = conn

	return nil
}

func (p *Page) Send(data interface{}) {

	if !p.HasConnections() {
		return
	}

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

type WebsocketPayload struct {
	Data  interface{}
	Page  WebsocketPage
	Error string
}
