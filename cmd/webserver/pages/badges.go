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
	t.setRandomBackground()

	err := returnTemplate(w, r, "badges", t)
	log.Err(err, r)
}

type badgesTemplate struct {
	GlobalTemplate
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
