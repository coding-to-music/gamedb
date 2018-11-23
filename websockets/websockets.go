package websockets

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"strings"

	"github.com/gamedb/website/logging"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
)

type WebsocketPage string

const (
	PageChanges  WebsocketPage = "changes"
	PageChat     WebsocketPage = "chat"
	PageNews     WebsocketPage = "news"
	PagePrices   WebsocketPage = "prices"
	PageProfile  WebsocketPage = "profile"
	PageAdmin    WebsocketPage = "admin"
	PagePackages WebsocketPage = "packages"
	PageApps     WebsocketPage = "apps"
)

var pages map[WebsocketPage]Page

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var ErrInvalidPage = errors.New("invalid page")

func init() {
	pages = map[WebsocketPage]Page{}
	for _, v := range []WebsocketPage{PageChanges, PageChat, PageNews, PagePrices, PageProfile, PageAdmin, PagePackages} {
		pages[v] = Page{
			name:        v,
			connections: map[int]*websocket.Conn{},
		}
	}
}

func WebsocketsHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	page, err := GetPage(WebsocketPage(id))
	if err != nil {

		bytes, err := json.Marshal(websocketPayload{Error: "Invalid page"})
		logging.Error(err)

		_, err = w.Write(bytes)
		logging.Error(err)
		return
	}

	// Upgrade the connection
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if !strings.Contains(err.Error(), "websocket: not a websocket handshake") {
			logging.Error(err)
		}
		return
	}

	page.setConnection(connection)
}

func GetPage(page WebsocketPage) (p Page, err error) {

	if val, ok := pages[page]; ok {
		return val, nil
	}
	return p, ErrInvalidPage
}

type Page struct {
	name        WebsocketPage
	connections map[int]*websocket.Conn
}

func (p Page) HasConnections() bool {
	return len(p.connections) > 0
}

func (p *Page) setConnection(conn *websocket.Conn) {
	p.connections[rand.Int()] = conn
}

func (p *Page) Send(data interface{}) {

	if !p.HasConnections() {
		return
	}

	payload := websocketPayload{}
	payload.Page = p.name
	payload.Data = data

	for k, v := range p.connections {
		err := v.WriteJSON(payload)
		if err != nil {

			// Clean up old connections
			if strings.Contains(err.Error(), "broken pipe") {

				err := v.Close()
				logging.Error(err)
				delete(p.connections, k)

			} else {
				logging.Error(err)
			}
		}
	}
}

type websocketPayload struct {
	Data  interface{}
	Page  WebsocketPage
	Error string
}
