package pages

import (
	"net/http"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

const (
	chatBotClientID = "567257603185311745"
)

func ChatBotRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", chatBotHandler)
	r.Get("/commands.json", chatBotRecentHandler)
	return r
}

func chatBotHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := chatBotTemplate{}
	t.fill(w, r, "Steam Bot", "Steam Discord Chat Bot")
	t.addAssetJSON2HTML()

	returnTemplate(w, r, "chat_bot", t)
}

type chatBotTemplate struct {
	globalTemplate
}

func (cbt chatBotTemplate) Link() string {
	return "https://discordapp.com/oauth2/authorize?client_id=" + chatBotClientID + "&scope=bot&permissions=0"
}

func (cbt chatBotTemplate) Commands() (ret [][]chatbot.Command) {

	var groupedMap = map[chatbot.CommandType][]chatbot.Command{}
	for _, v := range chatbot.CommandRegister {

		if _, ok := groupedMap[v.Type()]; ok {

			groupedMap[v.Type()] = append(groupedMap[v.Type()], v)

		} else {

			groupedMap[v.Type()] = []chatbot.Command{v}

		}
	}

	for _, v := range groupedMap {

		sort.Slice(v, func(i, j int) bool {
			return v[i].Example() < v[j].Example()
		})

		ret = append(ret, v)
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i][0].Type() < ret[j][0].Type()
	})

	return ret
}

func (cbt chatBotTemplate) Guilds() (guilds int) {

	var item = memcache.MemcacheChatBotGuildsCount

	err := memcache.GetSetInterface(item.Key, item.Expiration, &guilds, func() (i interface{}, err error) {

		discordChatBotSession, err := discordgo.New("Bot " + config.Config.DiscordChatBotToken.Get())
		if err != nil {
			return i, err
		}

		after := ""
		more := true
		count := 1

		for more {

			guilds, err := discordChatBotSession.UserGuilds(100, "", after)
			if err != nil {
				return i, err
			}

			if len(guilds) < 100 {
				more = false
			}

			for _, v := range guilds {
				after = v.ID
			}

			count += len(guilds)
		}

		return count, nil
	})

	zap.S().Error(err)

	return guilds
}

func chatBotRecentHandler(w http.ResponseWriter, r *http.Request) {

	commands, err := mongo.GetChatBotCommandsRecent()
	if err != nil {
		zap.S().Error(err)
		return
	}

	var last string
	var messages []queue.ChatBotPayload
	for _, v := range commands {

		// Show all command prefixes as a full stop
		if strings.HasPrefix(v.Message, "!") {
			v.Message = strings.Replace(v.Message, "!", ".", 1)
		}

		if last != v.AuthorID+v.Message { // Stop dupes

			messages = append(messages, queue.ChatBotPayload{
				AuthorID:     v.AuthorID,
				AuthorName:   v.AuthorName,
				AuthorAvatar: v.AuthorAvatar,
				Message:      v.Message,
			})
		}

		last = v.AuthorID + v.Message
	}

	returnJSON(w, r, messages)
}
