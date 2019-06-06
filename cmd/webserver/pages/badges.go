package pages

import (
	"net/http"
	"sort"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
)

func BadgesRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", badgesHandler)
	r.Mount("/{id}", BadgeRouter())
	return r
}

func badgesHandler(w http.ResponseWriter, r *http.Request) {

	t := badgesTemplate{}
	t.fill(w, r, "Badges", "")

	// var err error
	// t.PlayersMap, err = getBadgeCounts()
	// if err != nil {
	// 	returnErrorTemplate(w, r, errorTemplate{Code: 500})
	// 	return
	// }

	err := returnTemplate(w, r, "badges", t)
	log.Err(err, r)
}

type badgesTemplate struct {
	GlobalTemplate
	// PlayersMap map[int]badgeRowTemplate
}

func (bt badgesTemplate) GetSpecialBadges() (badges []mongo.PlayerBadge) {

	for _, v := range mongo.Badges {
		if v.IsSpecial() {
			badges = append(badges, v)
		}
	}

	sort.Slice(badges, func(i, j int) bool {
		return badges[i].BadgeID > badges[j].BadgeID
	})

	return badges
}

func (bt badgesTemplate) GetEventBadges() (badges []mongo.PlayerBadge) {

	for _, v := range mongo.Badges {
		if !v.IsSpecial() {
			badges = append(badges, v)
		}
	}

	sort.Slice(badges, func(i, j int) bool {
		return badges[i].AppID > badges[j].AppID
	})

	return badges
}

// type badgeRowTemplate struct {
// 	Players int
// 	Max     int
// 	MaxFoil int
// }

// func getBadgeCounts() (counts map[int]badgeRowTemplate, err error) {
//
// 	var item = helpers.MemcacheTrendingAppsCount
//
// 	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &counts, func() (interface{}, error) {
//
// 		var counts = map[int]badgeRowTemplate{}
//
// 		for _, v := range mongo.Badges {
// 			if v.IsSpecial() {
// 				counts[v.BadgeID] = badgeRowTemplate{
//
// 				}
// 			} else {
// 				counts[v.AppID] = badgeRowTemplate{
//
// 				}
// 			}
// 		}
//
// 		return counts, nil
// 	})
//
// 	return counts, err
// }
