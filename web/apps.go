package web

import (
	"net/http"
	"sync"

	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

const (
	appsSearchLimit = 96
)

func AppsHandler(w http.ResponseWriter, r *http.Request) {

	var wg sync.WaitGroup
	var err error

	// Get apps
	var apps []mysql.App
	wg.Add(1)
	go func() {

		apps, err = mysql.SearchApps(r.URL.Query(), appsSearchLimit, "id DESC", []string{})
		if err != nil {
			logger.Error(err)
		}

		wg.Done()

	}()

	// Get apps count
	var count int
	wg.Add(1)
	go func() {

		count, err = mysql.CountApps()
		if err != nil {
			logger.Error(err)
		}

		wg.Done()

	}()

	// Wait
	wg.Wait()

	// Template
	template := appsTemplate{}
	template.Fill(w, r, "Games")
	template.Apps = apps
	template.Count = count

	returnTemplate(w, r, "apps", template)
}

type appsTemplate struct {
	GlobalTemplate
	Apps  []mysql.App
	Count int
}
