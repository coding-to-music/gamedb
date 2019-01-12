package web

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
	"github.com/go-chi/chi"
)

func bundleHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid bundle ID."})
		return
	}

	idx, err := strconv.Atoi(id)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid bundle ID: " + id})
		return
	}

	//if !db.IsValidAppID(idx) {
	//	returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid bundle ID: " + id})
	//	return
	//}

	// Get bundle
	bundle, err := db.GetBundle(idx, []string{})
	if err != nil {

		if err == db.ErrRecordNotFound {
			returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Sorry but we can not find this bundle."})
			return
		}

		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the bundle.", Error: err})
		return
	}

	// Redirect to correct slug
	if r.URL.Path != bundle.GetPath() {
		http.Redirect(w, r, bundle.GetPath(), 302)
		return
	}

	// Template
	t := bundleTemplate{}
	t.Fill(w, r, bundle.Name, "")
	t.Bundle = bundle

	//
	var wg sync.WaitGroup

	// Get apps
	wg.Add(1)
	go func(bundle db.Bundle) {

		defer wg.Done()

		appIDs, err := bundle.GetAppIDs()
		if err != nil {
			log.Err(err)
			return
		}

		t.Apps, err = db.GetAppsByID(appIDs, []string{})
		log.Err(err)

	}(bundle)

	// Get packages
	wg.Add(1)
	go func() {

		defer wg.Done()

		appIDs, err := bundle.GetPackageIDs()
		if err != nil {
			log.Err(err)
			return
		}

		t.Packages, err = db.GetPackages(appIDs, []string{})
		log.Err(err)

	}()

	// Wait
	wg.Wait()

	err = returnTemplate(w, r, "bundle", t)
	log.Err(err, r)
}

type bundleTemplate struct {
	GlobalTemplate
	Bundle   db.Bundle
	Apps     []db.App
	Packages []db.Package
}
