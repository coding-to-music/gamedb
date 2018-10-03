package web

import (
	"net/http"
	"sync"

	"cloud.google.com/go/datastore"
	"github.com/steam-authority/steam-authority/db"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/structs"
)

const (
	changesLimit = 100
)

func ChangesHandler(w http.ResponseWriter, r *http.Request) {

	// Get changes
	var changes []structs.ChangesChangeTemplate
	resp, err := db.GetLatestChanges(changesLimit, 1)
	logger.Error(err)

	for _, v := range resp {
		changes = append(changes, structs.ChangesChangeTemplate{
			Change: v,
		})
	}

	//
	var wg = sync.WaitGroup{}

	// Get apps
	wg.Add(1)
	go func() {

		// Get app IDs
		var appIDs []int
		for _, v := range changes {
			appIDs = append(appIDs, v.Change.GetAppIDs()...)
		}

		// Get apps for all changes
		appsMap := make(map[int]db.App)
		apps, err := db.GetApps(appIDs, []string{"id", "name", "icon"})
		logger.Error(err)

		// Make app map
		for _, v := range apps {
			appsMap[v.ID] = v
		}

		// Add app to changes
		for k, v := range changes {

			for _, vv := range v.Change.GetAppIDs() {

				if val, ok := appsMap[vv]; ok {

					changes[k].Apps = append(changes[k].Apps, val)
				}
			}
		}

		wg.Done()

	}()

	// Get packages
	wg.Add(1)
	go func() {

		// Get package IDs
		var packageIDs []int
		for _, v := range changes {
			packageIDs = append(packageIDs, v.Change.GetPackageIDs()...)
		}

		// Get packages for all changes
		packagesMap := make(map[int]db.Package)
		packages, err := db.GetPackages(packageIDs, []string{"id", "name"})
		logger.Error(err)

		// Make package map
		for _, v := range packages {
			packagesMap[v.ID] = v
		}

		// Add app to changes
		for k, v := range changes {

			for _, vv := range v.Change.GetAppIDs() {

				if val, ok := packagesMap[vv]; ok {

					changes[k].Packages = append(changes[k].Packages, val)
				}
			}
		}

		wg.Done()

	}()

	// Wait
	wg.Wait()

	// Template
	t := changesTemplate{}
	t.Fill(w, r, "Changes")

	returnTemplate(w, r, "changes", t)
}

type changesTemplate struct {
	GlobalTemplate
}

func ChangesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	query.FillFromURL(r.URL.Query())

	var changes []db.Change

	client, ctx, err := db.GetDSClient()
	if err != nil {

		logger.Error(err)

	} else {

		q := datastore.NewQuery(db.KindChange).Limit(100).Order("-change_id")

		q, err = query.SetOrderOffsetDS(q, map[string]string{})
		if err != nil {

			logger.Error(err)

		} else {

			_, err := client.GetAll(ctx, q, &changes)
			logger.Error(err)
		}
	}

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = "10000"
	response.RecordsFiltered = "10000"
	response.Draw = query.Draw

	for _, v := range changes {

		response.AddRow(v.OutputForJSON())
	}

	response.output(w)
}
