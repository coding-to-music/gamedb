package pages

import (
	"net/http"
	"strconv"
	"sync"

	main2 "github.com/gamedb/website/cmd/consumers"
	"github.com/gamedb/website/pkg"
	"github.com/go-chi/chi"
)

func bundleRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", bundleHandler)
	r.Get("/{slug}", bundleHandler)
	return r
}

func bundleHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

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

	// if !db.IsValidAppID(idx) {
	// 	returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid bundle ID: " + id})
	// 	return
	// }

	// Get bundle
	bundle, err := pkg.GetBundle(idx, []string{})
	if err != nil {

		if err == pkg.ErrRecordNotFound {
			returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Sorry but we can not find this bundle."})
			return
		}

		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the bundle.", Error: err})
		return
	}

	// Template
	t := bundleTemplate{}
	t.fill(w, r, bundle.Name, "")
	t.Bundle = bundle

	//
	var wg sync.WaitGroup

	// Get apps
	wg.Add(1)
	go func(bundle pkg.Bundle) {

		defer wg.Done()

		appIDs, err := bundle.GetAppIDs()
		if err != nil {
			log.Err(err, r)
			return
		}

		t.Apps, err = pkg.GetAppsByID(appIDs, []string{})
		log.Err(err, r)

		// Queue missing apps
		if len(appIDs) != len(t.Apps) {
			for _, v := range appIDs {
				var found = false
				for _, vv := range t.Apps {
					if v == vv.ID {
						found = true
						break
					}
				}

				if !found {
					err = main2.ProduceApp(v)
					log.Err()
				}
			}
		}

	}(bundle)

	// Get packages
	wg.Add(1)
	go func() {

		defer wg.Done()

		appIDs, err := bundle.GetPackageIDs()
		if err != nil {
			log.Err(err, r)
			return
		}

		t.Packages, err = pkg.GetPackages(appIDs, []string{})
		log.Err(err, r)

	}()

	// Wait
	wg.Wait()

	err = returnTemplate(w, r, "bundle", t)
	log.Err(err, r)
}

type bundleTemplate struct {
	GlobalTemplate
	Bundle   pkg.Bundle
	Apps     []sql.App
	Packages []pkg.Package
}
