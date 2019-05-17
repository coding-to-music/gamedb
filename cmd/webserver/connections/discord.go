package connections

import (
	"net/http"

	"github.com/Jleagle/session-go/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"golang.org/x/oauth2"
)

type discord struct {
}

func (d discord) getID(r *http.Request, token *oauth2.Token) interface{} {

	discord, err := helpers.GetDiscordBot(token.AccessToken, false)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "Invalid token")
		log.Err(err)
		return nil
	}

	discordUser, err := discord.User("@me")
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1003)")
		log.Err(err)
		return nil
	}

	// if !discordUser.Verified { // Seems ot always be false
	// 	err = session.SetFlash(r, helpers.SessionBad, "This Discord account has not been verified")
	// 	log.Err(err)
	// 	return
	// }

	return discordUser.ID
}

func (d discord) getName() string {
	return "Discord"
}

func (d discord) getEnum() connectionEnum {
	return ConnectionDiscord
}

func (d discord) getConfig() oauth2.Config {
	return oauth2.Config{
		ClientID:     config.Config.DiscordClientID.Get(),
		ClientSecret: config.Config.DiscordClientSescret.Get(),
		Scopes:       []string{"identify"},
		RedirectURL:  config.Config.GameDBDomain.Get() + "/settings/discord-callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discordapp.com/api/oauth2/authorize",
			TokenURL: "https://discordapp.com/api/oauth2/token",
		},
	}
}

func (d discord) getEmptyVal() interface{} {
	return ""
}

func (d discord) LinkHandler(w http.ResponseWriter, r *http.Request) {
	linkOAuth(w, r, d)
}

func (d discord) UnlinkHandler(w http.ResponseWriter, r *http.Request) {
	unlink(w, r, d, mongo.EventUnlinkDiscord)
}

func (d discord) CallbackHandler(w http.ResponseWriter, r *http.Request) {

	callbackOAuth(w, r, d, mongo.EventLinkDiscord)

	err := session.Save(w, r)
	log.Err(err)

	http.Redirect(w, r, "/settings", http.StatusFound)
}
