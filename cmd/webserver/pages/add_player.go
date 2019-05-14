package pages

import (
	"net/http"
	"path"
	"strconv"

	"github.com/Jleagle/recaptcha-go"
	"github.com/Jleagle/session-go/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
)

func playerAddHandler(w http.ResponseWriter, r *http.Request) {

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
			if config.IsProd() {
				err = recaptcha.CheckFromRequest(r)
				if err != nil {

					if err == recaptcha.ErrNotChecked {
						return "Please check the captcha"
					}

					return err.Error()
				}
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

			err := session.SetFlash(r, helpers.SessionBad, message)
			log.Err(err)

			err = session.Save(w, r)
			log.Err(err)

			http.Redirect(w, r, "/players/add", http.StatusFound)
			return
		}
	}

	t := addPlayerTemplate{}
	t.fill(w, r, "Add Player", "Add yourself to the Steam DB.")
	t.RecaptchaPublic = config.Config.RecaptchaPublic.Get()
	t.setFlashes(w, r, true)

	//
	err := returnTemplate(w, r, "add_player", t)
	log.Err(err, r)
}

type addPlayerTemplate struct {
	GlobalTemplate
	RecaptchaPublic string
}
