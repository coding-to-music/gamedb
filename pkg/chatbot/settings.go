package chatbot

import (
	"html/template"
	"strings"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
)

type CommandSettings struct {
}

func (c CommandSettings) ID() string {
	return CSettings
}

func (CommandSettings) Regex() string {
	return `^[.|!]set (region) ([a-zA-Z]{2})`
}

func (CommandSettings) DisableCache() bool {
	return false
}

func (CommandSettings) PerProdCode() bool {
	return false
}

func (CommandSettings) Example() string {
	return ".set region uk"
}

func (CommandSettings) Description() template.HTML {

	var ccs []string
	for _, v := range i18n.GetProdCCs(true) {
		ccs = append(ccs, string(v.ProductCode))
	}

	//noinspection GoRedundantConversion
	return "Set your region for price commands <small>(Allowed regions: " + template.HTML(strings.Join(ccs, ", ")) + ")</small>"
}

func (CommandSettings) Type() CommandType {
	return TypeOther
}

func (c CommandSettings) Output(msg *discordgo.MessageCreate, _ steamapi.ProductCC) (message discordgo.MessageSend, err error) {

	matches := RegexCache[c.Regex()].FindStringSubmatch(msg.Message.Content)

	var setting = strings.ToLower(matches[1])
	var value = strings.ToLower(matches[2])
	var text string

	switch setting {
	case "region":

		if value == "gb" {
			value = "uk"
		}

		if steamapi.IsProductCC(value) {

			err = mysql.SetChatBotSettings(msg.Author.ID, func(s *mysql.ChatBotSetting) { s.ProductCode = steamapi.ProductCC(value) })
			if err != nil {
				log.ErrS(err)
				return
			}
			text = "Region set to " + strings.ToUpper(value) + " (" + string(i18n.GetProdCC(steamapi.ProductCC(value)).CurrencyCode) + ")"
		} else {
			text = "Invalid region, see .help"
		}
	default:
		text = "Invalid setting, see .help"
	}

	message.Content = "<@" + msg.Author.ID + ">, " + text
	return message, nil
}
