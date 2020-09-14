package pages

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/mysql/pics"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

func PackageRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", packageHandler)
	r.Get("/prices.json", packagePricesAjaxHandler)
	r.Get("/{slug}", packageHandler)
	return r
}

func packageHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Invalid package ID"})
		return
	}

	// Get package
	pack, err := mongo.GetPackage(id)
	if err != nil {

		if err == mongo.ErrNoDocuments {
			returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Sorry but we can not find this package."})
			return
		}

		err = helpers.IgnoreErrors(err, mongo.ErrInvalidPackageID)
		if err != nil {
			log.ErrS(err)
		}

		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the package."})
		return
	}

	//
	var wg sync.WaitGroup

	var appsMap = map[int]mongo.App{}
	var appsSlice []mongo.App
	wg.Add(1)
	go func() {

		defer wg.Done()

		// Get apps
		for _, v := range pack.Apps {
			appsMap[v] = mongo.App{ID: v}
		}

		appsSlice, err = mongo.GetAppsByID(pack.Apps, bson.M{"_id": 1, "name": 1, "icon": 1, "type": 1, "platforms": 1, "dlc": 1, "common": 1, "background": 1})
		if err != nil {
			log.ErrS(err)
			return
		}

		for _, v := range appsSlice {
			appsMap[v.ID] = v
		}

		// Add missing apps to queue
		var missingAppIDs []int
		for _, v := range appsMap {
			if v.Name == "" {
				missingAppIDs = append(missingAppIDs, v.ID)
			}
		}

		err = queue.ProduceSteam(queue.SteamMessage{AppIDs: missingAppIDs})
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		if err != nil {
			log.ErrS(err)
		}
	}()

	var bundles []mysql.Bundle
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		bundles, err = GetPackageBundles(pack)
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Wait
	wg.Wait()

	// Make banners
	banners := make(map[string][]string)
	var primary []string

	// if pack.GetExtended() == "prerelease" {
	// 	primary = append(primary, "This package is intended for developers and publishers only.")
	// }

	if len(primary) > 0 {
		banners["primary"] = primary
	}

	// Template
	t := packageTemplate{}
	if len(appsSlice) == 1 {
		t.setBackground(appsSlice[0], true, true)
	}
	t.fill(w, r, pack.GetName(), "Steam package")
	t.metaImage = pack.GetMetaImage()
	t.addAssetHighCharts()
	t.IncludeSocialJS = true
	t.Package = pack
	t.Apps = appsMap
	t.Bundles = bundles
	t.Canonical = pack.GetPath()

	// Update news, reviews etc
	func() {

		if helpers.IsBot(r.UserAgent()) {
			return
		}

		if !pack.ShouldUpdate() {
			return
		}

		err = queue.ProduceSteam(queue.SteamMessage{PackageIDs: []int{pack.ID}})
		if err == nil {
			t.addToast(Toast{Title: "Update", Message: "Package has been queued for an update", Success: true})
			log.Info("package queued", zap.String("ua", r.UserAgent()))
		}
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Functions that get called multiple times in the template
	t.Price = pack.Prices.Get(session.GetProductCC(r))
	t.Controller = pack.Controller
	t.Extended = t.Package.Extended.Formatted(pack.ID, pics.ExtendedKeys)

	//
	returnTemplate(w, r, "package", t)
}

type packageTemplate struct {
	globalTemplate
	Apps       map[int]mongo.App
	Bundles    []mysql.Bundle
	Banners    map[string][]string
	Controller pics.PICSController
	Extended   []pics.KeyValue
	Package    mongo.Package
	Price      helpers.ProductPrice
}

func (t packageTemplate) includes() []string {
	return []string{"includes/social.gohtml"}
}

func (p packageTemplate) ShowDev() bool {

	return len(p.Extended) > 0 || len(p.Controller) > 0 || len(p.Package.Depots) > 0
}

func packagePricesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	productPricesAjaxHandler(w, r, helpers.ProductTypePackage)
}
