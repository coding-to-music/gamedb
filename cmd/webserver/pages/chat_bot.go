package pages

import (
	"net/http"
	"sort"
	"time"

	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
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

	// Get amount of guilds
	func() {

		client, err := helpers.GetDiscord()
		if err != nil {
			log.Warning(err)
			return
		}

		after := ""
		more := true

		for more {

			guilds, err := client.UserGuilds(100, "", after)
			if err != nil {
				log.Warning(err)
				return
			}

			if len(guilds) < 100 {
				more = false
			}

			for _, v := range guilds {
				after = v.ID
			}

			t.Guilds = t.Guilds + len(guilds)
		}
	}()

	//
	sort.Slice(t.Commands, func(i, j int) bool {
		return t.Commands[i].Example() < t.Commands[j].Example()
	})

	err := returnTemplate(w, r, "chat_bot", t)
	log.Err(err, r)
}

type chatBotTemplate struct {
	GlobalTemplate
	Commands []chatbot.Command
	Guilds   int
}

func (cbt chatBotTemplate) AddBotLink() string {
	return "https://discordapp.com/oauth2/authorize?&client_id=" + clientID + "&scope=bot&permissions=2048"
}
