package oauth

import (
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/cenkalti/backoff/v4"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/steam"
	"github.com/yohcop/openid-go"
	"golang.org/x/oauth2"
)

type steamProvider struct {
}

func (c steamProvider) GetName() string {
	return "Steam"
}

func (c steamProvider) GetIcon() string {
	return "fab fa-steam"
}

func (c steamProvider) GetColour() string {
	return "#000000"
}

func (c steamProvider) GetEnum() ProviderEnum {
	return ProviderSteam
}

func (c steamProvider) Redirect(w http.ResponseWriter, r *http.Request, state string) {

	q := url.Values{}
	q.Set("openid.identity", "http://specs.openid.net/auth/2.0/identifier_select")
	q.Set("openid.claimed_id", "http://specs.openid.net/auth/2.0/identifier_select")
	q.Set("openid.ns", "http://specs.openid.net/auth/2.0")
	q.Set("openid.mode", "checkid_setup")
	q.Set("openid.realm", config.C.GameDBDomain+"/")
	q.Set("openid.return_to", config.C.GameDBDomain+"/oauth/in/steam?page="+state)

	http.Redirect(w, r, "https://steamcommunity.com/openid/login?"+q.Encode(), http.StatusFound)
}

func (c steamProvider) GetUser(r *http.Request, _ *oauth2.Token) (user User, err error) {

	// Get Steam ID
	resp, err := openid.Verify(config.C.GameDBDomain+r.URL.String(), openid.NewSimpleDiscoveryCache(), openid.NewSimpleNonceStore())
	if err != nil {
		return user, OauthError{err, "We could not verify your Steam account"}
	}

	i, err := strconv.ParseInt(path.Base(resp), 10, 64)
	if err != nil {
		return user, err
	}

	var player steamapi.PlayerSummary

	operation := func() (err error) {

		player, err = steam.GetSteamUnlimited().GetPlayer(i)
		if err == steamapi.ErrProfileMissing {
			return backoff.Permanent(err)
		}
		return err
	}

	policy := backoff.NewExponentialBackOff()

	err = backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.InfoS(err) })
	if err != nil {
		log.ErrS(err)
		return user, err
	}

	user.ID = strconv.FormatInt(int64(player.SteamID), 10)
	user.Username = player.PersonaName
	user.Avatar = player.AvatarFull

	return user, nil
}
