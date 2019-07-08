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
	r.Get("/groups.json", groupsTrendingAjaxHandler)
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

type groupsTemplate struct {
	GlobalTemplate
}

func groupsTrendingAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	if err != nil {
		log.Err(err)
		return
	}

	query.limit(r)

	// Filter
	filter := mongo.M{}

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

	filter["trending"] = mongo.M{"$ne": 0, "$exists": true}

	//
	var wg sync.WaitGroup

	// Get groups
	var groups []mongo.Group
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		columns := map[string]string{
			"1": "members",
			"2": "trending",
		}

		groups, err = mongo.GetGroups(100, query.getOffset64(), query.getOrderMongo(columns, nil), filter, nil)
		if err != nil {
			log.Err(err, r)
			return
		}

		for k := range groups {
			groups[k].Name = helpers.InsertNewLines(groups[k].Name)
			groups[k].Headline = helpers.InsertNewLines(groups[k].Headline)
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

	for _, group := range groups {
		response.AddRow([]interface{}{
			group.ID64,        // 0
			group.GetName(),   // 1
			group.GetPath(),   // 2
			group.GetIcon(),   // 3
			group.Headline,    // 4
			group.Members,     // 5
			group.URL,         // 6
			group.Type,        // 7
			group.GetLink(),   // 8
			group.Error != "", // 9
			group.Trending,    // 10
		})
	}

	response.output(w, r)
}
