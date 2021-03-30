package consumers

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

			if val, ok := err.(*elastic.Error); ok {

				switch val.Status {
				case 409:
					// Index conflict when two writes happen at the same time
					sendToRetryQueueWithDelay(message, time.Second)
					return
				case 404:
					// Row has not been created yet to update
					message.Ack()
					return
				}
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
	app.AchievementsAvg = mongoApp.AchievementsAverageCompletion
	app.AchievementsCount = mongoApp.AchievementsCount
	app.AchievementsIcons = mongoApp.Achievements
	app.Aliases = makeAppAliases(mongoApp.ID, mongoApp.Name)
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
	app.NameLC = strings.ToLower(mongoApp.Name)
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

// Mainly used for splitting words by syllable for abbreviations
var replacementMap = map[string]string{
	"battlefield":    "battle field",
	"battlegrounds":  "battle grounds",
	"borderlands":    "border lands",
	"civilization":   "civ",
	"cyberpunk":      "cyber punk",
	"payday":         "pay day",
	"playerunknowns": "player unknowns",
	"warframe":       "war frame",
	"simulator":      "sim",
	"fallout":        "fall out",
	"tabletop":       "table top",
	"40,000":         "40k", // Warhammer
}

// For aliases that can not be calculated
var aliasMap = map[int][]string{
	1172470: {"apex"},                  // Apex Legends
	261550:  {"mab2"},                  // Mount & Blade II: Bannerlord
	48700:   {"mab", "mabw"},           // Mount & Blade: Warband
	578080:  {"pubg"},                  // PLAYERUNKNOWN'S BATTLEGROUNDS
	359550:  {"r6"},                    // Tom Clancy's Rainbow Six Siege
	444200:  {"wot", "world of tanks"}, // World of Tanks Blitz
}

//goland:noinspection RegExpRedundantEscape
var (
	regexpVersionsAndRomanNumerals = regexp.MustCompile(`\b[IVX]{1,4}\b|\b[0-9]{1,2}\b`)

	regexpAppSuffix = regexp.MustCompile(strings.Join([]string{
		`\:\s`,           // Colon
		`\s\(`,           // Brackets
		`\s\w+ edition$`, // editions
		`\sgoty`,         // Game of the year
		`\sonline`,       // Online
	}, "|"))
)

func makeAppAliases(ID int, name string) (aliases []string) {

	// Add aliases
	if val, ok := aliasMap[ID]; ok {
		aliases = val
	}

	// Add variations
	for _, convertRomanToInt := range []bool{true, false} {
		for _, convertIntToRoman := range []bool{true, false} {
			for _, removeSymbols := range []bool{true, false} {
				for _, swapSymbols := range []bool{true, false} {
					for _, removeSuffixes := range []bool{true, false} {
						for _, removeSpaces := range []bool{true, false} {
							for _, splitSyllables := range []bool{true, false} {
								for _, removePrefixes := range []bool{true, false} {
									for _, removeNumbers := range []bool{true, false} {

										name2 := name

										if splitSyllables {
											words := strings.Split(name2, " ")
											for k, v := range words {
												if val, ok := replacementMap[strings.ToLower(v)]; ok {
													words[k] = val
												}
											}
											name2 = strings.Join(words, " ")
										}

										if removePrefixes {
											name2 = strings.TrimPrefix(name2, `the `)
											name2 = strings.TrimPrefix(name2, `Sid Meier's `)
											name2 = strings.TrimPrefix(name2, `Tom Clancy's `)
										}

										if removeSuffixes {
											name2 = regexpAppSuffix.Split(name2, 2)[0]
										}

										if removeSymbols {
											name2 = helpers.RegexNonAlphaNumericSpace.ReplaceAllString(name2, "")
										}

										if swapSymbols {
											name2 = helpers.RegexNonAlphaNumericSpace.ReplaceAllString(name2, " ")
										}

										// Swap roman numerals
										name2 = regexpVersionsAndRomanNumerals.ReplaceAllStringFunc(name2, func(part string) string {

											maxVersion := 30

											if convertRomanToInt {
												part = helpers.RegexSmallRomanOnly.ReplaceAllStringFunc(part, func(part string) string {
													i := roman.Arabic(part)
													if i > maxVersion {
														return part
													}
													return strconv.Itoa(i)
												})
											}

											if convertIntToRoman {
												part = regexpVersionsAndRomanNumerals.ReplaceAllStringFunc(part, func(part string) string {
													i, _ := strconv.Atoi(part)
													if i > maxVersion {
														return part
													}
													converted, err := roman.Roman(i)
													if err != nil {
														return part
													} else {
														return converted
													}
												})
											}

											return part
										})

										if removeSpaces {
											name2 = strings.ReplaceAll(name2, " ", "")
										}

										if removeNumbers {
											name2 = regexpVersionsAndRomanNumerals.ReplaceAllString(name2, "")
										}

										//
										aliases = append(aliases, uniqueWords(strings.TrimSpace(name2)))

										// Add abreviations
										if removeSymbols && !removeSpaces {
											aliases = append(aliases, makeAbbreviations(name2)...)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return helpers.UniqueString(aliases)
}

func makeAbbreviations(name string) []string {

	r1 := regexp.MustCompile(`\b[a-zA-Z]|\b\s[IVX]{1,4}\b|\b\s[0-9]{1,2}\b`) // With spaces
	r2 := regexp.MustCompile(`\b[a-zA-Z]|\b[IVX]{1,4}\b|\b[0-9]{1,2}\b`)     // Without spaces

	return []string{
		strings.Join(r1.FindAllString(name, -1), ""),
		strings.Join(r2.FindAllString(name, -1), ""),
	}
}

func uniqueWords(alias string) string {
	return strings.Join(helpers.UniqueString(strings.Split(alias, " ")), " ")
}
