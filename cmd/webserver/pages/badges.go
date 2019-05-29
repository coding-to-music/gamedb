package pages

import (
	"net/http"
	"strconv"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
	"github.com/gosimple/slug"
)

func BadgesRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", badgesHandler)
	r.Mount("/{id}", BadgeRouter())
	return r
}

const (
	specialImageBase = "https://steamcommunity-a.akamaihd.net/public/images/badges/"
	eventImageBase   = "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/items/"
)

type badge struct {
	BadgeID int
	AppID   int
	Image   string
	Name    string
	Max     int
	MaxFoil int
}

func (b badge) GetPath() string {
	return "/badges/" + strconv.Itoa(b.BadgeID) + "/" + slug.Make(b.Name)
}

func (b badge) GetIcon() string {
	if b.AppID > 0 {
		return eventImageBase + "/" + strconv.Itoa(b.AppID) + "/" + b.Image + ".png"
	}
	return specialImageBase + b.Image
}

var (
	specialBadges = []badge{
		{BadgeID: 36, Image: "36_springcleaning2019/gold_80x80.png", Name: "Spring Cleaning Event 2019"},
		{BadgeID: 34, Image: "34_lny2019/10_80.png", Name: "Lunar New Year 2019"},
		{BadgeID: 33, Image: "33_cozycottage2018/1000000_80.png", Name: "Winter 2018 Knick-Knack Collector"},
		{BadgeID: 31, Image: "30_steamawardnominations/level04_80.png", Name: "Steam Awards Nomination Committee 2018"},
		{BadgeID: 30, Image: "generic/RetiredModerator_80.png", Name: "Retired Community Moderator"},
		{BadgeID: 29, Image: "29_salien/6_80.png", Name: "Salien"},
		{BadgeID: 28, Image: "28_springcleaning2018/gold_80.png", Name: "Spring Cleaning Event 2018"},
		{BadgeID: 27, Image: "27_steamawardnominations/level04_80.png", Name: "Steam Awards Nomination Committee 2017"},
		{BadgeID: 26, Image: "26_summer2017_sticker/completionist.png", Name: "Sticker Completionist"},
		{BadgeID: 25, Image: "25_steamawardnominations/level04_80.png", Name: "Steam Awards Nomination Committee 2016"},
		{BadgeID: 24, Image: "24_winter2015_arg_red_herring/red_herring.png", Name: "Red Herring"},
		{BadgeID: 23, Image: "23_towerattack/wormhole.png", Name: "Monster Summer"},
		{BadgeID: 22, Image: "22_golden/owner_80.png", Name: "2014 HolBadgeIDay Profile Recipient"},
		{BadgeID: 21, Image: "21_auction/winner_80.png?v=2", Name: "Auction Participant/Winner"},
		{BadgeID: 20, Image: "16_summer2014/team_purple.png", Name: "Steam Summer Adventure 2014 - Purple Team"},
		{BadgeID: 19, Image: "16_summer2014/team_green.png", Name: "Steam Summer Adventure 2014 - Green Team"},
		{BadgeID: 18, Image: "16_summer2014/team_pink.png", Name: "Steam Summer Adventure 2014 - Pink Team"},
		{BadgeID: 17, Image: "16_summer2014/team_blue.png", Name: "Steam Summer Adventure 2014 - Blue Team"},
		{BadgeID: 16, Image: "16_summer2014/team_red.png", Name: "Steam Summer Adventure 2014 - Red Team"},
		{BadgeID: 15, Image: "15_hwbeta/hwbeta03_80.png", Name: "Steam Hardware Beta"},
		{BadgeID: 14, Image: "generic/TradingCardBeta_80.png", Name: "Trading Card Beta Tester"},
		{BadgeID: 13, Image: "13_gamecollector/25000_80.png", Name: "Owned Games"},
		{BadgeID: 12, Image: "generic/GameDeveloper_80.png", Name: "Steamworks Developer"},
		{BadgeID: 11, Image: "generic/ValveEmployee_80.png", Name: "Valve Employee"},
		{BadgeID: 10, Image: "generic/CommunityModerator_80.png", Name: "Steam Community Moderator"},
		{BadgeID: 9, Image: "09_communitytranslator/translator_level4_80.png", Name: "Steam Community Translator"},
		{BadgeID: 8, Image: "08_winter2012/winter2012_badge80.png", Name: "Steam HolBadgeIDay Sale 2012"},
		{BadgeID: 7, Image: "07_summer2012/Summer2012_stage3_80.png", Name: "Steam Summer Sale 2012"},
		{BadgeID: 6, Image: "06_winter2011/coal03_80.png", Name: "Steam HolBadgeIDay Sale 2011"},
		{BadgeID: 5, Image: "05_summer2011/tickets80.png", Name: "Steam Summer Camp"},
		{BadgeID: 4, Image: "04_treasurehunt/treasurehunt03_80.png", Name: "The Great Steam Treasure Hunt"},
		{BadgeID: 3, Image: "03_potato/potato03_80.png", Name: "The Potato Sack"},
		{BadgeID: 2, Image: "01_community/community03_80.png", Name: "Community Ambassador"},
		{BadgeID: 1, Image: "02_years/steamyears1002_80.png", Name: "Years of Service"},
	}
	eventBadges = []badge{
		{BadgeID: 991980, AppID: 991980, Image: "3c96df81a7f82f23b68356c51733793cdece8f63", Name: "2018: Winter Sale"},
		{BadgeID: 876740, AppID: 876740, Image: "9c677484f7f148045189a9dabe7efdf733e9e1f1", Name: "2018: Intergalactic Summer"},
		{BadgeID: 762800, AppID: 762800, Image: "0a10b3b3725de8f72cb48fd94daff296cc3dfe52", Name: "2017: Steam Awards"},
		{BadgeID: 639900, AppID: 639900, Image: "9dd59323d14eb5dba94328db80e27caaee4c29ea", Name: "2017: Summer Sale"},
		{BadgeID: 566020, AppID: 566020, Image: "604be0b040a1117a5b8b5579b3c6ec25e540f458", Name: "2016: Steam Awards"},
		{BadgeID: 480730, AppID: 480730, Image: "6b1280c07eedafdb3d9cac282f82da4365b9c98d", Name: "2016: Summer Sale"},
		{BadgeID: 425280, AppID: 425280, Image: "3442d0c36e5d549abf29872c9aec9f6e4364d23f", Name: "2015: HolAppIDay Sale"},
		{BadgeID: 368020, AppID: 368020, Image: "49715c47e076456e0aacec76a5a0d714cadea951", Name: "2015: Monster Summer Sale"},
		{BadgeID: 335590, AppID: 335590, Image: "b1c504dfaf4d073e5cf9c2d7d48c55c9cf11b7d3", Name: "2014: HolAppIDay Sale"},
		{BadgeID: 303700, AppID: 303700, Image: "b3c3fa2821b32ce6bcc127e5ee3ec47845c35308", Name: "2014: Summer Adventure"},
		{BadgeID: 267420, AppID: 267420, Image: "e041163b0c4d5cba61fb54612973612636cdd970", Name: "2013: HolAppIDay Sale"},
		{BadgeID: 245070, AppID: 245070, Image: "30a5de112a3512269cbc3d428fad3b9c507c56ba", Name: "2013: Summer Getaway"},
	}
)

func badgesHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	t := badgesTemplate{}
	t.fill(w, r, "Badges", "")
	t.SpecialBadges = specialBadges
	t.EventBadges = eventBadges
	t.SpecialBase = specialImageBase
	t.EventBase = eventImageBase

	err = returnTemplate(w, r, "badges", t)
	log.Err(err, r)
}

type badgesTemplate struct {
	GlobalTemplate
	EventBadges   []badge
	SpecialBadges []badge
	EventBase     string
	SpecialBase   string
}
