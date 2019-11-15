package pages

import (
	"net/http"
	"sync"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
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

	count, err := mongo.CountDocuments(mongo.CollectionGroups, nil, 0)
	log.Err(err, r)

	t.Count = helpers.ShortHandNumber(count)

	returnTemplate(w, r, "groups", t)
}

type groupsTemplate struct {
	GlobalTemplate
	Count string
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
	filter := bson.D{}

	search := helpers.RegexNonAlphaNumericSpace.ReplaceAllString(query.getSearchString("search"), "")
	if len(search) > 0 {
		filter = append(filter, bson.E{Key: "$or", Value: bson.A{
			bson.M{"name": bson.M{"$regex": search, "$options": "i"}},
			bson.M{"abbreviation": bson.M{"$regex": search, "$options": "i"}},
			bson.M{"url": bson.M{"$regex": search, "$options": "i"}},
		}})
	}

	typex := query.getSearchString("type")
	if typex == "group" || typex == "game" {
		filter = append(filter, bson.E{Key: "type", Value: typex})
	}

	showErrors := query.getSearchString("errors")
	if showErrors == "removed" {
		filter = append(filter, bson.E{Key: "error", Value: bson.M{"$exists": true, "$ne": ""}})
	} else if showErrors == "notremoved" {
		filter = append(filter, bson.E{Key: "error", Value: bson.M{"$exists": true, "$eq": ""}})
	}

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
			group.ID64,                         // 0
			group.GetName(),                    // 1
			group.GetPath(),                    // 2
			group.GetIcon(),                    // 3
			group.Headline,                     // 4
			group.Members,                      // 5
			group.URL,                          // 6
			group.Type,                         // 7
			group.GetURL(),                     // 8
			group.Error != "",                  // 9
			helpers.TrendValue(group.Trending), // 10
			group.ID,                           // 11
		})
	}

	response.output(w, r)
}
