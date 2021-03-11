package chatbot

import (
	"strings"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot/charts"
	"github.com/gamedb/gamedb/pkg/config"
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

func (CommandAppPrice) AllowDM() bool {
	return false
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

func (c CommandAppPrice) Slash() []*discordgo.ApplicationCommandOption {

	return []*discordgo.ApplicationCommandOption{
		{
			Name:        "game",
			Description: "The name or ID of the game",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
		{
			Name:        "region",
			Description: "The region code",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    false,
		},
	}
}

func (c CommandAppPrice) Output(_ string, region steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

	if inputs["game"] == "" {
		message.Content = "Missing game name"
		return message, nil
	}

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

	app := apps[0]
	price := app.Prices.Get(region)

	if price.Exists {
		message.Content = apps[0].GetName() + " is **" + price.GetFinal() + "** for " + strings.ToUpper(string(region))
	} else {
		message.Content = apps[0].GetName() + " currently has no price for " + strings.ToUpper(string(region))
	}

	message.Embed = &discordgo.MessageEmbed{
		Title:       app.GetName(),
		Description: apps[0].GetName() + " is **" + price.GetFinal() + "** for " + strings.ToUpper(string(region)),
		URL:         config.C.GlobalSteamDomain + app.GetPath(),
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: app.GetHeaderImage(), Width: 460, Height: 215},
		Footer:      getFooter(),
		Color:       greenHexDec,
		Image:       &discordgo.MessageEmbedImage{URL: charts.GetPriceChart(region, c.ID(), app.ID, "Price History")},
	}

	return message, nil
}
