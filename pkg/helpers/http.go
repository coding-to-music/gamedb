package helpers

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gamedb/gamedb/pkg/log"
)

var botRegex = regexp.MustCompile("(?i)bot|crawl|slurp|wget|curl|spider|yandex|baidu|google|msn|bing|yahoo|jeeves|twitter|facebook")

func IsBot(userAgent string) bool {

	if userAgent == "" {
		return true
	}

	return botRegex.MatchString(userAgent)
}

func GetWithTimeout(link string, timeout time.Duration) (b []byte, code int, err error) {
	return requestWithTimeout("GET", link, timeout)
}

func HeadWithTimeout(link string, timeout time.Duration) (code int, err error) {

	operation := func() (err error) {
		_, code, err = requestWithTimeout("HEAD", link, timeout)
		return err
	}

	policy := backoff.NewExponentialBackOff()
	policy.InitialInterval = time.Second

	err = backoff.RetryNotify(operation, backoff.WithMaxRetries(policy, 5), func(err error, t time.Duration) { log.Info(err) })
	return code, err
}

func requestWithTimeout(method string, link string, timeout time.Duration) (body []byte, code int, err error) {

	if link == "" {
		return nil, 0, err
	}

	u, err := url.Parse(link)
	if err != nil {
		return nil, 0, err
	}

	if !u.IsAbs() {
		return nil, 0, err
	}

	if timeout == 0 {
		timeout = time.Second * 10
	}

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, 0, err
	}

	clientWithTimeout := &http.Client{
		Timeout: timeout,
	}

	resp, err := clientWithTimeout.Do(req)
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		err := resp.Body.Close()
		log.Err(err)
	}()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return body, resp.StatusCode, err
}

func GetIP() string {
	for _, v := range []string{"http://ipinfo.io/ip", "http://myexternalip.com/raw", "https://ifconfig.co/ip"} {

		body, _, err := GetWithTimeout(v, 0)
		if err != nil {
			continue
		}

		return strings.TrimSpace(string(body))
	}

	return ""
}
