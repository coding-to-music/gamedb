package pages

import (
	"net/http"
	"path"
	"strconv"

	"github.com/Jleagle/recaptcha-go"
	"github.com/Jleagle/session-go/session"
	steam2 "github.com/Jleagle/steam-go/steam"
	"github.com/Jleagle/steam-go/steamid"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/steam"
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

			client := steam.GetSteam()
			id, err := steamid.ParsePlayerID(search)

			if err == nil && id > 0 {

				http.Redirect(w, r, "/players/"+strconv.FormatUint(uint64(id), 10), http.StatusFound)
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

			resp, b, err := client.ResolveVanityURL(search, steam2.VanityURLProfile)
			err = steam.AllowSteamCodes(err, b, nil)

			if err == nil && resp.Success > 0 && resp.SteamID > 0 {

				http.Redirect(w, r, "/players/"+strconv.FormatInt(int64(resp.SteamID), 10), http.StatusFound)
				return ""
			}

			return "Player " + search + " not found on Steam"
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
	t.Default = r.URL.Query().Get("search")

	//
	returnTemplate(w, r, "players_add", t)
}

type addPlayerTemplate struct {
	GlobalTemplate
	RecaptchaPublic string
	Default         string
}
