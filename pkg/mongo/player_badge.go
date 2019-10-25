package mongo

import (
	"errors"
	"html/template"
	"sort"
	"strconv"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gosimple/slug"
	. "go.mongodb.org/mongo-driver/bson"
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

func (pb PlayerBadge) BSON() (ret interface{}) {

	return M{
		"_id":                   pb.getKey(),
		"app_id":                pb.AppID,
		"app_name":              pb.AppName,
		"badge_completion_time": pb.BadgeCompletionTime,
		"badge_foil":            pb.BadgeFoil,
		"badge_icon":            pb.BadgeIcon,
		"badge_id":              pb.BadgeID,
		"badge_level":           pb.BadgeLevel,
		"badge_scarcity":        pb.BadgeScarcity,
		"badge_xp":              pb.BadgeXP,
		"player_id":             pb.PlayerID,
		"player_icon":           pb.PlayerIcon,
		"player_name":           pb.PlayerName,
	}
}

func (pb PlayerBadge) getKey() string {
	return strconv.FormatInt(pb.PlayerID, 10) + "-" + strconv.Itoa(pb.AppID) + "-" + strconv.Itoa(pb.BadgeID) + "-" + strconv.FormatBool(pb.BadgeFoil)
}

func (pb PlayerBadge) IsSpecial() bool {
	return pb.AppID == 0
}

func (pb PlayerBadge) GetUniqueID() int {
	if pb.IsSpecial() {
		return pb.BadgeID
	}
	return pb.AppID
}

func (pb PlayerBadge) GetName() string {
	return pb.BadgeName
}

func (pb PlayerBadge) GetPath() string {
	if pb.IsSpecial() {
		return "/badges/" + strconv.Itoa(pb.BadgeID) + "/" + slug.Make(pb.BadgeName)
	}
	return "/badges/" + strconv.Itoa(pb.AppID) + "/" + slug.Make(pb.BadgeName)
}

func (pb PlayerBadge) GetPlayerPath() string {
	return helpers.GetPlayerPath(pb.PlayerID, pb.PlayerName) + "#badges"
}

func (pb PlayerBadge) GetTimeFormatted() string {
	return pb.BadgeCompletionTime.Format(helpers.DateYearTime)
}

func (pb PlayerBadge) GetAppPath() string {
	return helpers.GetAppPath(pb.AppID, pb.AppName)
}

func (pb PlayerBadge) GetAppName() string {
	if pb.AppID == 0 {
		return "Special Badge"
	}
	return helpers.GetAppName(pb.AppID, pb.AppName)
}

func (pb PlayerBadge) GetBadgeIcon() string {

	if pb.BadgeIcon == "" {
		return helpers.DefaultAppIcon
	}

	if pb.AppID > 0 {
		return eventImageBase + "/" + strconv.Itoa(pb.AppID) + "/" + pb.BadgeIcon + ".png"
	}

	return specialImageBase + pb.BadgeIcon
}

func (pb PlayerBadge) GetPlayerIcon() string {
	return helpers.GetPlayerAvatar(pb.PlayerIcon)
}

func (pb PlayerBadge) GetSpecialPlayers() (int64, error) {
	return CountDocuments(CollectionPlayerBadges, pb.specialPlayersCountFilter(), 60*60*24*24)
}

func (pb PlayerBadge) SetSpecialPlayers() error {
	return SetCountDocuments(CollectionPlayerBadges, pb.specialPlayersCountFilter(), 60*60*24*24)
}

func (pb PlayerBadge) specialPlayersCountFilter() D {
	return D{{"app_id", 0}, {"badge_id", pb.BadgeID}}
}

// Cached
func (pb PlayerBadge) GetSpecialFirsts() (ret template.HTML, err error) {

	var item = helpers.MemcacheBadgeSpecialFirsts(pb.BadgeID)
	var firsts []PlayerBadge

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &firsts, func() (interface{}, error) {
		return pb.getSpecialFirsts()
	})

	if len(firsts) > 1 {
		return template.HTML(strconv.Itoa(len(firsts))) + " joint firsts", nil
	}

	sort.Slice(firsts, func(i, j int) bool {
		return firsts[i].PlayerName < firsts[j].PlayerName
	})

	for k, v := range firsts {
		ret += "<a href=" + template.HTML(v.GetPlayerPath()) + ">" + template.HTML(v.PlayerName) + "</a>"
		if k < len(firsts)-1 {
			ret += " / "
		}
	}

	return ret, err
}

// Not cached
func (pb PlayerBadge) getSpecialFirsts() (playerBadges []PlayerBadge, err error) {

	var max PlayerBadge

	err = FindOne(
		CollectionPlayerBadges,
		D{{"app_id", 0}, {"badge_id", pb.BadgeID}},
		D{{"badge_level", -1}, {"badge_completion_time", 1}},
		M{"badge_level": 1, "badge_completion_time": 1},
		&max,
	)

	return getBadges(
		0,
		2, // We don't show names if above 1
		D{{"app_id", 0}, {"badge_id", pb.BadgeID}, {"badge_level", max.BadgeLevel}, {"badge_completion_time", max.BadgeCompletionTime}},
		D{{"badge_completion_time", -1}},
		M{"badge_level": 1, "_id": -1, "player_id": 1, "player_name": 1},
	)
}

func (pb PlayerBadge) SetSpecialFirsts() (err error) {

	item := helpers.MemcacheBadgeSpecialFirsts(pb.BadgeID)

	firsts, err := pb.getSpecialFirsts()
	if err != nil {
		return err
	}

	return helpers.GetMemcache().SetInterface(item.Key, firsts, 60*60*24)
}

func (pb PlayerBadge) GetEventPlayers() (int64, error) {
	return CountDocuments(CollectionPlayerBadges, pb.eventPlayersCountFilter(), 60*60*24*24)
}

func (pb PlayerBadge) SetEventPlayers() error {
	return SetCountDocuments(CollectionPlayerBadges, pb.eventPlayersCountFilter(), 60*60*24*24)
}

func (pb PlayerBadge) eventPlayersCountFilter() D {
	return D{{"app_id", pb.AppID}, {"badge_id", M{"$gt": 0}}}
}

// Cached
func (pb PlayerBadge) GetEventMax() (max PlayerBadge, err error) {

	var item = helpers.MemcacheBadgeMaxEvent(pb.AppID)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &max, func() (interface{}, error) {
		return pb.getEventFirst(false)
	})

	return max, err
}

func (pb PlayerBadge) SetEventMax() (err error) {

	item := helpers.MemcacheBadgeMaxEvent(pb.AppID)

	max, err := pb.getEventFirst(false)
	if err != nil {
		return err
	}

	return helpers.GetMemcache().SetInterface(item.Key, max, 60*60*24)
}

// Cached
func (pb PlayerBadge) GetEventMaxFoil() (max PlayerBadge, err error) {

	var item = helpers.MemcacheBadgeMaxEventFoil(pb.AppID)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &max, func() (interface{}, error) {
		return pb.getEventFirst(true)
	})

	return max, err
}

func (pb PlayerBadge) SetEventFoilMax() (err error) {

	item := helpers.MemcacheBadgeMaxEventFoil(pb.AppID)

	max, err := pb.getEventFirst(true)
	if err != nil {
		return err
	}

	return helpers.GetMemcache().SetInterface(item.Key, max, 60*60*24)
}

func (pb PlayerBadge) getEventFirst(foil bool) (max PlayerBadge, err error) {

	operation := func() (err error) {

		err = FindOne(
			CollectionPlayerBadges,
			D{{"app_id", pb.AppID}, {"badge_id", M{"$gt": 0}}, {"badge_foil", foil}},
			D{{"badge_level", -1}, {"badge_completion_time", 1}},
			M{"badge_level": 1, "_id": -1, "player_id": 1, "player_name": 1},
			&max,
		)

		if max.BadgeLevel < 2 {
			return errors.New("mongo returned wrong result")
		}

		return err
	}

	policy := backoff.NewExponentialBackOff()
	policy.InitialInterval = time.Second

	err = backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Info(err, t) })
	if err != nil {
		log.Critical(err)
	}

	return max, err
}

func UpdatePlayerBadges(badges []PlayerBadge) (err error) {

	if badges == nil || len(badges) == 0 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	var writes []mongo.WriteModel
	for _, badge := range badges {

		write := mongo.NewReplaceOneModel()
		write.SetFilter(M{"_id": badge.getKey()})
		write.SetReplacement(badge.BSON())
		write.SetUpsert(true)

		writes = append(writes, write)
	}

	c := client.Database(MongoDatabase).Collection(CollectionPlayerBadges.String())

	_, err = c.BulkWrite(ctx, writes, options.BulkWrite())

	return err
}

func GetPlayerEventBadges(offset int64, filter D) (badges []PlayerBadge, err error) {
	return getBadges(offset, 100, filter, D{{"badge_completion_time", -1}}, nil)
}

func GetBadgePlayers(offset int64, filter D) (badges []PlayerBadge, err error) {
	return getBadges(offset, 100, filter, D{{"badge_level", -1}, {"badge_completion_time", 1}}, nil)
}

func getBadges(offset int64, limit int64, filter D, sort D, projection M) (badges []PlayerBadge, err error) {

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
		}
		badges = append(badges, badge)
	}

	return badges, cur.Err()
}

var Badges = map[int]PlayerBadge{
	1:      {BadgeID: 1, BadgeIcon: "02_years/steamyears1002_80.png", BadgeName: "Years of Service"},
	2:      {BadgeID: 2, BadgeIcon: "01_community/community03_80.png", BadgeName: "Community Ambassador"},
	3:      {BadgeID: 3, BadgeIcon: "03_potato/potato03_80.png", BadgeName: "The Potato Sack"},
	4:      {BadgeID: 4, BadgeIcon: "04_treasurehunt/treasurehunt03_80.png", BadgeName: "The Great Steam Treasure Hunt"},
	5:      {BadgeID: 5, BadgeIcon: "05_summer2011/tickets80.png", BadgeName: "Steam Summer Camp"},
	6:      {BadgeID: 6, BadgeIcon: "06_winter2011/coal03_80.png", BadgeName: "Steam Holiday Sale 2011"},
	7:      {BadgeID: 7, BadgeIcon: "07_summer2012/Summer2012_stage3_80.png", BadgeName: "Steam Summer Sale 2012"},
	8:      {BadgeID: 8, BadgeIcon: "08_winter2012/winter2012_badge80.png", BadgeName: "Steam Holiday Sale 2012"},
	9:      {BadgeID: 9, BadgeIcon: "09_communitytranslator/translator_level4_80.png", BadgeName: "Steam Community Translator"},
	10:     {BadgeID: 10, BadgeIcon: "generic/CommunityModerator_80.png", BadgeName: "Steam Community Moderator"},
	11:     {BadgeID: 11, BadgeIcon: "generic/ValveEmployee_80.png", BadgeName: "Valve Employee"},
	12:     {BadgeID: 12, BadgeIcon: "generic/GameDeveloper_80.png", BadgeName: "Steamworks Developer"},
	13:     {BadgeID: 13, BadgeIcon: "13_gamecollector/25000_80.png", BadgeName: "Owned Games"},
	14:     {BadgeID: 14, BadgeIcon: "generic/TradingCardBeta_80.png", BadgeName: "Trading Card Beta Tester"},
	15:     {BadgeID: 15, BadgeIcon: "15_hwbeta/hwbeta03_80.png", BadgeName: "Steam Hardware Beta"},
	16:     {BadgeID: 16, BadgeIcon: "16_summer2014/team_red.png", BadgeName: "Steam Summer Adventure 2014 - Red Team"},
	17:     {BadgeID: 17, BadgeIcon: "16_summer2014/team_blue.png", BadgeName: "Steam Summer Adventure 2014 - Blue Team"},
	18:     {BadgeID: 18, BadgeIcon: "16_summer2014/team_pink.png", BadgeName: "Steam Summer Adventure 2014 - Pink Team"},
	19:     {BadgeID: 19, BadgeIcon: "16_summer2014/team_green.png", BadgeName: "Steam Summer Adventure 2014 - Green Team"},
	20:     {BadgeID: 20, BadgeIcon: "16_summer2014/team_purple.png", BadgeName: "Steam Summer Adventure 2014 - Purple Team"},
	21:     {BadgeID: 21, BadgeIcon: "21_auction/winner_80.png?v=2", BadgeName: "Auction Participant/Winner"},
	22:     {BadgeID: 22, BadgeIcon: "22_golden/owner_80.png", BadgeName: "2014 Holiday Profile Recipient"},
	23:     {BadgeID: 23, BadgeIcon: "23_towerattack/wormhole.png", BadgeName: "Monster Summer"},
	24:     {BadgeID: 24, BadgeIcon: "24_winter2015_arg_red_herring/red_herring.png", BadgeName: "Red Herring"},
	25:     {BadgeID: 25, BadgeIcon: "25_steamawardnominations/level04_80.png", BadgeName: "Steam Awards Nomination Committee 2016"},
	26:     {BadgeID: 26, BadgeIcon: "26_summer2017_sticker/completionist.png", BadgeName: "Sticker Completionist"},
	27:     {BadgeID: 27, BadgeIcon: "27_steamawardnominations/level04_80.png", BadgeName: "Steam Awards Nomination Committee 2017"},
	28:     {BadgeID: 28, BadgeIcon: "28_springcleaning2018/gold_80.png", BadgeName: "Spring Cleaning Event 2018"},
	29:     {BadgeID: 29, BadgeIcon: "29_salien/6_80.png", BadgeName: "Salien"},
	30:     {BadgeID: 30, BadgeIcon: "generic/RetiredModerator_80.png", BadgeName: "Retired Community Moderator"},
	31:     {BadgeID: 31, BadgeIcon: "30_steamawardnominations/level04_80.png", BadgeName: "Steam Awards Nomination Committee 2018"},
	33:     {BadgeID: 33, BadgeIcon: "33_cozycottage2018/1000000_80.png", BadgeName: "Winter 2018 Knick-Knack Collector"},
	34:     {BadgeID: 34, BadgeIcon: "34_lny2019/10_80.png", BadgeName: "Lunar New Year 2019"},
	36:     {BadgeID: 36, BadgeIcon: "36_springcleaning2019/gold_80x80.png", BadgeName: "Spring Cleaning Event 2019"},
	37:     {BadgeID: 37, BadgeIcon: "37_summer2019/level1000000_80.png", BadgeName: "Steam Grand Prix 2019"},
	38:     {BadgeID: 38, BadgeIcon: "37_summer2019/hare_gold_80.png", BadgeName: "Steam Grand Prix 2019 - Team Hare"},
	39:     {BadgeID: 39, BadgeIcon: "37_summer2019/tortoise_gold_80.png", BadgeName: "Steam Grand Prix 2019 - Team Tortoise"},
	40:     {BadgeID: 40, BadgeIcon: "37_summer2019/corgi_gold_80.png", BadgeName: "Steam Grand Prix 2019 - Team Corgi"},
	41:     {BadgeID: 41, BadgeIcon: "37_summer2019/cockatiel_gold_80.png", BadgeName: "Steam Grand Prix 2019 - Team Cockatiel"},
	42:     {BadgeID: 42, BadgeIcon: "37_summer2019/pig_gold_80.png", BadgeName: "Steam Grand Prix 2019 - Team Pig"},
	245070: {BadgeID: 1, AppID: 245070, BadgeIcon: "30a5de112a3512269cbc3d428fad3b9c507c56ba", BadgeName: "2013: Summer Getaway"},
	267420: {BadgeID: 1, AppID: 267420, BadgeIcon: "e041163b0c4d5cba61fb54612973612636cdd970", BadgeName: "2013: Holdiay Sale"},
	303700: {BadgeID: 1, AppID: 303700, BadgeIcon: "b3c3fa2821b32ce6bcc127e5ee3ec47845c35308", BadgeName: "2014: Summer Adventure"},
	335590: {BadgeID: 1, AppID: 335590, BadgeIcon: "b1c504dfaf4d073e5cf9c2d7d48c55c9cf11b7d3", BadgeName: "2014: Holdiay Sale"},
	368020: {BadgeID: 1, AppID: 368020, BadgeIcon: "49715c47e076456e0aacec76a5a0d714cadea951", BadgeName: "2015: Monster Summer Sale"},
	425280: {BadgeID: 1, AppID: 425280, BadgeIcon: "3442d0c36e5d549abf29872c9aec9f6e4364d23f", BadgeName: "2015: Holdiay Sale"},
	480730: {BadgeID: 1, AppID: 480730, BadgeIcon: "6b1280c07eedafdb3d9cac282f82da4365b9c98d", BadgeName: "2016: Summer Sale"},
	566020: {BadgeID: 1, AppID: 566020, BadgeIcon: "604be0b040a1117a5b8b5579b3c6ec25e540f458", BadgeName: "2016: Steam Awards"},
	639900: {BadgeID: 1, AppID: 639900, BadgeIcon: "9dd59323d14eb5dba94328db80e27caaee4c29ea", BadgeName: "2017: Summer Sale"},
	762800: {BadgeID: 1, AppID: 762800, BadgeIcon: "0a10b3b3725de8f72cb48fd94daff296cc3dfe52", BadgeName: "2017: Steam Awards"},
	876740: {BadgeID: 1, AppID: 876740, BadgeIcon: "9c677484f7f148045189a9dabe7efdf733e9e1f1", BadgeName: "2018: Intergalactic Summer"},
	991980: {BadgeID: 1, AppID: 991980, BadgeIcon: "3c96df81a7f82f23b68356c51733793cdece8f63", BadgeName: "2018: Winter Sale"},
}
