package pages

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

func ChangeRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", changeHandler)
	return r
}

func changeHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Invaid Change ID."})
		return
	}

	change, err := mongo.GetChange(id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "We don't have this change in the database."})
			return
		}

		log.ErrS(r, err)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the change."})
		return
	}

	// Template
	t := changeTemplate{}
	t.fill(w, r, "change", change.GetName(), "Steam change")
	t.Change = change
	t.Apps = map[int]mongo.App{}
	t.Packages = map[int]mongo.Package{}
	t.Canonical = change.GetPath()

	//
	var wg sync.WaitGroup

	// Get apps
	wg.Add(1)
	go func() {

		defer wg.Done()

		apps, err := mongo.GetAppsByID(change.Apps, bson.M{"_id": 1, "icon": 1, "type": 1, "name": 1})
		if err != nil {

			log.ErrS(err)
			return
		}

		for _, v := range apps {
			t.Apps[v.ID] = v
		}
	}()

	// Get packages
	wg.Add(1)
	go func() {

		defer wg.Done()

		packagesSlice, err := mongo.GetPackagesByID(change.Packages, bson.M{})
		if err != nil {

			log.ErrS(err)
			return
		}

		for _, v := range packagesSlice {
			t.Packages[v.ID] = v
		}
	}()

	// Get previous
	wg.Add(1)
	go func() {

		defer wg.Done()

		err = mongo.FindOne(mongo.CollectionChanges, bson.D{{"_id", bson.M{"$lt": change.ID}}}, bson.D{{"_id", -1}}, bson.M{"_id": 1}, &t.Previous)
		if err != nil {
			err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
			log.ErrS(err)
		}
	}()

	// Get next
	wg.Add(1)
	go func() {

		defer wg.Done()

		err = mongo.FindOne(mongo.CollectionChanges, bson.D{{"_id", bson.M{"$gt": change.ID}}}, bson.D{{"_id", 1}}, bson.M{"_id": 1}, &t.Next)
		if err != nil {
			err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
			if err != nil {
				log.ErrS(err)
			}
		}
	}()

	// Wait
	wg.Wait()

	returnTemplate(w, r, t)
}

type changeTemplate struct {
	globalTemplate
	Change   mongo.Change
	Apps     map[int]mongo.App
	Packages map[int]mongo.Package
	Next     mongo.Change
	Previous mongo.Change
}
