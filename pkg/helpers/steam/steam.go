package steam

import (
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/config"
)

var (
	clientNormal     *steamapi.Steam
	clientNormalLock sync.Mutex
)

func GetSteam() *steamapi.Steam {

	clientNormalLock.Lock()
	defer clientNormalLock.Unlock()

	if clientNormal == nil {

		clientNormal = &steamapi.Steam{}
		clientNormal.SetKey(config.Config.SteamAPIKey.Get())
		clientNormal.SetLogger(steamLogger{})
		clientNormal.SetAPIRateLimit(time.Millisecond*950, 10)
		clientNormal.SetStoreRateLimit(time.Millisecond*1750, 10)
	}

	return clientNormal
}

var (
	clientUnlimited     *steamapi.Steam
	clientUnlimitedLock sync.Mutex
)

func GetSteamUnlimited() *steamapi.Steam {

	clientUnlimitedLock.Lock()
	defer clientUnlimitedLock.Unlock()

	if clientUnlimited == nil {

		clientUnlimited = &steamapi.Steam{}
		clientUnlimited.SetKey(config.Config.SteamAPIKey.Get())
		clientUnlimited.SetLogger(steamLogger{})
	}

	return clientUnlimited
}
