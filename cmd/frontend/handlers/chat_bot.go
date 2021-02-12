package handlers

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/i18n"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func ChatBotRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", chatBotHandler)
	r.Get("/chart.json", chatbotChartAjaxHandler)
	r.Get("/recent.json", chatBotRecentHandler)
	return r
}

func chatBotHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := chatBotTemplate{}
	t.fill(w, r, "chat_bot", "Steam Discord Chat Bot", "Steam Discord Chat Bot")
	t.addAssetJSON2HTML()
	t.addAssetHighCharts()
	t.Link = config.C.DiscordBotInviteURL
	t.Regions = i18n.GetProdCCs(true)
	t.Commands = chatbot.CommandRegister

	returnTemplate(w, r, t)
}

type chatBotTemplate struct {
	globalTemplate
	Link     string
	Regions  []i18n.ProductCountryCode
	Commands []chatbot.Command
}

//goland:noinspection RegExpRedundantEscape
var (
	regexpChatLegacy      = regexp.MustCompile(`\{\w+\}\??`)
	regexpChatLegacyStart = regexp.MustCompile(`^[.!]\w+`)
)

func (cbt chatBotTemplate) RenderLegacy(input string) (interaction interactions.Interaction) {

	interaction.ID = regexpChatLegacyStart.FindString(input)

	for _, v := range regexpChatLegacy.FindAllString(input, -1) {

		interaction.Options = append(interaction.Options, interactions.InteractionOption{
			Name:     strings.Trim(v, "{}?"),
			Required: !strings.Contains(v, "?"),
		})
	}

	return interaction
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

func chatbotChartAjaxHandler(w http.ResponseWriter, r *http.Request) {

	var hc influxHelper.HighChartsJSON

	callback := func() (interface{}, error) {

		// Requests
		builder := influxql.NewBuilder()
		builder.AddSelect(`sum("request")`, "sum_request")
		builder.SetFrom(influxHelper.InfluxGameDB, influxHelper.InfluxRetentionPolicyAllTime.String(), influxHelper.InfluxMeasurementChatBot.String())
		builder.AddWhere("time", ">", "now()-14d")
		builder.AddGroupByTime("1h")
		builder.SetFillNumber(0)

		resp, err := influxHelper.InfluxQuery(builder)
		if err != nil {
			log.Err(err.Error(), zap.String("query", builder.String()))
			return hc, err
		}

		var hc1 influxHelper.HighChartsJSON
		if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {
			hc1 = influxHelper.InfluxResponseToHighCharts(resp.Results[0].Series[0], true)
		}

		// Guilds
		builder = influxql.NewBuilder()
		builder.AddSelect(`max("guilds")`, "max_guilds")
		builder.SetFrom(influxHelper.InfluxGameDB, influxHelper.InfluxRetentionPolicyAllTime.String(), influxHelper.InfluxMeasurementChatBot.String())
		builder.AddWhere("time", ">", "now()-14d")
		builder.AddGroupByTime("1h")
		builder.SetFillPrevious()

		resp, err = influxHelper.InfluxQuery(builder)
		if err != nil {
			log.Err(err.Error(), zap.String("query", builder.String()))
			return hc, err
		}

		var hc2 influxHelper.HighChartsJSON
		if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {
			hc2 = influxHelper.InfluxResponseToHighCharts(resp.Results[0].Series[0], true)
		}

		//
		hc = influxHelper.HighChartsJSON{
			"sum_request": hc1["sum_request"],
			"max_guilds":  hc2["max_guilds"],
		}

		return hc, err
	}

	err := memcache.GetSetInterface(memcache.ItemChatbotCalls, &hc, callback)
	if err != nil {
		log.Err("GetSet memcache", zap.Error(err))
	}

	returnJSON(w, r, hc)
}
