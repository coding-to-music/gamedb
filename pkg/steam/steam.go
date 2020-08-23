package steam

import (
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/config"
)

var (
	clientNormal     *steamapi.Client
	clientNormalLock sync.Mutex
)

func GetSteam() *steamapi.Client {

	clientNormalLock.Lock()
	defer clientNormalLock.Unlock()

	if clientNormal == nil {

		clientNormal = steamapi.NewClient()
		clientNormal.SetKey(config.C.SteamAPIKey)
		clientNormal.SetLogger(steamLogger{})
		clientNormal.SetAPIRateLimit(time.Millisecond*950, 10)
		clientNormal.SetStoreRateLimit(time.Millisecond*1750, 10)
	}

	return clientNormal
}

var (
	clientUnlimited     *steamapi.Client
	clientUnlimitedLock sync.Mutex
)

func GetSteamUnlimited() *steamapi.Client {

	clientUnlimitedLock.Lock()
	defer clientUnlimitedLock.Unlock()

	if clientUnlimited == nil {

		clientUnlimited = steamapi.NewClient()
		clientUnlimited.SetKey(config.C.SteamAPIKey)
		clientUnlimited.SetLogger(steamLogger{})
	}

	return clientUnlimited
}
