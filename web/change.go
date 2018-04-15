package web

import (
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

func ChangeHandler(w http.ResponseWriter, r *http.Request) {

	change, err := datastore.GetChange(chi.URLParam(r, "id"))
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
	var apps []mysql.App
	wg.Add(1)
	go func() {

		apps, err = mysql.GetApps(change.Apps, []string{"id", "icon", "type", "name"})
		if err != nil {
			logger.Error(err)
		}

		wg.Done()

	}()

	// Get packages
	var packages []mysql.Package
	wg.Add(1)
	go func() {

		packages, err = mysql.GetPackages(change.Packages, []string{})
		if err != nil {
			logger.Error(err)
		}

		wg.Done()

	}()

	// Wait
	wg.Wait()

	// Template
	template := changeTemplate{}
	template.Fill(r, change.GetName())
	template.Change = change
	template.Apps = apps
	template.Packages = packages

	returnTemplate(w, r, "change", template)
}

type changeTemplate struct {
	GlobalTemplate
	Change   *datastore.Change
	Apps     []mysql.App
	Packages []mysql.Package
}
