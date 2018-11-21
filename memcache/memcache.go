package memcache

import (
	"encoding/json"
	"strconv"

	"github.com/Jleagle/steam-go/steam"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/logging"
	"github.com/spf13/viper"
)

const namespace = "game-db-"

var ErrCacheMiss = memcache.ErrCacheMiss
var client *memcache.Client
var day int32 = 86400
var (
	// Counts
	AppsCount         = memcache.Item{Key: "apps-count", Expiration: day}
	FreeAppsCount     = memcache.Item{Key: "free-apps-count", Expiration: day}
	PackagesCount     = memcache.Item{Key: "packages-count", Expiration: day}
	RanksCount        = memcache.Item{Key: "ranks-count", Expiration: day}
	CountPlayers      = memcache.Item{Key: "players-count", Expiration: day * 7}
	PlayerEventsCount = func(playerID int64) memcache.Item {
		return memcache.Item{Key: "players-events-count-" + strconv.FormatInt(playerID, 10), Expiration: day}
	}

	// Dropdowns
	TagKeyNames       = memcache.Item{Key: "tag-key-names", Expiration: day * 7}
	GenreKeyNames     = memcache.Item{Key: "genre-key-names", Expiration: day * 7}
	PublisherKeyNames = memcache.Item{Key: "publisher-key-names", Expiration: day * 7}
	DeveloperKeyNames = memcache.Item{Key: "developer-key-names", Expiration: day * 7}
	AppTypes          = memcache.Item{Key: "app-types", Expiration: day * 7}

	// Rows
	ChangeRow = func(changeID int64) memcache.Item {
		return memcache.Item{Key: "change-" + strconv.FormatInt(changeID, 10), Expiration: day * 30}
	}
	ConfigRow = func(key string) memcache.Item {
		return memcache.Item{Key: "config-item-" + key, Expiration: 0}
	}

	// Other
	MostExpensiveApp = func(code steam.CountryCode) memcache.Item {
		return memcache.Item{Key: "most-expensive-app-" + string(code), Expiration: day * 7}
	}
	PlayerRefreshed = func(playerID int64) memcache.Item {
		return memcache.Item{Key: "player-refreshed-" + strconv.FormatInt(playerID, 10), Expiration: 0, Value: []byte("x")}
	}
)

// Called from main
func Init() {
	getClient()
}

func getClient() *memcache.Client {

	if client == nil {
		client = memcache.New(viper.GetString("MEMCACHE_DSN"))
	}

	return client
}

// Returns []byte
func Get(key string, i interface{}) error {

	client := getClient()

	item, err := client.Get(namespace + key)
	if err != nil {
		logging.Info("Memcache miss: " + key)
		return err
	}

	return helpers.Unmarshal(item.Value, i)
}

func Set(key string, value interface{}, expiration int32) error {

	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	client := getClient()
	item := new(memcache.Item)
	item.Key = namespace + key
	item.Value = bytes
	item.Expiration = expiration

	return client.Set(item)
}

func SetItem(item memcache.Item) error {
	return Set(item.Key, item.Value, item.Expiration)
}

func GetSetInt(item memcache.Item, f func() (j int, err error)) (count int, err error) {

	err = Get(item.Key, &count)

	if err != nil && (err == memcache.ErrCacheMiss || err.Error() == "EOF") {

		count, err := f()
		if err != nil {
			return count, err
		}

		err = Set(item.Key, count, item.Expiration)
		return count, err
	}

	return count, err
}

func GetSetString(item memcache.Item, f func() (j string, err error)) (s string, err error) {

	err = Get(item.Key, &s)

	if err != nil && (err == memcache.ErrCacheMiss || err.Error() == "EOF") {

		s, err := f()
		if err != nil {
			return s, err
		}

		err = Set(item.Key, s, item.Expiration)
		return s, err
	}

	return s, err
}

func Delete(item memcache.Item) (err error) {

	client := getClient()
	err = client.Delete(namespace + item.Key)
	return helpers.IgnoreErrors(err, memcache.ErrCacheMiss)
}

func Inc(key string, delta ...uint64) (err error) {

	if len(delta) == 0 {
		delta = []uint64{1}
	}

	client := getClient()
	_, err = client.Increment(namespace+key, delta[0])

	return err
}

func Dec(key string, delta ...uint64) (err error) {

	if len(delta) == 0 {
		delta = []uint64{1}
	}

	client := getClient()
	_, err = client.Decrement(namespace+key, delta[0])

	return err
}

func DeleteAll() (err error) {
	return getClient().DeleteAll()
}
