package api

import (
	"errors"
	"net/http"

	"github.com/gamedb/gamedb/cmd/webserver/pages/api/generated"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
)

func (s Server) GetPlayersId(w http.ResponseWriter, r *http.Request) {

	s.call(w, r, func(w http.ResponseWriter, r *http.Request) (code int, response interface{}) {

		if id, ok := r.Context().Value("id").(int64); ok {

			id, err := helpers.IsValidPlayerID(id)
			if err != nil {
				return 404, err
			}

			player, err := mongo.GetPlayer(id)
			if err == mongo.ErrNoDocuments {

				ua := r.UserAgent()
				err2 := queue.ProducePlayer(queue.PlayerMessage{ID: id, UserAgent: &ua})
				log.Err(err2)

				return 404, "player not found, trying to add player"

			} else if err != nil {

				log.Err(err)
				return 500, err

			} else {

				ret := generated.PlayerResponse{}
				ret.Id = player.ID
				ret.Name = player.GetName()

				return 200, ret
			}
		}

		return 400, errors.New("invalid player ID")
	})
}

func (s Server) PostPlayersId(w http.ResponseWriter, r *http.Request) {

	s.call(w, r, func(w http.ResponseWriter, r *http.Request) (code int, response interface{}) {

		if id, ok := r.Context().Value("id").(int64); ok {

			err := queue.ProducePlayer(queue.PlayerMessage{ID: id, SkipGroups: true})
			if err == memcache.ErrInQueue {
				return 200, err
			} else if err != nil {
				log.Err(err)
				return 500, err
			}

			return 200, "Player queued"
		}

		return 400, errors.New("invalid app ID")
	})
}
