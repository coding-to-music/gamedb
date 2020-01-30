package mongo

import (
	"strconv"
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	specialImageBase = "https://steamcommunity-a.akamaihd.net/public/images/badges/"
	eventImageBase   = "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/items/"
)

type PlayerBadge struct {
	AppID               int       `bson:"app_id"`
	AppName             string    `bson:"app_name"`
	BadgeCompletionTime time.Time `bson:"badge_completion_time"`
	BadgeFoil           bool      `bson:"badge_foil"`
	BadgeIcon           string    `bson:"badge_icon"`
	BadgeID             int       `bson:"badge_id"`
	BadgeItemID         int64     `bson:"-"`
	BadgeLevel          int       `bson:"badge_level"`
	BadgeName           string    `bson:"-"`
	BadgeScarcity       int       `bson:"badge_scarcity"`
	BadgeXP             int       `bson:"badge_xp"`
	PlayerID            int64     `bson:"player_id"`
	PlayerName          string    `bson:"player_name"`
	PlayerIcon          string    `bson:"player_icon"`
}

func (badge PlayerBadge) BSON() bson.D {

	return bson.D{
		{"_id", badge.getKey()},
		{"app_id", badge.AppID},
		{"app_name", badge.AppName},
		{"badge_completion_time", badge.BadgeCompletionTime},
		{"badge_foil", badge.BadgeFoil},
		{"badge_icon", badge.BadgeIcon},
		{"badge_id", badge.BadgeID},
		{"badge_level", badge.BadgeLevel},
		{"badge_scarcity", badge.BadgeScarcity},
		{"badge_xp", badge.BadgeXP},
		{"player_id", badge.PlayerID},
		{"player_icon", badge.PlayerIcon},
		{"player_name", badge.PlayerName},
	}
}

func (badge PlayerBadge) getKey() string {
	return strconv.FormatInt(badge.PlayerID, 10) + "-" + strconv.Itoa(badge.AppID) + "-" + strconv.Itoa(badge.BadgeID) + "-" + strconv.FormatBool(badge.BadgeFoil)
}

func (badge PlayerBadge) GetPlayerCommunityLink() string {

	var dir string
	if badge.IsSpecial() {
		dir = "gamecards"
	} else {
		dir = "badges"
	}

	return "https://steamcommunity.com/profiles/" + strconv.FormatInt(badge.PlayerID, 10) + "/" + dir + "/" + strconv.Itoa(badge.BadgeID)
}

func (badge PlayerBadge) IsSpecial() bool {
	return badge.AppID == 0
}

func (badge PlayerBadge) GetUniqueID() int {
	if badge.IsSpecial() {
		return badge.BadgeID
	}
	return badge.AppID
}

func (badge PlayerBadge) GetName() string {
	return badge.BadgeName
}

func (badge PlayerBadge) GetPath() string {
	if badge.IsSpecial() {
		return "/badges/" + strconv.Itoa(badge.BadgeID) + "/" + slug.Make(badge.BadgeName)
	}
	return "/badges/" + strconv.Itoa(badge.AppID) + "/" + slug.Make(badge.BadgeName)
}

func (badge PlayerBadge) GetPlayerPath() string {
	return helpers.GetPlayerPath(badge.PlayerID, badge.PlayerName) + "#badges"
}

func (badge PlayerBadge) GetTimeFormatted() string {
	return badge.BadgeCompletionTime.Format(helpers.DateYearTime)
}

func (badge PlayerBadge) GetAppPath() string {
	return helpers.GetAppPath(badge.AppID, badge.AppName)
}

func (badge PlayerBadge) GetAppName() string {
	if badge.AppID == 0 {
		return "Special Badge"
	}
	return helpers.GetAppName(badge.AppID, badge.AppName)
}

func (badge PlayerBadge) GetBadgeIcon() string {

	if badge.BadgeIcon == "" {
		return helpers.DefaultAppIcon
	}

	if strings.HasPrefix(badge.BadgeIcon, "http") {
		return badge.BadgeIcon
	}

	if badge.AppID > 0 {
		return eventImageBase + "/" + strconv.Itoa(badge.AppID) + "/" + badge.BadgeIcon + ".png"
	}

	return specialImageBase + badge.BadgeIcon
}

func (badge PlayerBadge) GetPlayerIcon() string {
	return helpers.GetPlayerAvatar(badge.PlayerIcon)
}

func UpdatePlayerBadges(badges []PlayerBadge) (err error) {

	if len(badges) == 0 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	var writes []mongo.WriteModel
	for _, badge := range badges {

		write := mongo.NewReplaceOneModel()
		write.SetFilter(bson.M{"_id": badge.getKey()})
		write.SetReplacement(badge.BSON())
		write.SetUpsert(true)

		writes = append(writes, write)
	}

	c := client.Database(MongoDatabase).Collection(CollectionPlayerBadges.String())

	_, err = c.BulkWrite(ctx, writes, options.BulkWrite())

	return err
}

func GetPlayerEventBadges(offset int64, filter bson.D) (badges []PlayerBadge, err error) {
	return getPlayerBadges(offset, 100, filter, bson.D{{"badge_completion_time", -1}}, nil)
}

func GetBadgePlayers(offset int64, filter bson.D) (badges []PlayerBadge, err error) {
	return getPlayerBadges(offset, 100, filter, bson.D{{"badge_level", -1}, {"badge_completion_time", 1}}, nil)
}

func getPlayerBadges(offset int64, limit int64, filter bson.D, sort bson.D, projection bson.M) (badges []PlayerBadge, err error) {

	cur, ctx, err := Find(CollectionPlayerBadges, offset, limit, sort, filter, projection, nil)
	if err != nil {
		return badges, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var badge PlayerBadge
		err := cur.Decode(&badge)
		if err != nil {
			log.Err(err, badge.getKey())
		} else {
			badges = append(badges, badge)
		}
	}

	return badges, cur.Err()
}

var GlobalBadges = map[int]PlayerBadge{
	1:       {BadgeID: 1, BadgeIcon: "02_years/steamyears1002_80.png", BadgeName: "Years of Service"},
	2:       {BadgeID: 2, BadgeIcon: "01_community/community03_80.png", BadgeName: "Community Ambassador"},
	3:       {BadgeID: 3, BadgeIcon: "03_potato/potato03_80.png", BadgeName: "The Potato Sack"},
	4:       {BadgeID: 4, BadgeIcon: "04_treasurehunt/treasurehunt03_80.png", BadgeName: "The Great Steam Treasure Hunt"},
	5:       {BadgeID: 5, BadgeIcon: "05_summer2011/tickets80.png", BadgeName: "Steam Summer Camp"},
	6:       {BadgeID: 6, BadgeIcon: "06_winter2011/coal03_80.png", BadgeName: "Steam Holiday Sale 2011"},
	7:       {BadgeID: 7, BadgeIcon: "07_summer2012/Summer2012_stage3_80.png", BadgeName: "Steam Summer Sale 2012"},
	8:       {BadgeID: 8, BadgeIcon: "08_winter2012/winter2012_badge80.png", BadgeName: "Steam Holiday Sale 2012"},
	9:       {BadgeID: 9, BadgeIcon: "09_communitytranslator/translator_level4_80.png", BadgeName: "Steam Community Translator"},
	10:      {BadgeID: 10, BadgeIcon: "generic/CommunityModerator_80.png", BadgeName: "Steam Community Moderator"},
	11:      {BadgeID: 11, BadgeIcon: "generic/ValveEmployee_80.png", BadgeName: "Valve Employee"},
	12:      {BadgeID: 12, BadgeIcon: "generic/GameDeveloper_80.png", BadgeName: "Steamworks Developer"},
	13:      {BadgeID: 13, BadgeIcon: "13_gamecollector/25000_80.png", BadgeName: "Owned Games"},
	14:      {BadgeID: 14, BadgeIcon: "generic/TradingCardBeta_80.png", BadgeName: "Trading Card Beta Tester"},
	15:      {BadgeID: 15, BadgeIcon: "15_hwbeta/hwbeta03_80.png", BadgeName: "Steam Hardware Beta"},
	16:      {BadgeID: 16, BadgeIcon: "16_summer2014/team_red.png", BadgeName: "Steam Summer Adventure 2014 - Red Team"},
	17:      {BadgeID: 17, BadgeIcon: "16_summer2014/team_blue.png", BadgeName: "Steam Summer Adventure 2014 - Blue Team"},
	18:      {BadgeID: 18, BadgeIcon: "16_summer2014/team_pink.png", BadgeName: "Steam Summer Adventure 2014 - Pink Team"},
	19:      {BadgeID: 19, BadgeIcon: "16_summer2014/team_green.png", BadgeName: "Steam Summer Adventure 2014 - Green Team"},
	20:      {BadgeID: 20, BadgeIcon: "16_summer2014/team_purple.png", BadgeName: "Steam Summer Adventure 2014 - Purple Team"},
	21:      {BadgeID: 21, BadgeIcon: "21_auction/winner_80.png?v=2", BadgeName: "Auction Participant/Winner"},
	22:      {BadgeID: 22, BadgeIcon: "22_golden/owner_80.png", BadgeName: "2014 Holiday Profile Recipient"},
	23:      {BadgeID: 23, BadgeIcon: "23_towerattack/wormhole.png", BadgeName: "Monster Summer"},
	24:      {BadgeID: 24, BadgeIcon: "24_winter2015_arg_red_herring/red_herring.png", BadgeName: "Red Herring"},
	25:      {BadgeID: 25, BadgeIcon: "25_steamawardnominations/level04_80.png", BadgeName: "Steam Awards Nomination Committee 2016"},
	26:      {BadgeID: 26, BadgeIcon: "26_summer2017_sticker/completionist.png", BadgeName: "Sticker Completionist"},
	27:      {BadgeID: 27, BadgeIcon: "27_steamawardnominations/level04_80.png", BadgeName: "Steam Awards Nomination Committee 2017"},
	28:      {BadgeID: 28, BadgeIcon: "28_springcleaning2018/gold_80.png", BadgeName: "Spring Cleaning Event 2018"},
	29:      {BadgeID: 29, BadgeIcon: "29_salien/6_80.png", BadgeName: "Salien"},
	30:      {BadgeID: 30, BadgeIcon: "generic/RetiredModerator_80.png", BadgeName: "Retired Community Moderator"},
	31:      {BadgeID: 31, BadgeIcon: "30_steamawardnominations/level04_80.png", BadgeName: "Steam Awards Nomination Committee 2018"},
	32:      {BadgeID: 32, BadgeIcon: "generic/ValveEmployee_80.png", BadgeName: "Valve Moderator"},
	33:      {BadgeID: 33, BadgeIcon: "33_cozycottage2018/1000000_80.png", BadgeName: "Winter 2018 Knick-Knack Collector"},
	34:      {BadgeID: 34, BadgeIcon: "34_lny2019/10_80.png", BadgeName: "Lunar New Year 2019"},
	35:      {BadgeID: 34, BadgeIcon: "https://steamcommunity-a.akamaihd.net/economy/image/-9a81dlWLwJ2UUGcVs_nsVtzdOEdtWwKGZZLQHTxH5rd9eDAjcFyv45SRYAFMIcKL_PArgVSL403ulRUWEndVKv8h56EAgQkalZSsuOnegRm1aqed2oStIXlkIHez6aiNe6CkzIAuJcgiLGU8I6kjgz6ux07-Ytsxtc/96fx96f", BadgeName: "Lunar New Year 2019 Golden Profile"},
	36:      {BadgeID: 36, BadgeIcon: "36_springcleaning2019/gold_80x80.png", BadgeName: "Spring Cleaning Event 2019"},
	37:      {BadgeID: 37, BadgeIcon: "37_summer2019/level1000000_80.png", BadgeName: "Steam Grand Prix 2019"},
	38:      {BadgeID: 38, BadgeIcon: "37_summer2019/hare_gold_80.png", BadgeName: "Steam Grand Prix 2019 - Team Hare"},
	39:      {BadgeID: 39, BadgeIcon: "37_summer2019/tortoise_gold_80.png", BadgeName: "Steam Grand Prix 2019 - Team Tortoise"},
	40:      {BadgeID: 40, BadgeIcon: "37_summer2019/corgi_gold_80.png", BadgeName: "Steam Grand Prix 2019 - Team Corgi"},
	41:      {BadgeID: 41, BadgeIcon: "37_summer2019/cockatiel_gold_80.png", BadgeName: "Steam Grand Prix 2019 - Team Cockatiel"},
	42:      {BadgeID: 42, BadgeIcon: "37_summer2019/pig_gold_80.png", BadgeName: "Steam Grand Prix 2019 - Team Pig"},
	43:      {BadgeID: 43, BadgeIcon: "43_steamawardnominations/level04_80.png", BadgeName: "Steam Awards Nomination Committee 2019"},
	44:      {BadgeID: 44, BadgeIcon: "44_winter2019/level15_80.png", BadgeName: "Winter Sale 2019"},
	45:      {BadgeID: 45, BadgeIcon: "45_steamville2019/key_to_city_80.png", BadgeName: "Steamville 2019"},
	46:      {BadgeID: 46, BadgeIcon: "46_lny2020/10_80.png", BadgeName: "Lunar New Year 2020"},
	245070:  {BadgeID: 1, AppID: 245070, BadgeIcon: "30a5de112a3512269cbc3d428fad3b9c507c56ba", BadgeName: "2013: Summer Getaway"},
	267420:  {BadgeID: 1, AppID: 267420, BadgeIcon: "e041163b0c4d5cba61fb54612973612636cdd970", BadgeName: "2013: Holdiay Sale"},
	303700:  {BadgeID: 1, AppID: 303700, BadgeIcon: "b3c3fa2821b32ce6bcc127e5ee3ec47845c35308", BadgeName: "2014: Summer Adventure"},
	335590:  {BadgeID: 1, AppID: 335590, BadgeIcon: "b1c504dfaf4d073e5cf9c2d7d48c55c9cf11b7d3", BadgeName: "2014: Holdiay Sale"},
	368020:  {BadgeID: 1, AppID: 368020, BadgeIcon: "49715c47e076456e0aacec76a5a0d714cadea951", BadgeName: "2015: Monster Summer Sale"},
	425280:  {BadgeID: 1, AppID: 425280, BadgeIcon: "3442d0c36e5d549abf29872c9aec9f6e4364d23f", BadgeName: "2015: Holdiay Sale"},
	480730:  {BadgeID: 1, AppID: 480730, BadgeIcon: "6b1280c07eedafdb3d9cac282f82da4365b9c98d", BadgeName: "2016: Summer Sale"},
	566020:  {BadgeID: 1, AppID: 566020, BadgeIcon: "604be0b040a1117a5b8b5579b3c6ec25e540f458", BadgeName: "2016: Steam Awards"},
	639900:  {BadgeID: 1, AppID: 639900, BadgeIcon: "9dd59323d14eb5dba94328db80e27caaee4c29ea", BadgeName: "2017: Summer Sale"},
	762800:  {BadgeID: 1, AppID: 762800, BadgeIcon: "0a10b3b3725de8f72cb48fd94daff296cc3dfe52", BadgeName: "2017: Steam Awards"},
	876740:  {BadgeID: 1, AppID: 876740, BadgeIcon: "9c677484f7f148045189a9dabe7efdf733e9e1f1", BadgeName: "2018: Intergalactic Summer"},
	991980:  {BadgeID: 1, AppID: 991980, BadgeIcon: "3c96df81a7f82f23b68356c51733793cdece8f63", BadgeName: "2018: Winter Sale"},
	1195670: {BadgeID: 1, AppID: 1195670, BadgeIcon: "581a14e34100d7f4955ceff9365a7e40a89b57c8", BadgeName: "2019: Winter Sale"},
}
