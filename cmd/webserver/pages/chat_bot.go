package pages

import (
	"net/http"
	"sort"
	"time"

	"github.com/gamedb/website/pkg/chatbot"
	"github.com/gamedb/website/pkg/log"
	"github.com/go-chi/chi"
)

const (
	clientID = "567257603185311745"
)

func ChatBotRouter() http.Handler {
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
	t.Commands = chatbot.CommandRegister

	sort.Slice(t.Commands, func(i, j int) bool {
		return t.Commands[i].Example() > t.Commands[j].Example()
	})

	err := returnTemplate(w, r, "chat_bot", t)
	log.Err(err, r)
}

type chatBotTemplate struct {
	GlobalTemplate
	Commands []chatbot.Command
}

func (cbt chatBotTemplate) AddBotLink() string {
	return "https://discordapp.com/oauth2/authorize?&client_id=" + clientID + "&scope=bot&permissions=2048"
}
