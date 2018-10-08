package web

import (
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/steam-authority/steam-authority/db"
	"github.com/steam-authority/steam-authority/logger"
)

func ChangeHandler(w http.ResponseWriter, r *http.Request) {

	change, err := db.GetChange(chi.URLParam(r, "id"))
	if err != nil {
		if err.Error() == "datastore: no such entity" {
			returnErrorTemplate(w, r, 404, "We can't find this change in our database, there may not be one with this ID.")
			return
		} else {
			logger.Error(err)
			returnErrorTemplate(w, r, 500, err.Error())
			return
		}
	}

	var wg sync.WaitGroup

	// Get apps
	var apps []db.App
	wg.Add(1)
	go func() {

		apps, err = db.GetAppsByID(change.GetAppIDs(), []string{"id", "icon", "type", "name"})
		if err != nil {
			logger.Error(err)
		}

		wg.Done()

	}()

	// Get packages
	var packages []db.Package
	wg.Add(1)
	go func() {

		packages, err = db.GetPackages(change.GetPackageIDs(), []string{})
		if err != nil {
			logger.Error(err)
		}

		wg.Done()

	}()

	// Wait
	wg.Wait()

	// Template
	t := changeTemplate{}
	t.Fill(w, r, change.GetName())
	t.Change = change
	t.Apps = apps
	t.Packages = packages

	returnTemplate(w, r, "change", t)
}

type changeTemplate struct {
	GlobalTemplate
	Change   db.Change
	Apps     []db.App
	Packages []db.Package
}
