package web

import (
	"net/http"
	"time"

	"github.com/gamedb/website/log"
	"github.com/go-chi/chi"
)

const (
	clientID = "567257603185311745"
)

func chatBotRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", chatBotHandler)
	return r
}

func chatBotHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour*24)

	// Template
	t := chatBotTemplate{}
	t.fill(w, r, "Chat", "The Game DB community.")

	err := returnTemplate(w, r, "chat_bot", t)
	log.Err(err, r)
}

type chatBotTemplate struct {
	GlobalTemplate
}

func (cbt chatBotTemplate) AddBotLink() string {
	return "https://discordapp.com/oauth2/authorize?&client_id=" + clientID + "&scope=bot&permissions=2048"
}
