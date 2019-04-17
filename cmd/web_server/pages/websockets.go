package pages

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gamedb/website/pkg"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func websocketsRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/{id:[a-z]+}", websocketsHandler)
	return r
}

func websocketsHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	page, err := pkg.GetPage(pkg.WebsocketPage(id))
	if err != nil {

		bytes, err := json.Marshal(pkg.WebsocketPayload{Error: "Invalid page"})
		log.Err(err)

		_, err = w.Write(bytes)
		log.Err(err)
		return
	}

	// Upgrade the connection
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if !strings.Contains(err.Error(), "websocket: not a websocket handshake") {
			log.Err(err)
		}
		return
	}

	err = page.SetConnection(connection)
	if err != nil {
		log.Err(err)
	}
}
