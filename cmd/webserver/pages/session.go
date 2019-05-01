package pages

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gamedb/gamedb/pkg/log"
)

// func SessionRouter() http.Handler {
// 	r := chi.NewRouter()
// 	r.Get("/", SessionHandler)
// 	return r
// }

func SessionHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, 0)

	t := GlobalTemplate{}
	t.fill(w, r, "", "")

	stringMap := map[string]interface{}{
		"contactPage":        t.contactPage,
		"flashesBad":         t.flashesBad,
		"flashesGood":        t.flashesGood,
		"isAdmin":            t.isAdmin(),
		"isLocal":            t.isLocal(),
		"isLoggedIn":         t.isLoggedIn(),
		"loginPage":          t.loginPage,
		"showAds":            t.showAds(),
		"toasts":             t.toasts,
		"userCountry":        t.UserCountry,
		"userCurrencySymbol": t.UserCurrencySymbol,
		"userEmail":          t.userEmail,
		"userID":             strconv.Itoa(t.userID), // Too long for JS int
		"userLevel":          t.userLevel,
		"userName":           t.userName,
	}

	b, err := json.Marshal(stringMap)
	log.Err(err, r)

	err = returnJSON(w, r, b)
	log.Err(err, r)
}
