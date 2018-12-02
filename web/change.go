package web

import (
	"net/http"
	"strconv"
	"sync"

	"cloud.google.com/go/datastore"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
	"github.com/go-chi/chi"
)

func changeHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Invaid Change ID.", Error: err})
		return
	}

	change, err := db.GetChange(id)
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
	t.Fill(w, r, change.GetName(), "")
	t.Change = change
	t.Apps = map[int]db.App{}
	t.Packages = map[int]db.Package{}

	//
	var wg sync.WaitGroup

	// Get apps
	wg.Add(1)
	go func() {

		defer wg.Done()

		for _, v := range change.Apps {
			t.Apps[v.ID] = db.App{ID: v.ID, Name: v.Name}
		}

		appsSlice, err := db.GetAppsByID(change.GetAppIDs(), []string{"id", "icon", "type", "name"})
		if err != nil {

			log.Log(err)
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

		for _, v := range change.Packages {
			t.Packages[v.ID] = db.Package{ID: v.ID, PICSName: v.Name}
		}

		packagesSlice, err := db.GetPackages(change.GetPackageIDs(), []string{})
		if err != nil {

			log.Log(err)
			return
		}

		for _, v := range packagesSlice {
			t.Packages[v.ID] = v
		}

	}()

	// Wait
	wg.Wait()

	err = returnTemplate(w, r, "change", t)
	log.Log(err)
}

type changeTemplate struct {
	GlobalTemplate
	Change   db.Change
	Apps     map[int]db.App
	Packages map[int]db.Package
}
