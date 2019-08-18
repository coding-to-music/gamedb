package helpers

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
)

var (
	steamClient     *steam.Steam
	steamClientLock sync.Mutex

	steamClientUnlimited     *steam.Steam
	steamClientUnlimitedLock sync.Mutex
)

type steamLogger struct {
}

func GetSteam() *steam.Steam {

	steamClientLock.Lock()
	defer steamClientLock.Unlock()

	if steamClient == nil {

		steamClient = &steam.Steam{}
		steamClient.SetKey(config.Config.SteamAPIKey.Get())
		steamClient.SetUserAgent("gamedb.online")
		steamClient.SetAPIRateLimit(time.Millisecond*1000, 10)
		steamClient.SetStoreRateLimit(time.Millisecond*1800, 10)
		steamClient.SetLogger(steamLogger{})
	}

	return steamClient
}

func GetSteamUnlimited() *steam.Steam {

	steamClientUnlimitedLock.Lock()
	defer steamClientUnlimitedLock.Unlock()

	if steamClientUnlimited == nil {

		steamClientUnlimited = &steam.Steam{}
		steamClientUnlimited.SetKey(config.Config.SteamAPIKey.Get())
		steamClientUnlimited.SetUserAgent("gamedb.online")
		steamClientUnlimited.SetLogger(steamLogger{})
	}

	return steamClientUnlimited
}

func (l steamLogger) Write(i steam.Log) {
	if config.IsLocal() {
		// log.Info(i.String(), log.LogNameSteam)
	}
}

func AllowSteamCodes(err error, bytes []byte, allowedCodes []int) error {

	// if err == steam.ErrHTMLResponse {
	// 	log.Err(err, string(bytes))
	// 	time.Sleep(time.Second * 30)
	// }

	err2, ok := err.(steam.Error)
	if ok {
		if allowedCodes != nil && SliceHasInt(allowedCodes, err2.Code) {
			return nil
		}
	}
	return err
}

// Downgrade some Steam errors to info.
func LogSteamError(err error, interfaces ...interface{}) {

	isError := func() bool {

		if strings.Contains(err.Error(), "invalid character '<' looking for beginning of value") {
			return false
		}

		if strings.Contains(err.Error(), "unexpected end of JSON input") {
			return false
		}

		if strings.Contains(err.Error(), "Bad Gateway") {
			return false
		}

		if strings.Contains(err.Error(), "connection reset by peer") {
			return false
		}

		if strings.Contains(err.Error(), "Client.Timeout exceeded while awaiting headers") {
			return false
		}

		if strings.Contains(err.Error(), "TLS handshake timeout") {
			return false
		}

		if strings.Contains(err.Error(), "XML syntax error") {
			return false
		}

		if strings.Contains(err.Error(), "expected element type <memberList> but have <html>") {
			return false
		}

		return true
	}()

	interfaces = append(interfaces, err)

	if isError {
		log.Err(interfaces...)
	} else {
		log.Info(interfaces...)
	}
}

func GetAgeCheckCookieJar() (jar *cookiejar.Jar, err error) {

	cookieURL, _ := url.Parse("https://store.steampowered.com")

	jar, err = cookiejar.New(nil)
	if err != nil {
		return jar, err
	}

	jar.SetCookies(cookieURL, []*http.Cookie{
		{Name: "birthtime", Value: "536457601", Path: "/", Domain: "store.steampowered.com"},
		{Name: "lastagecheckage", Value: "1-January-1987", Path: "/", Domain: "store.steampowered.com"},
		{Name: "mature_content", Value: "1", Path: "/", Domain: "store.steampowered.com"},
	})

	return jar, err
}
