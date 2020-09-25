package pages

import (
	"net/http"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
)

const (
	chatBotClientID = "567257603185311745"
)

func ChatBotRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", chatBotHandler)
	r.Get("/recent.json", chatBotRecentHandler)
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

	if config.C.DiscordChatBotToken == "" {
		log.ErrS("Missing environment variables")
		return 0
	}

	var item = memcache.MemcacheChatBotGuildsCount

	err := memcache.GetSetInterface(item.Key, item.Expiration, &guilds, func() (i interface{}, err error) {

		discordChatBotSession, err := discordgo.New("Bot " + config.C.DiscordChatBotToken)
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

	if err != nil {
		log.ErrS(err)
	}

	return guilds
}

func chatBotRecentHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, false)

	commands, err := mongo.GetChatBotCommandsRecent()
	if err != nil {
		log.ErrS(err)
		return
	}

	var guildIDs []string
	for _, v := range commands {
		guildIDs = append(guildIDs, v.GuildID)
	}

	guilds, err := mongo.GetGuilds(guildIDs)

	var last string
	var response = datatable.NewDataTablesResponse(r, query, 100, 100, nil)
	for _, command := range commands {

		// Show all command prefixes as a full stop
		if strings.HasPrefix(command.Message, "!") {
			command.Message = strings.Replace(command.Message, "!", ".", 1)
		}

		if last != command.AuthorID+command.Message { // Stop dupes

			response.AddRow([]interface{}{
				command.AuthorID,                     // 0
				command.AuthorName,                   // 1
				command.AuthorAvatar,                 // 2
				command.Message,                      // 3
				command.Time.Unix(),                  // 4
				command.Time.Format(helpers.DateSQL), // 5
				guilds[command.GuildID].Name,         // 6
			})
		}

		last = command.AuthorID + command.Message
	}

	returnJSON(w, r, response)
}
