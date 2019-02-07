package helpers

import (
	"net/http"
	"net/url"
)

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
