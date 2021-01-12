package chatbot

import (
	"strings"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
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

func (CommandSettings) Description() string {
	return "Update or view a Global Steam Discord setting"
}

func (CommandSettings) Type() CommandType {
	return TypeOther
}

func (c CommandSettings) LegacyInputs(input string) map[string]string {

	matches := RegexCache[c.Regex()].FindStringSubmatch(input)

	return map[string]string{
		"setting": matches[1],
		"value":   matches[2],
	}
}

func (c CommandSettings) Slash() []interactions.InteractionOption {

	return []interactions.InteractionOption{
		{
			Name:        "setting",
			Description: "The setting to set/retrieve",
			Type:        interactions.InteractionOptionTypeString,
			Required:    true,
			Choices: []interactions.InteractionChoice{
				{"Region", "region"},
			},
		},
		{
			Name:        "value",
			Description: "The value to set, leave empty to retrieve the value",
			Type:        interactions.InteractionOptionTypeString,
			Required:    false,
		},
	}
}

func (c CommandSettings) Output(authorID string, _ steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

	var setting = strings.ToLower(inputs["setting"])
	var value = strings.ToLower(inputs["value"])
	var text string

	switch setting {
	case "region":

		if value == "gb" {
			value = "uk"
		}

		if steamapi.IsProductCC(value) {

			err = mysql.SetChatBotSettings(authorID, func(s *mysql.ChatBotSetting) { s.ProductCode = steamapi.ProductCC(value) })
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

	message.Content = text
	return message, nil
}
