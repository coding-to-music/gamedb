package queue

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/rabbit-go"
	roman "github.com/StefanSchroeder/Golang-Roman"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/olivere/elastic/v7"
	"go.uber.org/zap"
)

type AppsSearchMessage struct {
	App    *mongo.App             `json:"app"`
	AppID  int                    `json:"app_id"`
	Fields map[string]interface{} `json:"fields"` // Optional
}

func (m AppsSearchMessage) Queue() rabbit.QueueName {
	return QueueAppsSearch
}

func appsSearchHandler(message *rabbit.Message) {

	payload := AppsSearchMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	if len(payload.Fields) > 0 && payload.AppID > 0 {

		err = elasticsearch.UpdateDocumentFields(elasticsearch.IndexApps, strconv.Itoa(payload.AppID), payload.Fields)
		if err != nil {

			if val, ok := err.(*elastic.Error); ok && (val.Status == 404 || val.Status == 409) {
				log.Info("Updating missing Elastic row", zap.Error(err), zap.Int("app", payload.AppID))
				if val.Status == 404 {
					sendToRetryQueueWithDelay(message, time.Minute)
				} else {
					sendToRetryQueueWithDelay(message, time.Second)
				}
				return
			}

			log.Err("Saving to Elastic", zap.Error(err), zap.Int("app", payload.AppID))
			sendToRetryQueue(message)
			return
		}

		message.Ack()
		return
	}

	var mongoApp mongo.App

	if payload.AppID > 0 {

		mongoApp, err = mongo.GetApp(payload.AppID)
		if err != nil {
			log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
			sendToRetryQueue(message)
			return
		}

	} else if payload.App != nil {

		mongoApp = *payload.App

	} else {

		log.ErrS(message.Message.Body)
		sendToFailQueue(message)
		return
	}

	app := elasticsearch.App{}

	if mongoApp.PlayerPeakWeek > 0 && mongoApp.GroupFollowers > 0 {
		app.Aliases = makeAppAliases(mongoApp.ID, mongoApp.Name)
	}

	app.AchievementsAvg = mongoApp.AchievementsAverageCompletion
	app.AchievementsCount = mongoApp.AchievementsCount
	app.AchievementsIcons = mongoApp.Achievements
	app.Background = mongoApp.Background
	app.Categories = mongoApp.Categories
	app.Developers = mongoApp.Developers
	app.FollowersCount = mongoApp.GroupFollowers
	app.Genres = mongoApp.Genres
	app.GroupID = mongoApp.GroupID
	app.Icon = mongoApp.Icon
	app.ID = mongoApp.ID
	app.MicroTrailor = mongoApp.GetMicroTrailer()
	app.Name = mongoApp.Name
	app.Platforms = mongoApp.Platforms
	app.PlayersCount = mongoApp.PlayerPeakWeek
	app.Prices = mongoApp.Prices
	app.Publishers = mongoApp.Publishers
	app.ReleaseDateOriginal = mongoApp.ReleaseDate
	app.ReleaseDate = mongoApp.ReleaseDateUnix
	app.ReleaseDateRounded = time.Unix(mongoApp.ReleaseDateUnix, 10).Truncate(time.Hour * 24).Unix()
	app.ReviewScore = mongoApp.ReviewsScore
	app.ReviewsCount = mongoApp.ReviewsCount
	app.Tags = mongoApp.Tags
	app.Trend = mongoApp.PlayerTrend
	app.Type = mongoApp.Type
	app.WishlistAvg = mongoApp.WishlistAvgPosition
	app.WishlistCount = mongoApp.WishlistCount

	b, _ := json.Marshal(mongoApp.Movies)
	app.Movies = string(b)
	app.MoviesCount = len(mongoApp.Movies)

	b, _ = json.Marshal(mongoApp.Screenshots)
	app.Screenshots = string(b)
	app.ScreenshotsCount = len(mongoApp.Screenshots)

	err = elasticsearch.IndexApp(app)
	if err != nil {
		log.ErrS(err)
		sendToRetryQueue(message)
		return
	}

	message.Ack()
}

var aliasMap = map[int][]string{
	813780:  {"aoe", "aoe2"},           // Age of Empires II: Definitive Edition
	221380:  {"aoe", "aoe2"},           // Age of Empires II (2013)
	105450:  {"aoe", "aoe3"},           // Age of Empires® III: Complete Collection
	1017900: {"aoe", "aoede"},          // Age of Empires: Definitive Edition
	105430:  {"aoe", "aoeo"},           // Age of Empires Online
	1172470: {"apex"},                  // Apex Legends
	346110:  {"ark"},                   // ARK: Survival Evolved
	49520:   {"bl", "bl2"},             // Borderlands 2
	397540:  {"bl", "bl3"},             // Borderlands 3
	8980:    {"bl", "goty"},            // Borderlands GOTY
	1091500: {"cp", "cp2077", "cyber"}, // Cyberpunk 2077
	570:     {"dota", "dota2"},         // Dota 2
	8500:    {"eve", "eo"},             // EVE Online
	261550:  {"mab2"},                  // Mount & Blade II: Bannerlord
	48700:   {"mab", "mabw"},           // Mount & Blade: Warband
	24240:   {"pd", "pd1", "pdth"},     // PAYDAY: The Heist
	218620:  {"pd", "pd2"},             // PAYDAY 2
	578080:  {"pubg"},                  // PLAYERUNKNOWN'S BATTLEGROUNDS
	3900:    {"civ", "civ4"},           // Sid Meier's Civilization IV
	8930:    {"civ", "civ5"},           // Sid Meier's Civilization V
	289070:  {"civ", "civ6"},           // Sid Meier's Civilization VI
	359550:  {"r6"},                    // Tom Clancy's Rainbow Six Siege
	230410:  {"wf"},                    // Warframe
	444200:  {"wot"},                   // World of Tanks Blitz
}

//goland:noinspection RegExpRedundantEscape
var (
	regexpRoman                          = regexp.MustCompile(`[IVX]{1,4}|[0-9]{1,2}`)
	regexpFirstLetterAndRomansAndNumbers = regexp.MustCompile(`[IVX]{1,4}|[0-9]{1,2}|\b[a-zA-Z]`)
	regexpSplitOnEnding                  = regexp.MustCompile(`\s\(|\:\s`)
)

func makeAppAliases(ID int, name string) (aliases []string) {

	// Add aliases
	if val, ok := aliasMap[ID]; ok {
		aliases = val
	}

	// Add abreviations
	aliases = append(aliases, strings.Join(regexpFirstLetterAndRomansAndNumbers.FindAllString(helpers.RegexNonAlphaNumericSpace.ReplaceAllString(name, ""), -1), ""))

	// Add variations
	for _, convertRomanToInt := range []bool{true, false} {
		for _, convertIntToRoman := range []bool{true, false} {
			for _, removeSymbols := range []bool{true, false} {
				for _, removeEndings := range []bool{true, false} {

					name2 := name

					if removeEndings {
						name2 = regexpSplitOnEnding.Split(name2, 2)[0]
					}

					if removeSymbols {
						name2 = helpers.RegexNonAlphaNumericSpace.ReplaceAllString(name2, "")
					}

					// Swap roman numerals
					name2 = regexpRoman.ReplaceAllStringFunc(name2, func(part string) string {
						if convertRomanToInt {
							part = helpers.RegexSmallRomanOnly.ReplaceAllStringFunc(part, func(part string) string {
								return strconv.Itoa(roman.Arabic(part))
							})
						}
						if convertIntToRoman {
							part = regexpRoman.ReplaceAllStringFunc(part, func(part string) string {
								i, _ := strconv.Atoi(part)
								if i <= 20 {
									return part
								}
								return roman.Roman(i)
							})
						}

						return part
					})

					//
					aliases = append(aliases, name2)
				}
			}
		}
	}

	return helpers.UniqueString(aliases)
}
