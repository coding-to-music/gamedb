package helpers

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/mssola/user_agent"
	"go.uber.org/zap"
)

func IsBot(userAgent string) bool {

	if userAgent == "" || strings.Contains(strings.ToLower(userAgent), strings.ToLower("bot")) {
		return true
	}

	return user_agent.New(userAgent).Bot()
}

func Get(link string, timeout time.Duration, headers http.Header) (b []byte, code int, err error) {
	return requestWithTimeout("GET", link, timeout, headers, nil)
}

func Delete(link string, timeout time.Duration, headers http.Header) (b []byte, code int, err error) {
	return requestWithTimeout("DELETE", link, timeout, headers, nil)
}

func Head(link string, timeout time.Duration) (code int, err error) {

	operation := func() (err error) {
		_, code, err = requestWithTimeout("HEAD", link, timeout, nil, nil)
		if code == 404 {
			return backoff.Permanent(err)
		}
		return err
	}

	policy := backoff.NewExponentialBackOff()
	policy.InitialInterval = time.Second * 3

	err = backoff.RetryNotify(operation, backoff.WithMaxRetries(policy, 5), func(err error, t time.Duration) { log.Info("Doing a HEAD call", zap.Error(err), zap.Duration("duration", timeout), zap.String("link", link)) })
	return code, err
}

func Post(link string, data url.Values, headers http.Header) (b []byte, code int, err error) {
	return requestWithTimeout("POST", link, 0, headers, data)
}

var ErrNon200 = errors.New("server responded with error")

func requestWithTimeout(method string, link string, timeout time.Duration, headers http.Header, data url.Values) (body []byte, code int, err error) {

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

	var reader io.Reader
	if len(data) > 0 {
		reader = bytes.NewBufferString(data.Encode())
	}

	req, err := http.NewRequest(method, u.String(), reader)
	if err != nil {
		return nil, 0, err
	}

	if len(headers) > 0 {
		req.Header = headers
	}

	clientWithTimeout := &http.Client{
		Timeout: timeout,
	}

	resp, err := clientWithTimeout.Do(req)
	if err != nil {
		return nil, 0, err
	}

	defer Close(resp.Body)

	if resp.StatusCode != 200 {
		return nil, resp.StatusCode, ErrNon200
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return body, resp.StatusCode, err
}

func GetIP() string {

	if config.IsLocal() {
		return "127.0 0.1"
	}

	for _, v := range []string{"http://ipinfo.io/ip", "http://myexternalip.com/raw", "https://ifconfig.co/ip"} {

		body, code, err := Get(v, 0, nil)
		if err != nil || code != 200 {
			continue
		}

		return strings.TrimSpace(string(body))
	}

	return ""
}
