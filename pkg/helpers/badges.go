package helpers

import (
	"strconv"
	"strings"

	"github.com/gosimple/slug"
)

// Dont input a unique ID
func IsBadgeSpecial(appID int) bool {
	return appID == 0
}

func IsBadgeEvent(appID int) bool {
	val, ok := BuiltInEventBadges[appID]
	return ok && val.BadgeID > 0
}

func IsBadgeGame(appID int) bool {
	return !IsBadgeSpecial(appID) && !IsBadgeEvent(appID)
}

func GetBadgeUniqueID(appID int, badgeID int) int {

	if IsBadgeSpecial(appID) {
		return badgeID
	}
	return appID
}

func GetBadgeName(badgeName string, uniqueID int) string {

	if val, ok := BuiltInSpecialBadges[uniqueID]; ok {
		return val.Name
	}

	if val, ok := BuiltInEventBadges[uniqueID]; ok {
		return val.Name
	}

	return GetAppName(uniqueID, badgeName)
}

func GetBadgePath(badgeName string, appID int, badgeID int, foil bool) string {

	var f string
	if foil {
		f = "?foil=1"
	}

	var id = GetBadgeUniqueID(appID, badgeID)
	var name = GetBadgeName(badgeName, id)

	return "/badges/" + strconv.Itoa(id) + "/" + slug.Make(name) + f
}

const (
	specialImageBase = "https://steamcommunity-a.akamaihd.net/public/images/badges/"
	eventImageBase   = "https://cdn.cloudflare.steamstatic.com/steamcommunity/public/images/items/"
)

func GetBadgeIcon(icon string, appID int, badgeID int) string {

	if IsBadgeEvent(appID) {
		if val, ok := BuiltInEventBadges[appID]; ok {
			return eventImageBase + strconv.Itoa(val.AppID) + "/" + val.Icon + ".png"
		}
	}

	if IsBadgeSpecial(appID) {
		if val, ok := BuiltInSpecialBadges[badgeID]; ok {
			return specialImageBase + val.Icon
		}
	}

	if strings.HasPrefix(icon, "http") || strings.HasPrefix(icon, "/") {
		return icon
	}

	if appID > 0 && icon != "" {
		return GetAppIcon(appID, icon)
	}

	// Return blank for badge html
	return ""
}

type BuiltInbadge struct {
	AppID   int
	BadgeID int
	Name    string
	Icon    string
}

func (badge BuiltInbadge) GetPath(foil bool) string {
	return GetBadgePath(badge.Name, badge.AppID, badge.BadgeID, foil)
}

func (badge BuiltInbadge) GetIcon() string {
	return GetBadgeIcon(badge.Icon, badge.AppID, badge.BadgeID)
}

func (badge BuiltInbadge) GetName() string {
	return GetBadgeName(badge.Name, GetBadgeUniqueID(badge.AppID, badge.BadgeID))
}

func (badge BuiltInbadge) ID() int {
	return GetBadgeUniqueID(badge.AppID, badge.BadgeID)
}

func (badge BuiltInbadge) IsSpecial() bool {
	return IsBadgeSpecial(badge.AppID)
}

var BuiltInSpecialBadges = map[int]BuiltInbadge{
	// Special
	1:  {BadgeID: 1, Icon: "02_years/steamyears1_80.png", Name: "Years of Service"},
	2:  {BadgeID: 2, Icon: "01_community/community03_80.png", Name: "Community Ambassador"},
	3:  {BadgeID: 3, Icon: "03_potato/potato03_80.png", Name: "The Potato Sack"},
	4:  {BadgeID: 4, Icon: "04_treasurehunt/treasurehunt03_80.png", Name: "The Great Steam Treasure Hunt"},
	5:  {BadgeID: 5, Icon: "05_summer2011/tickets80.png", Name: "Steam Summer Camp"},
	6:  {BadgeID: 6, Icon: "06_winter2011/coal03_80.png", Name: "Steam Holiday Sale 2011"},
	7:  {BadgeID: 7, Icon: "07_summer2012/Summer2012_stage3_80.png", Name: "Steam Summer Sale 2012"},
	8:  {BadgeID: 8, Icon: "08_winter2012/winter2012_badge80.png", Name: "Steam Holiday Sale 2012"},
	9:  {BadgeID: 9, Icon: "09_communitytranslator/translator_level4_80.png", Name: "Steam Community Translator"},
	10: {BadgeID: 10, Icon: "generic/CommunityModerator_80.png", Name: "Steam Community Moderator"},
	11: {BadgeID: 11, Icon: "generic/ValveEmployee_80.png", Name: "Valve Employee"},
	12: {BadgeID: 12, Icon: "generic/GameDeveloper_80.png", Name: "Steamworks Developer"},
	13: {BadgeID: 13, Icon: "13_gamecollector/25000_80.png", Name: "Owned Games"},
	14: {BadgeID: 14, Icon: "generic/TradingCardBeta_80.png", Name: "Trading Card Beta Tester"},
	15: {BadgeID: 15, Icon: "15_hwbeta/hwbeta03_80.png", Name: "Steam Hardware Beta"},
	16: {BadgeID: 16, Icon: "16_summer2014/team_red.png", Name: "Steam Summer Adventure 2014 - Red Team"},
	17: {BadgeID: 17, Icon: "16_summer2014/team_blue.png", Name: "Steam Summer Adventure 2014 - Blue Team"},
	18: {BadgeID: 18, Icon: "16_summer2014/team_pink.png", Name: "Steam Summer Adventure 2014 - Pink Team"},
	19: {BadgeID: 19, Icon: "16_summer2014/team_green.png", Name: "Steam Summer Adventure 2014 - Green Team"},
	20: {BadgeID: 20, Icon: "16_summer2014/team_purple.png", Name: "Steam Summer Adventure 2014 - Purple Team"},
	21: {BadgeID: 21, Icon: "21_auction/winner_80.png?v=2", Name: "Auction Participant/Winner"},
	22: {BadgeID: 22, Icon: "22_golden/owner_80.png", Name: "2014 Holiday Profile Recipient"},
	23: {BadgeID: 23, Icon: "23_towerattack/wormhole.png", Name: "Monster Summer"},
	24: {BadgeID: 24, Icon: "24_winter2015_arg_red_herring/red_herring.png", Name: "Red Herring"},
	25: {BadgeID: 25, Icon: "25_steamawardnominations/level04_80.png", Name: "Steam Awards Nomination Committee 2016"},
	26: {BadgeID: 26, Icon: "26_summer2017_sticker/completionist.png", Name: "Sticker Completionist"},
	27: {BadgeID: 27, Icon: "27_steamawardnominations/level04_80.png", Name: "Steam Awards Nomination Committee 2017"},
	28: {BadgeID: 28, Icon: "28_springcleaning2018/gold_80.png", Name: "Spring Cleaning Event 2018"},
	29: {BadgeID: 29, Icon: "29_salien/6_80.png", Name: "Salien"},
	30: {BadgeID: 30, Icon: "generic/RetiredModerator_80.png", Name: "Retired Community Moderator"},
	31: {BadgeID: 31, Icon: "30_steamawardnominations/level04_80.png", Name: "Steam Awards Nomination Committee 2018"},
	32: {BadgeID: 32, Icon: "generic/ValveEmployee_80.png", Name: "Valve Moderator"},
	33: {BadgeID: 33, Icon: "33_cozycottage2018/1000000_80.png", Name: "Winter 2018 Knick-Knack Collector"},
	34: {BadgeID: 34, Icon: "34_lny2019/10_80.png", Name: "Lunar New Year 2019"},
	35: {BadgeID: 34, Icon: "https://steamcommunity-a.akamaihd.net/economy/image/-9a81dlWLwJ2UUGcVs_nsVtzdOEdtWwKGZZLQHTxH5rd9eDAjcFyv45SRYAFMIcKL_PArgVSL403ulRUWEndVKv8h56EAgQkalZSsuOnegRm1aqed2oStIXlkIHez6aiNe6CkzIAuJcgiLGU8I6kjgz6ux07-Ytsxtc/96fx96f", Name: "Lunar New Year 2019 Golden Profile"},
	36: {BadgeID: 36, Icon: "36_springcleaning2019/gold_80x80.png", Name: "Spring Cleaning Event 2019"},
	37: {BadgeID: 37, Icon: "37_summer2019/level1000000_80.png", Name: "Steam Grand Prix 2019"},
	38: {BadgeID: 38, Icon: "37_summer2019/hare_gold_80.png", Name: "Steam Grand Prix 2019 - Team Hare"},
	39: {BadgeID: 39, Icon: "37_summer2019/tortoise_gold_80.png", Name: "Steam Grand Prix 2019 - Team Tortoise"},
	40: {BadgeID: 40, Icon: "37_summer2019/corgi_gold_80.png", Name: "Steam Grand Prix 2019 - Team Corgi"},
	41: {BadgeID: 41, Icon: "37_summer2019/cockatiel_gold_80.png", Name: "Steam Grand Prix 2019 - Team Cockatiel"},
	42: {BadgeID: 42, Icon: "37_summer2019/pig_gold_80.png", Name: "Steam Grand Prix 2019 - Team Pig"},
	43: {BadgeID: 43, Icon: "43_steamawardnominations/level04_80.png", Name: "Steam Awards Nomination Committee 2019"},
	44: {BadgeID: 44, Icon: "44_winter2019/level15_80.png", Name: "Winter Sale 2019"},
	45: {BadgeID: 45, Icon: "45_steamville2019/key_to_city_80.png", Name: "Steamville 2019"},
	46: {BadgeID: 46, Icon: "46_lny2020/10_80.png", Name: "Lunar New Year 2020"},
	47: {BadgeID: 47, Icon: "47_springcleaning2020/dewey_badge_3.0_80x80.png", Name: "Spring Cleaning Event 2020"},
	48: {BadgeID: 48, Icon: "48_communitycontributor/1_80.png", Name: "Community Contributor"},
	49: {BadgeID: 49, Icon: "49_communitypatron/1_80.png", Name: "Community Patron"},
	50: {BadgeID: 50, Icon: "50_steamawardnominations/level04_80.png", Name: "Steam Award Nominations"},
}

var BuiltInEventBadges = map[int]BuiltInbadge{
	245070:  {BadgeID: 1, AppID: 245070, Icon: "30a5de112a3512269cbc3d428fad3b9c507c56ba", Name: "2013: Summer Getaway"},
	267420:  {BadgeID: 1, AppID: 267420, Icon: "e041163b0c4d5cba61fb54612973612636cdd970", Name: "2013: Holdiay Sale"},
	303700:  {BadgeID: 1, AppID: 303700, Icon: "b3c3fa2821b32ce6bcc127e5ee3ec47845c35308", Name: "2014: Summer Adventure"},
	335590:  {BadgeID: 1, AppID: 335590, Icon: "b1c504dfaf4d073e5cf9c2d7d48c55c9cf11b7d3", Name: "2014: Holdiay Sale"},
	368020:  {BadgeID: 1, AppID: 368020, Icon: "49715c47e076456e0aacec76a5a0d714cadea951", Name: "2015: Monster Summer Sale"},
	425280:  {BadgeID: 1, AppID: 425280, Icon: "3442d0c36e5d549abf29872c9aec9f6e4364d23f", Name: "2015: Holdiay Sale"},
	480730:  {BadgeID: 1, AppID: 480730, Icon: "6b1280c07eedafdb3d9cac282f82da4365b9c98d", Name: "2016: Summer Sale"},
	566020:  {BadgeID: 1, AppID: 566020, Icon: "604be0b040a1117a5b8b5579b3c6ec25e540f458", Name: "2016: Steam Awards"},
	639900:  {BadgeID: 1, AppID: 639900, Icon: "9dd59323d14eb5dba94328db80e27caaee4c29ea", Name: "2017: Summer Sale"},
	762800:  {BadgeID: 1, AppID: 762800, Icon: "0a10b3b3725de8f72cb48fd94daff296cc3dfe52", Name: "2017: Steam Awards"},
	876740:  {BadgeID: 1, AppID: 876740, Icon: "9c677484f7f148045189a9dabe7efdf733e9e1f1", Name: "2018: Intergalactic Summer"},
	991980:  {BadgeID: 1, AppID: 991980, Icon: "3c96df81a7f82f23b68356c51733793cdece8f63", Name: "2018: Winter Sale"},
	1195670: {BadgeID: 1, AppID: 1195670, Icon: "581a14e34100d7f4955ceff9365a7e40a89b57c8", Name: "2019: Winter Sale"},
	1263950: {BadgeID: 1, AppID: 1263950, Icon: "24a62a6fa825d6b0a174675fb97a01e9df81d030", Name: "2020: The Debut Collection"},
	1343890: {BadgeID: 1, AppID: 1343890, Icon: "89bf983bd41c23c8e4f4e24155e995e770e51070", Name: "2020: Summer Road Trip"},
	1465680: {BadgeID: 1, AppID: 1465680, Icon: "64f30989f0794a5437b4b501eb932b2f119616ae", Name: "2020: Winter Sale"},
	1492660: {BadgeID: 1, AppID: 1492660, Icon: "526005d8e2862a721f0079c52eb422fa2f293344", Name: "2020: Winter Collection"},
	// 1083560: {BadgeID: 1, AppID: 1083560, Icon: "", Name: ""},
}
