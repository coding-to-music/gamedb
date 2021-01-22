package pages

import (
	"net/http"
	"strings"

	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
)

func FeaturesRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", featuresHandler)
	return r
}

func featuresHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	t := uniqueTemplate{}
	t.fill(w, r, "features", "Global Steam Features", "Global Steam Features")
	t.PlayerMetrics = mongo.PlayerRankFields
	t.SpecialBadges = helpers.BuiltInSpecialBadges
	t.EventBagdes = helpers.BuiltInEventBadges
	t.Countries, err = elasticsearch.AggregatePlayerCountries()
	if err != nil {
		log.ErrS(err)
	}
	t.AppCount, err = mongo.CountDocuments(mongo.CollectionApps, nil, 0)
	if err != nil {
		log.ErrS(err)
	}
	t.ArticleCount, err = mongo.CountDocuments(mongo.CollectionAppArticles, nil, 0)
	if err != nil {
		log.ErrS(err)
	}
	t.GroupsCount, err = mongo.CountDocuments(mongo.CollectionGroups, nil, 0)
	if err != nil {
		log.ErrS(err)
	}

	for k := range t.Countries {
		if strings.Contains(k, "-") {
			delete(t.Countries, k)
		}
	}

	returnTemplate(w, r, t)
}

type uniqueTemplate struct {
	globalTemplate
	PlayerMetrics map[string]mongo.RankMetric
	SpecialBadges map[int]helpers.BuiltInbadge
	EventBagdes   map[int]helpers.BuiltInbadge
	Countries     map[string]int64
	AppCount      int64
	ArticleCount  int64
	GroupsCount   int64
}
