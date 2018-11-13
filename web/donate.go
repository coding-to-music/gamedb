package web

import "net/http"

func DonateHandler(w http.ResponseWriter, r *http.Request) {

	t := GlobalTemplate{}
	t.Fill(w, r, "Donate")
	t.Description = "Help pay for the server costs or just buy me a beer."

	returnTemplate(w, r, "donate", t)
}
