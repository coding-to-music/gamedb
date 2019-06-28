package helpers

import (
	"strings"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
)

var steamClient *steam.Steam

func GetSteam() *steam.Steam {

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

type steamLogger struct {
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

func LogSteamError(err error, interfaces ...interface{}) {

	if config.IsProd() {

		if strings.Contains(err.Error(), "invalid character '<' looking for beginning of value") {
			return
		}

		if strings.Contains(err.Error(), "unexpected end of JSON input") {
			return
		}
	}

	interfaces = append(interfaces, err)

	log.Err(interfaces)
}
