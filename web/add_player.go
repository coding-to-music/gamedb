package web

import (
	"errors"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/Jleagle/recaptcha-go"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
)

func init() {
	recaptcha.SetSecret(config.Config.RecaptchaPrivate)
}

func playerAddHandler(w http.ResponseWriter, r *http.Request) {

	setCacheHeaders(w, time.Hour*24)

	if r.Method == http.MethodPost {

		err := func() (err error) {

			// Parse form
			err = r.ParseForm()
			if err != nil {
				return err
			}

			search := r.PostFormValue("id")
			search = path.Base(search)

			steam := helpers.GetSteam()
			id, err := steam.GetID(search)

			if err == nil && id > 0 {

				http.Redirect(w, r, "/players/"+strconv.FormatInt(id, 10), 302)
				return
			}

			// Recaptcha
			err = recaptcha.CheckFromRequest(r)
			if err != nil {

				if err == recaptcha.ErrNotChecked {
					return errors.New("please check the captcha")
				}

				return err
			}

			resp, b, err := steam.ResolveVanityURL(search, 1)
			err = helpers.HandleSteamStoreErr(err, b, nil)

			if err == nil && resp.Success > 0 && resp.SteamID > 0 {

				http.Redirect(w, r, "/players/"+strconv.FormatInt(int64(resp.SteamID), 10), 302)
				return
			}

			return errors.New("player not found")
		}()

		log.Err(err)
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
