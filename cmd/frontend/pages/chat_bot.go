package pages

import (
	"github.com/gamedb/gamedb/pkg/i18n"
	"net/http"
	"sort"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/config"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	influx "github.com/influxdata/influxdb1-client"
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
	t.fill(w, r, "chat_bot", "Steam Discord Chat Bot", "Steam Discord Chat Bot")
	t.addAssetJSON2HTML()
	t.Link = config.C.DiscordBotInviteURL
	t.Regions = i18n.GetProdCCs()

	returnTemplate(w, r, t)
}

type chatBotTemplate struct {
	globalTemplate
	Link    string
	Regions []i18n.ProductCountryCode
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

	err := memcache.GetSetInterface(memcache.MemcacheChatBotGuildsCount, &guilds, func() (i interface{}, err error) {

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

		// Save to Influx
		_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, influx.Point{
			Measurement: influxHelper.InfluxMeasurementChatBot.String(),
			Fields: map[string]interface{}{
				"guilds": count,
			},
			Precision: "h",
		})
		if err != nil {
			log.ErrS(err)
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
	if err != nil {
		log.ErrS(err)
	}

	var response = datatable.NewDataTablesResponse(r, query, 100, 100, nil)
	for _, command := range commands {

		response.AddRow(command.GetTableRowJSON(guilds))
	}

	returnJSON(w, r, response)
}
