package pages

import (
	"net/http"
	"sync"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
)

func GroupsRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", groupsHandler)
	r.Get("/trending", groupsTrendingHandler)
	r.Get("/trending/trending.json", groupsTrendingAjaxHandler)
	r.Get("/groups.json", groupsAjaxHandler)
	r.Mount("/{id}", GroupRouter())
	return r
}

func groupsHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	t := groupsTemplate{}
	t.fill(w, r, "Groups", "A database of all Steam groups")

	err = returnTemplate(w, r, "groups", t)
	log.Err(err, r)
}

func groupsTrendingHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	t := groupsTemplate{}
	t.fill(w, r, "Trending Groups", "A database of all Steam groups")

	err = returnTemplate(w, r, "groups_trending", t)
	log.Err(err, r)
}

type groupsTemplate struct {
	GlobalTemplate
}

func filterGroups(query DataTablesQuery) (filter mongo.M) {

	filter = mongo.M{}

	search := query.getSearchString("search")
	if len(search) >= 2 {
		filter["$or"] = mongo.A{
			mongo.M{"$text": mongo.M{"$search": search}},
			mongo.M{"_id": search},
			mongo.M{"id": search},
		}
	}

	typ := query.getSearchString("type")
	if typ == "group" || typ == "game" {
		filter["type"] = typ
	}

	showErrors := query.getSearchString("errors")
	if showErrors == "removed" {
		filter["error"] = mongo.M{"$exists": true, "$ne": ""}
	} else if showErrors == "notremoved" {
		filter["error"] = mongo.M{"$exists": true, "$eq": ""}
	}

	return filter
}

func groupsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	if err != nil {
		log.Err(err)
		return
	}

	query.limit(r)

	var filter = filterGroups(query)

	//
	var wg sync.WaitGroup

	// Get groups
	var groups []mongo.Group
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		groups, err = mongo.GetGroups(100, query.getOffset64(), mongo.D{{"members", -1}}, filter, nil)
		if err != nil {
			log.Err(err, r)
			return
		}

		for k := range groups {
			groups[k].Name = helpers.InsertNewLines(groups[k].Name, 20)
			groups[k].Headline = helpers.InsertNewLines(groups[k].Headline, 10)
		}
	}(r)

	// Get total
	var total int64
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionGroups, filter, 0)
		log.Err(err, r)

	}(r)

	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = total
	response.RecordsFiltered = total
	response.Draw = query.Draw
	response.limit(r)

	for _, v := range groups {
		response.AddRow(v.OutputForJSON())
	}

	response.output(w, r)
}

func groupsTrendingAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	if err != nil {
		log.Err(err)
		return
	}

	query.limit(r)

	var filter = filterGroups(query)

	filter["trending"] = mongo.M{"$ne": 0, "$exists": true}

	//
	var wg sync.WaitGroup

	// Get groups
	var groups []mongo.Group
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		groups, err = mongo.GetGroups(100, query.getOffset64(), mongo.D{{"trending", -1}}, filter, nil)
		if err != nil {
			log.Err(err, r)
			return
		}

		for k := range groups {
			groups[k].Name = helpers.InsertNewLines(groups[k].Name, 20)
			groups[k].Headline = helpers.InsertNewLines(groups[k].Headline, 10)
		}
	}(r)

	// Get total
	var total int64
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionGroups, filter, 0)
		log.Err(err, r)

	}(r)

	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = total
	response.RecordsFiltered = total
	response.Draw = query.Draw
	response.limit(r)

	for _, v := range groups {
		response.AddRow(v.OutputForJSON())
	}

	response.output(w, r)
}
