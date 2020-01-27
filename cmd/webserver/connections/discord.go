package connections

import (
	"net/http"

	"github.com/Jleagle/session-go/session"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"golang.org/x/oauth2"
)

type discordConnection struct {
}

func (d discordConnection) getID(r *http.Request, token *oauth2.Token) interface{} {

	discord, err := discordgo.New("Bearer " + token.AccessToken)
	if err != nil {
		log.Err(err, r)
		return nil
	}

	discordUser, err := discord.User("@me")
	if err != nil {
		log.Err(err, r)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1003)")
		log.Err(err, r)
		return nil
	}

	// if !discordUser.Verified { // Seems ot always be false
	// 	err = session.SetFlash(r, helpers.SessionBad, "This Discord account has not been verified")
	// 	log.Err(err, r)
	// 	return
	// }

	return discordUser.ID
}

func (d discordConnection) getName() string {
	return "Discord"
}

func (d discordConnection) getEnum() connectionEnum {
	return ConnectionDiscord
}

func (d discordConnection) getConfig(login bool) oauth2.Config {

	var redirectURL string
	if login {
		redirectURL = config.Config.GameDBDomain.Get() + "/login/discord-callback"
	} else {
		redirectURL = config.Config.GameDBDomain.Get() + "/settings/discord-callback"
	}

	return oauth2.Config{
		ClientID:     config.Config.DiscordClientID.Get(),
		ClientSecret: config.Config.DiscordClientSescret.Get(),
		Scopes:       []string{"identify"},
		RedirectURL:  redirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discordapp.com/api/oauth2/authorize",
			TokenURL: "https://discordapp.com/api/oauth2/token",
		},
	}
}

func (d discordConnection) LinkHandler(w http.ResponseWriter, r *http.Request) {
	linkOAuth(w, r, d, false)
}

func (d discordConnection) UnlinkHandler(w http.ResponseWriter, r *http.Request) {
	unlink(w, r, d, mongo.EventUnlinkDiscord)
}

func (d discordConnection) LinkCallbackHandler(w http.ResponseWriter, r *http.Request) {

	callbackOAuth(r, d, mongo.EventLinkDiscord, false)

	err := session.Save(w, r)
	log.Err(err, r)

	http.Redirect(w, r, "/settings", http.StatusFound)
}

func (d discordConnection) LoginHandler(w http.ResponseWriter, r *http.Request) {
	linkOAuth(w, r, d, true)
}

func (d discordConnection) LoginCallbackHandler(w http.ResponseWriter, r *http.Request) {

	callbackOAuth(r, d, mongo.EventLogin, true)

	http.Redirect(w, r, "/login", http.StatusFound)
}
