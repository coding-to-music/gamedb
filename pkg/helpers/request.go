package helpers

import "regexp"

var botRegex = regexp.MustCompile("/bot|crawl|slurp|wget|curl|spider|yandex|baidu|google|msn|bing|yahoo|jeeves|twitter|facebook/i")

func IsBot(userAgent string) bool {

	if userAgent == "" {
		return true
	}

	return botRegex.MatchString(userAgent)
}
