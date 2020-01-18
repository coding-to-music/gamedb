package helpers

import (
	"html/template"
	"strconv"
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gosimple/slug"
)

type ProductType string

const (
	ProductTypeApp     ProductType = "product"
	ProductTypePackage ProductType = "package"

	DefaultAppIcon = "/assets/img/no-app-image-square.jpg"
)

func IsValidAppID(id int) bool {
	return id != 0
}

func GetAppPath(id int, name string) string {

	p := "/apps/" + strconv.Itoa(id)

	if name != "" {
		p = p + "/" + slug.Make(name)
	}

	return p
}

func GetAppName(id int, name string) string {

	if name != "" {
		return strings.TrimSpace(name)
	} else if id > 0 {
		return "App " + strconv.Itoa(id)
	}
	return "Unknown App"
}

func GetAppIcon(id int, icon string) string {

	if icon == "" {
		return DefaultAppIcon
	} else if strings.HasPrefix(icon, "/") || strings.HasPrefix(icon, "http") {
		return icon
	}
	return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(id) + "/" + icon + ".jpg"
}

func GetAppReleaseState(state string) (ret string) {

	switch state {
	case "preloadonly":
		return "Preload Only"
	case "prerelease":
		return "Pre Release"
	case "released":
		return "Released"
	case "":
		return "Unreleased"
	default:
		log.Warning("Missing state: " + state)
		return strings.Title(state)
	}
}

func GetAppReleaseDateNice(releaseDateUnix int64, releaseDate string) string {

	if releaseDateUnix == 0 {
		return releaseDate
	}

	return time.Unix(releaseDateUnix, 0).Format(DateYear)
}

func GetAppStoreLink(appID int) string {
	name := config.Config.GameDBShortName.Get()
	return "https://store.steampowered.com/app/" + strconv.Itoa(appID) + "?utm_source=" + name + "&utm_medium=link&curator_clanid=" // todo curator_clanid
}

func GetAppType(appType string) (ret string) {

	switch appType {
	case "dlc":
		return "DLC"
	case "":
		return "Unknown"
	default:
		return strings.Title(appType)
	}
}

//
type AppAchievement struct {
	Name        string  `json:"n"`
	Icon        string  `json:"i"`
	Description string  `json:"d"`
	Completed   float64 `json:"c"`
	Active      bool    `json:"a"`
}

func (a AppAchievement) GetIcon() string {
	if strings.HasSuffix(a.Icon, ".jpg") {
		return a.Icon
	}
	return DefaultAppIcon
}

//
type AppImage struct {
	PathFull      string `json:"f"`
	PathThumbnail string `json:"t"`
}

type AppVideo struct {
	PathFull      string `json:"f"`
	PathThumbnail string `json:"s"`
	Title         string `json:"t"`
}

type AppStat struct {
	Name        string `json:"n"`
	Default     int    `json:"d"`
	DisplayName string `json:"o"`
}

type AppSteamSpy struct {
	SSAveragePlaytimeTwoWeeks int `json:"aw"`
	SSAveragePlaytimeForever  int `json:"af"`
	SSMedianPlaytimeTwoWeeks  int `json:"mw"`
	SSMedianPlaytimeForever   int `json:"mf"`
	SSOwnersLow               int `json:"ol"`
	SSOwnersHigh              int `json:"oh"`
}

func (ss AppSteamSpy) GetSSAveragePlaytimeTwoWeeks() float64 {
	return RoundFloatTo1DP(float64(ss.SSAveragePlaytimeTwoWeeks) / 60)
}

func (ss AppSteamSpy) GetSSAveragePlaytimeForever() float64 {
	return RoundFloatTo1DP(float64(ss.SSAveragePlaytimeForever) / 60)
}

func (ss AppSteamSpy) GetSSMedianPlaytimeTwoWeeks() float64 {
	return RoundFloatTo1DP(float64(ss.SSMedianPlaytimeTwoWeeks) / 60)
}

func (ss AppSteamSpy) GetSSMedianPlaytimeForever() float64 {
	return RoundFloatTo1DP(float64(ss.SSMedianPlaytimeForever) / 60)
}

type AppReviewSummary struct {
	Positive int
	Negative int
	Reviews  []AppReview
}

func (r AppReviewSummary) GetTotal() int {
	return r.Negative + r.Positive
}

func (r AppReviewSummary) GetPositivePercent() float64 {
	return float64(r.Positive) / float64(r.GetTotal()) * 100
}

func (r AppReviewSummary) GetNegativePercent() float64 {
	return float64(r.Negative) / float64(r.GetTotal()) * 100
}

type AppReview struct {
	Review     string `json:"r"`
	Vote       bool   `json:"v"`
	VotesGood  int    `json:"g"`
	VotesFunny int    `json:"f"`
	Created    string `json:"c"`
	PlayerPath string `json:"p"`
	PlayerName string `json:"n"`
}

func (ar AppReview) HTML() template.HTML {
	return template.HTML(ar.Review)
}
