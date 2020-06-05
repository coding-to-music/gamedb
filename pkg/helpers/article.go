package helpers

import (
	"html"
	"html/template"
	"regexp"
	"strings"
)

const articleImageBase = "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/clans/"

var trimArticles = regexp.MustCompile(`(?s)^(<br>)+|(<br>)+$`)

func GetArticleBody(body string) template.HTML {

	body = strings.ReplaceAll(body, "{STEAM_CLAN_IMAGE}", articleImageBase)
	body = BBCodeCompiler.Compile(body)
	body = html.UnescapeString(body)
	body = trimArticles.ReplaceAllString(body, "")

	return template.HTML(body)
}
