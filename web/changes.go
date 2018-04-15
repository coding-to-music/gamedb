package web

import (
	"net/http"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

func ChangesHandler(w http.ResponseWriter, r *http.Request) {

	// Get changes
	changes, err := datastore.GetLatestChanges(50)
	if err != nil {
		logger.Error(err)
	}

	// Get apps/packages
	appIDs := make([]int, 0)
	packageIDs := make([]int, 0)
	for _, v := range changes {
		appIDs = append(appIDs, v.Apps...)
		packageIDs = append(packageIDs, v.Packages...)
	}

	// Get apps for all changes
	appsMap := make(map[int]mysql.App)
	apps, err := mysql.GetApps(appIDs, []string{"id", "name"})

	for _, v := range apps {
		appsMap[v.ID] = v
	}

	// Get packages for all changes
	packagesMap := make(map[int]mysql.Package)
	packages, err := mysql.GetPackages(packageIDs, []string{"id", "name"})

	for _, v := range packages {
		packagesMap[v.ID] = v
	}

	// todo, sort packagesMap by id

	// Template
	template := changesTemplate{}
	template.Fill(r, "Changes")
	template.Changes = changes
	template.Apps = appsMap
	template.Packages = packagesMap

	returnTemplate(w, r, "changes", template)
}

// todo, Just pass through a new struct with all the correct info instead of changes and maps to get names
type changesTemplate struct {
	GlobalTemplate
	Changes  []datastore.Change
	Apps     map[int]mysql.App
	Packages map[int]mysql.Package
}
