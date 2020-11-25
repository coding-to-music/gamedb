package pages

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

func FriendsRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/friends.json", friendsJSONHandler)
	return r
}

func friendsJSONHandler(w http.ResponseWriter, r *http.Request) {

	//goland:noinspection GoPreferNilSlice
	var ret = []helpers.Tuple{}

	defer func() {
		b, _ := json.Marshal(ret)
		_, _ = w.Write(b)
	}()

	playerID := session.GetPlayerIDFromSesion(r)
	if playerID == 0 {
		return
	}

	callback := func() (interface{}, error) {
		return mongo.GetFriends(playerID, 0, 0, bson.D{{"name", 1}}, bson.D{{"name", bson.M{"$ne": ""}}})
	}

	var friends []mongo.PlayerFriend
	var item = memcache.MemcachePlayerFriends(playerID)
	err := memcache.GetSetInterface(item.Key, item.Expiration, &friends, callback)
	if err != nil {
		log.ErrS(err)
	}

	for _, v := range friends {
		ret = append(ret, helpers.Tuple{Key: strconv.FormatInt(v.FriendID, 10), Value: v.GetName()})
	}
}
