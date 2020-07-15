package chatbot

import (
	"html/template"
	"strings"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
)

type CommandAppPrice struct {
}

func (c CommandAppPrice) ID() string {
	return CAppPrice
}

func (CommandAppPrice) Regex() string {
	return `^[.|!]price\s?([a-zA-Z]{2})?\s(.*)`
}

func (CommandAppPrice) DisableCache() bool {
	return false
}

func (CommandAppPrice) Example() string {
	return ".price {region}? {game}"
}

func (CommandAppPrice) Description() template.HTML {

	var ccs []string
	for _, v := range i18n.GetProdCCs(true) {
		ccs = append(ccs, string(v.ProductCode))
	}

	//noinspection GoRedundantConversion
	return "Get the price of a game <small>(Allowed regions: " + template.HTML(strings.Join(ccs, ", ")) + ")</small>"
}

func (CommandAppPrice) Type() CommandType {
	return TypeGame
}

func (c CommandAppPrice) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	matches := RegexCache[c.Regex()].FindStringSubmatch(msg.Message.Content)

	apps, _, _, err := elasticsearch.SearchApps(1, 0, matches[2], false, false, false)
	if err != nil {
		return message, err
	} else if len(apps) == 0 {
		message.Content = "Game **" + matches[2] + "** not found"
		return message, nil
	}

	app := apps[0]

	var code = steamapi.ProductCCUS

	if matches[1] == "" {

		settings, err := mysql.GetChatBotSettings(msg.Author.ID)
		if err != nil {
			log.Err(err)
			return message, err
		}

		code = settings.ProductCode

	} else {
		matches[1] = strings.ToLower(matches[1])
		if steamapi.IsProductCC(matches[1]) {
			code = steamapi.ProductCC(matches[1])
		}
	}

	price := app.Prices.Get(code)

	if price.Exists {
		message.Content = "<@" + msg.Author.ID + ">, " + app.GetName() + " is **" + price.GetFinal() + "** for " + strings.ToUpper(string(code))
		return message, nil
	}

	message.Content = "<@" + msg.Author.ID + ">, " + app.GetName() + " has no price for " + strings.ToUpper(string(code))
	return message, nil
}
