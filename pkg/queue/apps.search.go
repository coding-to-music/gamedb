package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/elastic"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type AppsSearchMessage struct {
	App mongo.App `json:"app"`
}

var aliases = map[int][]string{
	813780:  {"aoe", "aoe2"},   // Age of Empires II: Definitive Edition
	221380:  {"aoe", "aoe2"},   // Age of Empires II (2013)
	105450:  {"aoe", "aoe3"},   // Age of Empires® III: Complete Collection
	1017900: {"aoe"},           // Age of Empires: Definitive Edition
	105430:  {"aoe"},           // Age of Empires Online
	346110:  {"ark"},           // ARK: Survival Evolved
	49520:   {"bl", "bl2"},     // Borderlands 2
	397540:  {"bl", "bl3"},     // Borderlands 3
	8980:    {"bl"},            // Borderlands GOTY
	10:      {"cs"},            // Counter-Strike
	730:     {"csgo", "cs:go"}, // Counter-Strike: Global Offensive
	240:     {"css"},           // Counter-Strike: Source
	570:     {"dota", "dota2"}, // Dota 2
	8500:    {"eve"},           // EVE Online
	4000:    {"gm"},            // Garry's Mod
	12100:   {"gta", "gta3"},   // Grand Theft Auto III
	12210:   {"gta", "gta4"},   // Grand Theft Auto IV: The Complete Edition
	271590:  {"gta5"},          // Grand Theft Auto V
	12120:   {"gta", "gtasa"},  // Grand Theft Auto: San Andreas
	12110:   {"gta", "gtavc"},  // Grand Theft Auto: Vice City
	70:      {"hl"},            // Half-Life
	546560:  {"hla"},           // Half-Life: Alyx
	220:     {"hl2"},           // Half-Life 2
	500:     {"l4d"},           // Left 4 Dead
	550:     {"l4d2"},          // Left 4 Dead 2
	238960:  {"poe"},           // Path of Exile
	218620:  {"pd2"},           // PAYDAY 2
	252950:  {"rl"},            // Rocket League
	1172620: {"sot"},           // Sea of Thieves
	3900:    {"civ", "civ4"},   // Sid Meier's Civilization IV
	8930:    {"civ", "civ5"},   // Sid Meier's Civilization V
	289070:  {"civ", "civ6"},   // Sid Meier's Civilization VI
	985890:  {"sor", "sor4"},   // Streets of Rage 4
	440:     {"tf2"},           // Team Fortress 2
	359550:  {"r6"},            // Tom Clancy's Rainbow Six Siege
}

func appsSearchHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := AppsSearchMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		app := elastic.App{}
		app.ID = payload.App.ID
		app.Name = payload.App.Name
		app.Players = payload.App.PlayerPeakWeek
		app.Icon = payload.App.Icon
		app.Followers = payload.App.GroupFollowers
		app.ReviewScore = payload.App.ReviewsScore
		app.Prices = payload.App.Prices
		app.Tags = payload.App.Tags
		app.Genres = payload.App.Genres
		app.Categories = payload.App.Categories
		app.Publishers = payload.App.Publishers
		app.Developers = payload.App.Developers
		app.Type = payload.App.Type
		app.Platforms = payload.App.Platforms

		if val, ok := aliases[payload.App.ID]; ok {
			app.Aliases = val
		}

		err = elastic.IndexApp(app)
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
}
