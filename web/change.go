package web

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/logging"
	"github.com/go-chi/chi"
)

func ChangeHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		returnErrorTemplate(w, r, 404, "Invaid Change ID")
		return
	}

	change, err := db.GetChange(id)
	if err != nil {
		if err.Error() == "datastore: no such entity" {
			returnErrorTemplate(w, r, 404, "We can't find this change in our database, there may not be one with this ID.")
			return
		} else {
			logging.Error(err)
			returnErrorTemplate(w, r, 500, err.Error())
			return
		}
	}

	// Template
	t := changeTemplate{}
	t.Fill(w, r, change.GetName())
	t.Change = change

	//
	var wg sync.WaitGroup

	// Get apps
	wg.Add(1)
	go func() {

		for _, v := range change.Apps {
			t.Apps[v.ID] = db.App{ID: v.ID, Name: v.Name}
		}

		appsSlice, err := db.GetAppsByID(change.GetAppIDs(), []string{"id", "icon", "type", "name"})
		if err != nil {

			logging.Error(err)

		} else {

			for _, v := range appsSlice {
				t.Apps[v.ID] = v
			}

		}

		wg.Done()
	}()

	// Get packages
	wg.Add(1)
	go func() {

		for _, v := range change.Packages {
			t.Packages[v.ID] = db.Package{ID: v.ID, PICSName: v.Name}
		}

		packagesSlice, err := db.GetPackages(change.GetPackageIDs(), []string{})
		if err != nil {

			logging.Error(err)

		} else {

			for _, v := range packagesSlice {
				t.Packages[v.ID] = v
			}

		}

		wg.Done()
	}()

	// Wait
	wg.Wait()

	returnTemplate(w, r, "change", t)
}

type changeTemplate struct {
	GlobalTemplate
	Change   db.Change
	Apps     map[int]db.App
	Packages map[int]db.Package
}
