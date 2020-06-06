package helpers

import (
	"io/ioutil"
	"strings"

	"github.com/gamedb/gamedb/pkg/log"
)

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
