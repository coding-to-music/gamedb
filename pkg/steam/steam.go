package steam

import (
	"path"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/Jleagle/steam-go/steamid"
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
		clientNormal.SetTimeout(time.Second * 10)
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
		clientUnlimited.SetTimeout(time.Second * 10)
	}

	return clientUnlimited
}

type TempPlayer struct {
	ID          int64
	PersonaName string
	Avatar      string
	Level       int
	PlayTime    int
	Games       int
	Friends     int
	VACBans     int
	GameBans    int
	LastBan     time.Time
}

func GetPlayer(search string) (player TempPlayer, err error) {

	search = strings.TrimSpace(path.Base(search))

	playerID, err := steamid.ParsePlayerID(search)
	if err != nil {

		resp, err := GetSteam().ResolveVanityURL(search, steamapi.VanityURLProfile)
		err = AllowSteamCodes(err)
		if err != nil {
			return player, err
		}

		player.ID = int64(resp.SteamID)

	} else {
		player.ID = int64(playerID)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {

		defer wg.Done()

		resp, err := GetSteam().GetSteamLevel(player.ID)
		err = AllowSteamCodes(err)
		if err != nil {
			LogSteamError(err)
			return
		}

		player.Level = resp
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		summary, err := GetSteam().GetPlayer(player.ID)
		if err == steamapi.ErrProfileMissing {
			return
		}
		if err = AllowSteamCodes(err); err != nil {
			LogSteamError(err)
			return
		}

		player.PersonaName = summary.PersonaName
		player.Avatar = summary.AvatarHash
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		resp, err := GetSteam().GetOwnedGames(player.ID)
		err = AllowSteamCodes(err)
		if err != nil {
			LogSteamError(err)
			return
		}

		var playtime = 0
		for _, v := range resp.Games {
			playtime += v.PlaytimeForever
		}

		player.PlayTime = playtime
		player.Games = len(resp.Games)
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		resp, err := GetSteam().GetPlayerBans(player.ID)
		err = AllowSteamCodes(err)
		if err != nil {
			LogSteamError(err)
			return
		}

		player.LastBan = time.Now().Add(time.Hour * 24 * time.Duration(resp.DaysSinceLastBan) * -1)
		player.GameBans = resp.NumberOfGameBans
		player.VACBans = resp.NumberOfVACBans
	}()

	// 	wg.Add(1)
	// 	go func() {
	//
	// 		defer wg.Done()
	//
	// 		resp, err := GetSteam().GetFriendList(player.ID)
	// 		err = AllowSteamCodes(err, 401, 404)
	// 		if err != nil {
	// 			log.ErrS(err)
	// 			return
	// 		}
	//
	// 		player.Friends = len(resp)
	// 	}()

	wg.Wait()

	return player, nil
}
