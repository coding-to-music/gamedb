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
	"github.com/gamedb/gamedb/pkg/steam"
	"github.com/yohcop/openid-go"
	"go.uber.org/zap"
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

func (c steamProvider) GetType() ProviderType {
	return TypeOpenID
}

func (c steamProvider) HasEmail() bool {
	return false
}

func (c steamProvider) Redirect(w http.ResponseWriter, r *http.Request, page string) {

	q := url.Values{}
	q.Set("openid.identity", "http://specs.openid.net/auth/2.0/identifier_select")
	q.Set("openid.claimed_id", "http://specs.openid.net/auth/2.0/identifier_select")
	q.Set("openid.ns", "http://specs.openid.net/auth/2.0")
	q.Set("openid.mode", "checkid_setup")
	q.Set("openid.realm", config.C.GameDBDomain+"/")
	q.Set("openid.return_to", config.C.GameDBDomain+"/oauth/in/steam?page="+page)

	http.Redirect(w, r, "https://steamcommunity.com/openid/login?"+q.Encode(), http.StatusFound)
}

func (c steamProvider) GetUser(r *http.Request) (user User, err error) {

	// Get Steam ID
	resp, err := openid.Verify(config.C.GameDBDomain+r.URL.String(), openid.NewSimpleDiscoveryCache(), openid.NewSimpleNonceStore())
	if err != nil {
		return user, err
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

	err = backoff.RetryNotify(operation, backoff.WithMaxRetries(policy, 5), func(err error, t time.Duration) { zap.L().Info("Finding Steam player info for oauth", zap.Error(err)) })
	if err != nil {
		return user, err
	}

	user.ID = strconv.FormatInt(int64(player.SteamID), 10)
	user.Username = player.PersonaName
	user.Avatar = player.AvatarFull

	return user, nil
}
