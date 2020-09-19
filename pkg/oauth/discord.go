package oauth

import (
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
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
	return "7289DA"
}

func (c discordProvider) GetEnum() ProviderEnum {
	return ProviderDiscord
}

func (c discordProvider) GetConfig() oauth2.Config {

	return oauth2.Config{
		ClientID:     config.C.DiscordClientID,
		ClientSecret: config.C.DiscordClientSescret,
		Scopes:       []string{"identify"},
		RedirectURL:  config.C.GameDBDomain + "/oauth/in/" + string(c.GetEnum()),
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discordapp.com/api/oauth2/authorize",
			TokenURL: "https://discordapp.com/api/oauth2/token",
		},
	}
}

func (c discordProvider) GetUser(_ *http.Request, token *oauth2.Token) (user User, err error) {

	discord, err := discordgo.New("Bearer " + token.AccessToken)
	if err != nil {
		return user, err
	}

	discordUser, err := discord.User("@me")
	if err != nil {
		return user, OauthError{err, "An error occurred (1003)"}
	}

	// if !discordUser.Verified { // Seems to always be false
	// 	err = session.SetFlash(r, helpers.SessionBad, "This Discord account has not been verified")
	// 	log.ErrS(err)
	// 	return
	// }

	user.Token = token.AccessToken
	user.ID = discordUser.ID
	user.Username = discordUser.Username
	user.Email = discordUser.Email
	user.Avatar = discordUser.AvatarURL("64")

	return user, nil
}
