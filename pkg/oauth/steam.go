package oauth

import (
	"net/http"
	"path"

	"github.com/gamedb/gamedb/pkg/config"
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
	return "000"
}

func (c steamProvider) GetEnum() ProviderEnum {
	return ProviderSteam
}

func (c steamProvider) GetConfig() oauth2.Config {
	return oauth2.Config{}
}

func (c steamProvider) GetUser(r *http.Request, _ *oauth2.Token) (user User, err error) {

	// Get Steam ID
	resp, err := openid.Verify(config.C.GameDBDomain+r.URL.String(), openid.NewSimpleDiscoveryCache(), openid.NewSimpleNonceStore())
	if err != nil {
		return user, OauthError{err, "We could not verify your Steam account"}
	}

	// todo
	user.ID = path.Base(resp)
	// user.Username = discordUser.Username
	// user.Email = discordUser.Email
	// user.Avatar = discordUser.AvatarURL("64")

	return user, nil
}
