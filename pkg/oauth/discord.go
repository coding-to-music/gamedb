package oauth

import (
	"encoding/json"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"golang.org/x/oauth2"
)

type discordProvider struct {
}

func (c discordProvider) GetName() string {
	return "Discord"
}

func (c discordProvider) GetIcon() string {
	return "fab fa-discord"
}

func (c discordProvider) GetColour() string {
	return "#7289DA"
}

func (c discordProvider) GetEnum() ProviderEnum {
	return ProviderDiscord
}

func (c discordProvider) GetType() ProviderType {
	return TypeOAuth
}

func (c discordProvider) Redirect(w http.ResponseWriter, r *http.Request, state string) {
	conf := c.GetConfig()
	http.Redirect(w, r, conf.AuthCodeURL(state), http.StatusFound)
}

func (c discordProvider) GetUser(token *oauth2.Token) (user User, err error) {

	discord, err := discordgo.New("Bearer " + token.AccessToken)
	if err != nil {
		return user, err
	}

	discordUser, err := discord.User("@me")
	if err != nil {
		return user, err
	}

	// if !discordUser.Verified { // Seems to always be false
	// 	err = session.SetFlash(r, helpers.SessionBad, "This Discord account has not been verified")
	// 	log.ErrS(err)
	// 	return
	// }

	b, err := json.Marshal(token)
	if err != nil {
		log.ErrS(err)
	}

	user.Token = string(b)
	user.ID = discordUser.ID
	user.Username = discordUser.Username
	user.Email = discordUser.Email
	user.Avatar = discordUser.AvatarURL("64")

	return user, nil
}

func (c discordProvider) GetConfig() oauth2.Config {

	return oauth2.Config{
		ClientID:     config.C.DiscordClientID,
		ClientSecret: config.C.DiscordClientSescret,
		Scopes:       []string{"identify", "email"},
		RedirectURL:  config.C.GameDBDomain + "/oauth/in/" + string(c.GetEnum()),
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discordapp.com/api/oauth2/authorize",
			TokenURL: "https://discordapp.com/api/oauth2/token",
		},
	}
}
