package helpers

import "regexp"

func IsBot(userAgent string) bool {

	r, _ := regexp.Compile("/bot|crawl|slurp|spider|google|msn|bing|yahoo|jeeves|facebook/i")
	if (userAgent != "") && r.MatchString(userAgent) {
		return true
	}
	return false
}
