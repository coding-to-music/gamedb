package websockets

import (
	"fmt"
	"strings"
	"sync"

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
	PageGroup    WebsocketPage = "group"
	PageNews     WebsocketPage = "news"
	PagePackage  WebsocketPage = "package"
	PagePackages WebsocketPage = "packages"
	PagePrices   WebsocketPage = "prices"
	PagePlayer   WebsocketPage = "profile"
	PageChatBot  WebsocketPage = "chat-bot"
)

var (
	Pages = map[WebsocketPage]*Page{}
)

func init() {

	pagesSlice := []WebsocketPage{
		PageAdmin,
		PageApp,
		PageBundle,
		PageBundles,
		PageChanges,
		PageChat,
		PageGroup,
		PageNews,
		PagePackage,
		PagePackages,
		PagePrices,
		PagePlayer,
		PageChatBot,
	}
	for _, v := range pagesSlice {
		Pages[v] = &Page{
			name:        v,
			connections: map[uuid.UUID]*websocket.Conn{},
		}
	}
}

func GetPage(page WebsocketPage) (ret *Page) {

	if val, ok := Pages[page]; ok {
		return val
	}

	return ret
}

type Page struct {
	name        WebsocketPage
	connections map[uuid.UUID]*websocket.Conn
	sync.Mutex
}

func (p Page) GetName() WebsocketPage {
	return p.name
}

func (p Page) CountConnections() int {

	// p.Lock()
	// defer p.Unlock()

	return len(p.connections)
}

func (p *Page) AddConnection(conn *websocket.Conn) {

	p.Lock()
	defer p.Unlock()

	p.connections[uuid.NewV4()] = conn
}

func (p *Page) Send(data interface{}) {

	p.Lock()
	defer p.Unlock()

	count := p.CountConnections()
	if count > 0 {

		var connsToDelete []uuid.UUID

		payload := WebsocketPayload{}
		payload.Page = p.name
		payload.Data = data
		payload.Subs = count

		for k, v := range p.connections {

			err := v.WriteJSON(payload)
			if err != nil {

				connsToDelete = append(connsToDelete, k)

				if !strings.Contains(err.Error(), "broken pipe") &&
					!strings.Contains(err.Error(), "connection reset by peer") {
					log.Err(err, fmt.Sprint(payload))
				}
			}
		}

		for _, v := range connsToDelete {
			delete(p.connections, v)
		}
	}
}

type WebsocketPayload struct {
	Data  interface{}
	Page  WebsocketPage
	Error string
	Subs  int
}
