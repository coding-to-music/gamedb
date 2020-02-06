package steam

import (
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/config"
)

var (
	clientNormal     *steam.Steam
	clientNormalLock sync.Mutex
)

func GetSteam() *steam.Steam {

	clientNormalLock.Lock()
	defer clientNormalLock.Unlock()

	if clientNormal == nil {

		clientNormal = &steam.Steam{}
		clientNormal.SetKey(config.Config.SteamAPIKey.Get())
		clientNormal.SetUserAgent("gamedb.online")
		clientNormal.SetAPIRateLimit(time.Millisecond*950, 10)
		clientNormal.SetStoreRateLimit(time.Millisecond*1750, 10)
		clientNormal.SetLogger(steamLogger{})
	}

	return clientNormal
}

var (
	clientUnlimited     *steam.Steam
	clientUnlimitedLock sync.Mutex
)

func GetSteamUnlimited() *steam.Steam {

	clientUnlimitedLock.Lock()
	defer clientUnlimitedLock.Unlock()

	if clientUnlimited == nil {

		clientUnlimited = &steam.Steam{}
		clientUnlimited.SetKey(config.Config.SteamAPIKey.Get())
		clientUnlimited.SetUserAgent("gamedb.online")
		clientUnlimited.SetLogger(steamLogger{})
	}

	return clientUnlimited
}
