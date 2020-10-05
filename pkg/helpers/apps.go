package helpers

import (
	"html/template"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gosimple/slug"
)

type ProductType int

const (
	ProductTypeApp ProductType = iota
	ProductTypePackage
)

const(
	AppIconBase    = "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" // Also the base for achievements
	DefaultAppIcon = "/assets/img/no-app-image-square.jpg"
)

func IsValidAppID(id int) bool {
	return id > 0 // Zero is valid
}

func GetAppPath(id int, name string) string {

	p := "/games/" + strconv.Itoa(id)

	if name != "" {
		p = p + "/" + slug.Make(name)
	}

	return p
}

func GetAppCommunityLink(appID int) string {
	name := config.C.GameDBShortName
	return "https://steamcommunity.com/app/" + strconv.Itoa(appID) + "?utm_source=" + name + "&utm_medium=link&curator_clanid=" // todo curator_clanid
}

func GetAppName(id int, name string) string {

	name = strings.TrimSpace(name)

	if name != "" {
		return name
	} else if id > 0 {
		return "App " + strconv.Itoa(id)
	}
	return "Unknown App"
}

func GetAppIcon(id int, icon string) string {

	if strings.HasPrefix(icon, "/") || strings.HasPrefix(icon, "http") {
		return icon
	}

	if IsValidAppID(id) && RegexSha1Only.MatchString(icon) {
		return AppIconBase + strconv.Itoa(id) + "/" + icon + ".jpg"
	}

	return DefaultAppIcon
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
		log.Warn("Missing state: " + state)
		return strings.Title(state)
	}
}

func GetAppReleaseDateNice(releaseDateOriginal int64, releaseDateSteam int64, releaseDate string) string {

	if releaseDateOriginal > 0 {
		return time.Unix(releaseDateOriginal, 0).Format(DateYear)
	} else if releaseDateSteam > 0 {
		return time.Unix(releaseDateSteam, 0).Format(DateYear)
	} else if releaseDate != "" {
		return releaseDate
	} else {
		return "-" // Can't return empty, for Discord
	}
}

func GetAppStoreLink(appID int) string {
	name := config.C.GameDBShortName
	return "https://store.steampowered.com/app/" + strconv.Itoa(appID) + "?utm_source=" + name + "&utm_medium=link&curator_clanid=" // todo curator_clanid
}

func GetAppType(appType string) (ret string) {

	switch appType {
	case "dlc":
		return "DLC"
	case "advertising":
		return "Advertisement"
	case "music":
		return "Soundtrack"
	case "":
		return "Unknown"
	default:
		return strings.Title(appType)
	}
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

var microTrailerRegex = regexp.MustCompile(`\/[a-z0-9_]+\.`)

func (video AppVideo) Micro() string {
	return microTrailerRegex.ReplaceAllString(video.PathFull, "/microtrailer.")
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

func (ar AppReview) GetReview() template.HTML {
	return template.HTML(BBCodeCompiler.Compile(ar.Review))
}

type SystemRequirement struct {
	Key string
	Val string
}

func (sr SystemRequirement) Format() template.HTML {

	switch sr.Val {
	case "0":
		return `<i class="fas fa-times text-danger"></i>`
	case "1":
		return `<i class="fas fa-check text-success"></i>`
	case "warn":
		return `<span class="text-warning">Warn</span>`
	case "deny":
		return `<span class="text-danger">Deny</span>`
	default:
		return template.HTML(sr.Val)
	}
}

func GetAppReviewScore(score float64) string {
	return FloatToString(score, 2) + "%"
}

func GetAppFollowers(groupID string, followers int) string {

	if groupID == "" {
		return "-"
	}

	return humanize.Comma(int64(followers))
}

func GetAppPlayLink(appID int) template.URL {
	return template.URL("steam://run/" + strconv.Itoa(appID))
}
