package chatbot

import (
	"strings"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/i18n"
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

func (CommandAppPrice) PerProdCode() bool {
	return true
}

func (CommandAppPrice) Example() string {
	return ".price {region}? {game}"
}

func (CommandAppPrice) Description() string {
	return "Retrieve information about the price of a game"
}

func (CommandAppPrice) Type() CommandType {
	return TypeGame
}

func (c CommandAppPrice) LegacyInputs(input string) map[string]string {

	matches := RegexCache[c.Regex()].FindStringSubmatch(input)

	return map[string]string{
		"region": matches[1],
		"game":   matches[2],
	}
}

func (c CommandAppPrice) Slash() []interactions.InteractionOption {

	return []interactions.InteractionOption{
		{
			Name:        "game",
			Description: "The name or ID of the game",
			Type:        interactions.InteractionOptionTypeString,
			Required:    true,
		},
		{
			Name:        "region",
			Description: "The region code",
			Type:        interactions.InteractionOptionTypeString,
			Required:    false,
		},
	}
}

func (c CommandAppPrice) Output(_ string, region steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

	apps, err := elasticsearch.SearchAppsSimple(1, inputs["game"])
	if err != nil {
		return message, err
	} else if len(apps) == 0 {
		message.Content = "Game **" + inputs["game"] + "** not found on Steam"
		return message, nil
	}

	if inputs["region"] != "" {
		val, ok := i18n.ProductCountryCodes[steamapi.ProductCC(strings.ToLower(inputs["region"]))]
		if ok {
			if val.Enabled {
				region = val.ProductCode
			} else {
				message.Content = "We are not currently tracking " + strings.ToUpper(inputs["region"])
				return message, nil
			}
		} else {
			message.Content = "Invalid region: " + strings.ToUpper(inputs["region"])
			return message, nil
		}
	}

	price := apps[0].Prices.Get(region)

	if price.Exists {
		message.Content = apps[0].GetName() + " is **" + price.GetFinal() + "** for " + strings.ToUpper(string(region))
		return message, nil
	}

	message.Content = apps[0].GetName() + " has no price for " + strings.ToUpper(string(region))
	return message, nil
}
