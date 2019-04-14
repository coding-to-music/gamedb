package web

import (
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/Jleagle/recaptcha-go"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
)

func playerAddHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour*24)

	if r.Method == http.MethodPost {

		message := func() string {

			// Parse form
			err := r.ParseForm()
			if err != nil {
				return err.Error()
			}

			search := r.PostFormValue("search")
			if search == "" {
				return "Please enter a search term"
			}

			search = path.Base(search)

			steam := helpers.GetSteam()
			id, err := steam.GetID(search)

			if err == nil && id > 0 {

				http.Redirect(w, r, "/players/"+strconv.FormatInt(id, 10), http.StatusFound)
				return ""
			}

			// Recaptcha
			err = recaptcha.CheckFromRequest(r)
			if err != nil {

				if err == recaptcha.ErrNotChecked {
					return "Please check the captcha"
				}

				return err.Error()
			}

			resp, b, err := steam.ResolveVanityURL(search, 1)
			err = helpers.HandleSteamStoreErr(err, b, nil)

			if err == nil && resp.Success > 0 && resp.SteamID > 0 {

				http.Redirect(w, r, "/players/"+strconv.FormatInt(int64(resp.SteamID), 10), http.StatusFound)
				return ""
			}

			return "Player " + search + "not found on Steam"
		}()

		if message != "" {
			err := session.SetBadFlash(w, r, message)
			log.Err(err)
			http.Redirect(w, r, "/players/add", http.StatusFound)
			return
		}
	}

	t := addPlayerTemplate{}
	t.fill(w, r, "Add Player", "Add yourself to the Steam DB.")
	t.RecaptchaPublic = config.Config.RecaptchaPublic

	//
	err := returnTemplate(w, r, "add_player", t)
	log.Err(err, r)
}

type addPlayerTemplate struct {
	GlobalTemplate
	RecaptchaPublic string
}
