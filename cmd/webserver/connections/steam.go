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

type steam struct {
}

func (s steam) getID(r *http.Request, token *oauth2.Token) interface{} {

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

func (s steam) getName() string {
	return "Steam"
}

func (s steam) getEnum() connectionEnum {
	return ConnectionSteam
}

func (s steam) getConfig() oauth2.Config {
	return oauth2.Config{}
}

func (s steam) getEmptyVal() interface{} {
	return 0
}

func (s steam) LinkHandler(w http.ResponseWriter, r *http.Request) {

}

func (s steam) UnlinkHandler(w http.ResponseWriter, r *http.Request) {
	unlink(w, r, s, mongo.EventUnlinkSteam)
}

func (s steam) CallbackHandler(w http.ResponseWriter, r *http.Request) {

	callback(w, r, s, mongo.EventLinkSteam, nil)

	err := session.Save(w, r)
	log.Err(err)

	http.Redirect(w, r, "/settings", http.StatusFound)
}
