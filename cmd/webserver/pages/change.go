package pages

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gamedb/website/pkg/log"
	"github.com/gamedb/website/pkg/mongo"
	"github.com/gamedb/website/pkg/sql"
	"github.com/go-chi/chi"
)

func ChangeRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", changeHandler)
	return r
}

func changeHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Invaid Change ID.", Error: err})
		return
	}

	change, err := mongo.GetChange(id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "We don't have this change in the database."})
			return
		}

		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the change.", Error: err})
		return
	}

	// Template
	t := changeTemplate{}
	t.fill(w, r, change.GetName(), "")
	t.Change = change
	t.Apps = map[int]sql.App{}
	t.Packages = map[int]sql.Package{}

	//
	var wg sync.WaitGroup

	// Get apps
	wg.Add(1)
	go func() {

		defer wg.Done()

		appsSlice, err := sql.GetAppsByID(change.Apps, []string{"id", "icon", "type", "name"})
		if err != nil {

			log.Err(err, r)
			return
		}

		for _, v := range appsSlice {
			t.Apps[v.ID] = v
		}

	}()

	// Get packages
	wg.Add(1)
	go func() {

		defer wg.Done()

		packagesSlice, err := sql.GetPackages(change.Packages, []string{})
		if err != nil {

			log.Err(err, r)
			return
		}

		for _, v := range packagesSlice {
			t.Packages[v.ID] = v
		}

	}()

	// Wait
	wg.Wait()

	err = returnTemplate(w, r, "change", t)
	log.Err(err, r)
}

type changeTemplate struct {
	GlobalTemplate
	Change   mongo.Change
	Apps     map[int]sql.App
	Packages map[int]sql.Package
}
