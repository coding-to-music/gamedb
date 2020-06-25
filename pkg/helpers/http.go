package helpers

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
)

var botRegex = regexp.MustCompile("(?i)bot|crawl|slurp|wget|curl|spider|yandex|baidu|google|msn|bing|yahoo|jeeves|twitter|facebook")

func IsBot(userAgent string) bool {

	if userAgent == "" {
		return true
	}

	return botRegex.MatchString(userAgent)
}

func GetWithTimeout(url string, timeout time.Duration) (*http.Response, error) {

	if timeout == 0 {
		timeout = time.Second * 10
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	clientWithTimeout := &http.Client{
		Timeout: timeout,
	}

	return clientWithTimeout.Do(req)
}

// Returns 0 on fail
func GetResponseCode(link string) (code int) {

	if link == "" {
		return 0
	}

	u, err := url.Parse(link)
	if err != nil {
		return 0
	}

	if !u.IsAbs() {
		return 0
	}

	resp, err := http.Head(u.String())
	if err != nil {
		return 0
	}

	return resp.StatusCode
}

func GetIP() string {
	for _, v := range []string{"http://ipinfo.io/ip", "http://myexternalip.com/raw", "https://ifconfig.co/ip"} {

		resp, err := GetWithTimeout(v, 0)
		if err != nil {
			continue
		}
		//noinspection GoDeferInLoop
		defer func() {
			err := resp.Body.Close()
			log.Err(err)
		}()

		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		return strings.TrimSpace(string(bytes))
	}

	return ""
}
