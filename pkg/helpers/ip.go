package helpers

import (
	"io/ioutil"
	"net/http"
	"strings"
)

func GetIP() string {
	for _, v := range []string{"http://ipinfo.io/ip", "http://myexternalip.com/raw", "https://ifconfig.co/ip"} {

		resp, err := http.Get(v)
		if err != nil {
			continue
		}
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		return strings.TrimSpace(string(bytes))
	}

	return ""
}
