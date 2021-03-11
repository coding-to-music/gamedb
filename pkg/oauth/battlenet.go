package oauth

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"golang.org/x/oauth2"
)

type battlenetProvider struct {
}

func (c battlenetProvider) GetName() string {
	return "Battle.net (US)"
}

func (c battlenetProvider) GetIcon() string {
	return "fab fa-battle-net"
}

func (c battlenetProvider) GetColour() string {
	return "#193354"
}

func (c battlenetProvider) GetEnum() ProviderEnum {
	return ProviderBattlenet
}

func (c battlenetProvider) GetType() ProviderType {
	return TypeOAuth
}

func (c battlenetProvider) Redirect(w http.ResponseWriter, r *http.Request, state string) {
	conf := c.GetConfig()
	http.Redirect(w, r, conf.AuthCodeURL(state), http.StatusFound)
}

func (c battlenetProvider) HasEmail() bool {
	return false
}

func (c battlenetProvider) GetUser(token *oauth2.Token) (user User, err error) {

	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+token.AccessToken)

	b, _, err := helpers.Get("https://us.battle.net/oauth/userinfo", 0, headers)
	if err != nil {
		return user, err
	}

	resp := BattlenetUser{}
	err = json.Unmarshal(b, &resp)
	if err != nil {
		return user, err
	}

	b, err = json.Marshal(token)
	if err != nil {
		log.ErrS(err)
	}

	user.Token = string(b)
	user.ID = strconv.Itoa(resp.ID)
	user.Username = resp.Battletag
	user.Avatar = "https://play-lh.googleusercontent.com/NhrJdTU7jrwvxk6riquxbLeL1sPhyXt8P5Bxkm47FZjQgcnmQsx8veLg2kIIjOGuANE=s180-rw"

	return user, nil
}

func (c battlenetProvider) GetConfig() oauth2.Config {

	return oauth2.Config{
		ClientID:     config.C.BattlenetClient,
		ClientSecret: config.C.BattlenetSecret,
		Scopes:       []string{"openid"},
		RedirectURL:  config.C.GlobalSteamDomain + "/oauth/in/" + string(c.GetEnum()),
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://us.battle.net/oauth/authorize",
			TokenURL: "https://us.battle.net/oauth/token",
		},
	}
}

type BattlenetUser struct {
	Sub       string `json:"sub"`
	ID        int    `json:"id"`
	Battletag string `json:"battletag"`
}
