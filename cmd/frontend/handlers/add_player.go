package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/Jleagle/steam-go/steamid"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/steam"
)

func playerAddHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {

		message := func() string {

			// Recaptcha
			// if config.IsProd() {
			// 	err = recaptcha.CheckFromRequest(r)
			// 	if err != nil {
			//
			// 		if err == recaptcha.ErrNotChecked {
			// 			return "Please check the captcha"
			// 		}
			//
			// 		return err.Error()
			// 	}
			// }

			// Parse form
			err := r.ParseForm()
			if err != nil {
				return err.Error()
			}

			search := r.PostFormValue("search")
			if search == "" {
				return "Please enter a search term"
			}

			search = strings.TrimSpace(search)

			split := strings.Split(search, "/id/")
			if len(split) > 1 {
				search = split[1]
			}

			split = strings.Split(search, "/profiles/")
			if len(split) > 1 {
				search = split[1]
			}

			search = strings.Split(search, "/")[0]

			// Check if search term is a Steam ID
			id, err := steamid.ParsePlayerID(search)
			if err == nil && id > 0 {
				http.Redirect(w, r, "/players/"+fmt.Sprint(id), http.StatusFound)
				return ""
			}

			// Check in Steam API
			resp, err := steam.GetSteam().ResolveVanityURL(search, steamapi.VanityURLProfile)
			err = steam.AllowSteamCodes(err)
			if err != nil {
				log.ErrS(err)
			}

			if resp.SteamID > 0 {
				http.Redirect(w, r, "/players/"+fmt.Sprint(resp.SteamID), http.StatusFound)
				return ""
			}

			return "Player " + search + " not found on Steam"
		}()

		if message != "" {

			session.SetFlash(r, session.SessionBad, message)
			session.Save(w, r)

			http.Redirect(w, r, "/players/add", http.StatusFound)
			return
		}
	}

	t := addPlayerTemplate{}
	t.fill(w, r, "players_add", "Add Player", "Start tracking your stats in Global Steam.")
	t.RecaptchaPublic = config.C.RecaptchaPublic
	t.Default = r.URL.Query().Get("search")

	//
	returnTemplate(w, r, t)
}

type addPlayerTemplate struct {
	globalTemplate
	RecaptchaPublic string
	Default         string
}
