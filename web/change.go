package web

import (
	"net/http"
	"strconv"
	"sync"

	"cloud.google.com/go/datastore"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/mongo"
	"github.com/go-chi/chi"
)

func changeHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Invaid Change ID.", Error: err})
		return
	}

	change, err := mongo.GetChange(id)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
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
	t.Apps = map[int]db.App{}
	t.Packages = map[int]db.Package{}

	//
	var wg sync.WaitGroup

	// Get apps
	wg.Add(1)
	go func() {

		defer wg.Done()

		appsSlice, err := db.GetAppsByID(change.Apps, []string{"id", "icon", "type", "name"})
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

		packagesSlice, err := db.GetPackages(change.Packages, []string{})
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
	Apps     map[int]db.App
	Packages map[int]db.Package
}
