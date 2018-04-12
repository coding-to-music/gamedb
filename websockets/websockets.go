package websockets

import (
	"math/rand"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/steam-authority/steam-authority/logger"
)

const (
	CHANGES = "changes"
	CHAT    = "chat"
	NEWS    = "news"
	PRICES  = "prices"
)

var connections = map[int]*websocket.Conn{}
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func HasConnections() (bool) {
	return len(connections) > 0
}

func Send(page string, data interface{}) {

	payload := websocketPayload{}
	payload.Page = page
	payload.Data = data

	for k, v := range connections {
		err := v.WriteJSON(payload)
		if err != nil {

			// Clean up old connections
			if strings.Contains(err.Error(), "broken pipe") {

				v.Close()
				delete(connections, k)

			} else {
				logger.Error(err)
			}
		}
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {

	// Upgrade the connection
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if !strings.Contains(err.Error(), "websocket: not a websocket handshake") {
			logger.Error(err)
		}
		return
	}

	connections[rand.Int()] = connection
}

// Properties must be exported so websocket can read them.
type websocketPayload struct {
	Data interface{}
	Page string
}
