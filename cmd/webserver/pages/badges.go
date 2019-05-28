package pages

import (
	"net/http"
	"strconv"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
	"github.com/gosimple/slug"
)

const badgeImageBase = "https://steamcommunity-a.akamaihd.net/public/images/badges/"

type badge struct {
	ID    int
	Image string
	Name  string
	Max   int
}

func (b badge) GetPath() string {
	return "/badges/" + strconv.Itoa(b.ID) + "/" + slug.Make(b.Name)
}

var badges = []badge{
	{ID: 36, Image: "36_springcleaning2019/gold_80x80.png", Name: "Spring Cleaning Event 2019"},
	{ID: 34, Image: "34_lny2019/10_80.png", Name: "Lunar New Year 2019"},
	{ID: 33, Image: "33_cozycottage2018/1000000_80.png", Name: "Winter 2018 Knick-Knack Collector"},
	{ID: 31, Image: "30_steamawardnominations/level04_80.png", Name: "Steam Awards Nomination Committee 2018"},
	{ID: 30, Image: "generic/RetiredModerator_80.png", Name: "Retired Community Moderator"},
	{ID: 29, Image: "29_salien/6_80.png", Name: "Salien"},
	{ID: 28, Image: "28_springcleaning2018/gold_80.png", Name: "Spring Cleaning Event 2018"},
	{ID: 27, Image: "27_steamawardnominations/level04_80.png", Name: "Steam Awards Nomination Committee 2017"},
	{ID: 26, Image: "26_summer2017_sticker/completionist.png", Name: "Sticker Completionist"},
	{ID: 25, Image: "25_steamawardnominations/level04_80.png", Name: "Steam Awards Nomination Committee 2016"},
	{ID: 24, Image: "24_winter2015_arg_red_herring/red_herring.png", Name: "Red Herring"},
	{ID: 23, Image: "23_towerattack/wormhole.png", Name: "Monster Summer"},
	{ID: 22, Image: "22_golden/owner_80.png", Name: "2014 Holiday Profile Recipient"},
	{ID: 21, Image: "21_auction/winner_80.png?v=2", Name: "Auction Participant/Winner"},
	{ID: 20, Image: "16_summer2014/team_purple.png", Name: "Steam Summer Adventure 2014 - Purple Team"},
	{ID: 19, Image: "16_summer2014/team_green.png", Name: "Steam Summer Adventure 2014 - Green Team"},
	{ID: 18, Image: "16_summer2014/team_pink.png", Name: "Steam Summer Adventure 2014 - Pink Team"},
	{ID: 17, Image: "16_summer2014/team_blue.png", Name: "Steam Summer Adventure 2014 - Blue Team"},
	{ID: 16, Image: "16_summer2014/team_red.png", Name: "Steam Summer Adventure 2014 - Red Team"},
	{ID: 15, Image: "15_hwbeta/hwbeta03_80.png", Name: "Steam Hardware Beta"},
	{ID: 14, Image: "generic/TradingCardBeta_80.png", Name: "Trading Card Beta Tester"},
	{ID: 13, Image: "13_gamecollector/25000_80.png", Name: "Owned Games"},
	{ID: 12, Image: "generic/GameDeveloper_80.png", Name: "Steamworks Developer"},
	{ID: 11, Image: "generic/ValveEmployee_80.png", Name: "Valve Employee"},
	{ID: 10, Image: "generic/CommunityModerator_80.png", Name: "Steam Community Moderator"},
	{ID: 9, Image: "09_communitytranslator/translator_level4_80.png", Name: "Steam Community Translator"},
	{ID: 8, Image: "08_winter2012/winter2012_badge80.png", Name: "Steam Holiday Sale 2012"},
	{ID: 7, Image: "07_summer2012/Summer2012_stage3_80.png", Name: "Steam Summer Sale 2012"},
	{ID: 6, Image: "06_winter2011/coal03_80.png", Name: "Steam Holiday Sale 2011"},
	{ID: 5, Image: "05_summer2011/tickets80.png", Name: "Steam Summer Camp"},
	{ID: 4, Image: "04_treasurehunt/treasurehunt03_80.png", Name: "The Great Steam Treasure Hunt"},
	{ID: 3, Image: "03_potato/potato03_80.png", Name: "The Potato Sack"},
	{ID: 2, Image: "01_community/community03_80.png", Name: "Community Ambassador"},
	{ID: 1, Image: "02_years/steamyears1002_80.png", Name: "Years of Service"},
}

func BadgesRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", badgesHandler)
	r.Mount("/{id}", BadgeRouter())
	return r
}

func badgesHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	t := badgesTemplate{}
	t.fill(w, r, "Badges", "")
	t.Badges = badges
	t.ImageBase = badgeImageBase

	err = returnTemplate(w, r, "badges", t)
	log.Err(err, r)
}

type badgesTemplate struct {
	GlobalTemplate
	Badges    []badge
	ImageBase string
}
