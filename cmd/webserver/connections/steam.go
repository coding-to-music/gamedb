package connections

import (
	"net/http"
	"path"
	"strconv"

	"github.com/Jleagle/session-go/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/yohcop/openid-go"
	"golang.org/x/oauth2"
)

type steamConnection struct {
}

func (s steamConnection) getID(r *http.Request, token *oauth2.Token) interface{} {

	// Get Steam ID
	openID, err := openid.Verify(config.Config.GameDBDomain.Get()+r.URL.String(), openid.NewSimpleDiscoveryCache(), openid.NewSimpleNonceStore())
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "We could not verify your Steam account")
		log.Err(err)
		return nil
	}

	steamIDString := path.Base(openID)
	steamID, err := strconv.ParseInt(steamIDString, 10, 64)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1001)")
		log.Err(err)
		return nil
	}

	return steamID
}

func (s steamConnection) getName() string {
	return "Steam"
}

func (s steamConnection) getEnum() connectionEnum {
	return ConnectionSteam
}

func (s steamConnection) getConfig(login bool) oauth2.Config {
	return oauth2.Config{}
}

func (s steamConnection) LinkHandler(w http.ResponseWriter, r *http.Request) {

}

func (s steamConnection) UnlinkHandler(w http.ResponseWriter, r *http.Request) {
	unlink(w, r, s, mongo.EventUnlinkSteam)
}

func (s steamConnection) LinkCallbackHandler(w http.ResponseWriter, r *http.Request) {

	callback(r, s, mongo.EventLinkSteam, nil, false)

	err := session.Save(w, r)
	log.Err(err)

	http.Redirect(w, r, "/settings", http.StatusFound)
}

func (s steamConnection) LoginHandler(w http.ResponseWriter, r *http.Request) {

	// id := s.getID(r, nil)
}

func (s steamConnection) LoginCallbackHandler(w http.ResponseWriter, r *http.Request) {

	callback(r, s, mongo.EventLogin, nil, true)

	err := session.Save(w, r)
	log.Err(err)

	http.Redirect(w, r, "/login", http.StatusFound)
}
