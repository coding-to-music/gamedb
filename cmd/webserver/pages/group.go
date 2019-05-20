package pages

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
)

func GroupRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", groupHandler)
	r.Get("/{slug}", groupHandler)
	return r
}

func groupHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid group ID"})
		return
	}

	idx, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid group ID: " + id})
		return
	}

	// if !db.IsValidAppID(idx) {
	// 	returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid bundle ID: " + id})
	// 	return
	// }

	// Get bundle
	group, err := mongo.GetGroup(idx)
	if err != nil {

		if err == sql.ErrRecordNotFound {
			returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Sorry but we can not find this group"})
			return
		}

		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the group", Error: err})
		return
	}

	t := groupTemplate{}
	t.fill(w, r, "Groups", "")
	t.Group = group
	t.Summary = helpers.RenderHTMLAndBBCode(group.Summary)

	err = returnTemplate(w, r, "group", t)
	log.Err(err, r)
}

type groupTemplate struct {
	GlobalTemplate
	Group   mongo.Group
	Summary template.HTML
}
