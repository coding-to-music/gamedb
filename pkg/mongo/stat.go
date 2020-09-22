package mongo

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/cenkalti/backoff/v4"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type Stat struct {
	Type          StatsType                      `bson:"type"`
	ID            int                            `bson:"id"`
	Name          string                         `bson:"name"`
	Apps          int                            `bson:"apps"`
	AppsPercnt    float32                        `bson:"apps_percent"`
	MeanPrice     map[steamapi.ProductCC]float32 `bson:"mean_price"`
	MeanScore     float32                        `bson:"mean_score"`
	MeanPlayers   float64                        `bson:"mean_players"`
	MedianPrice   map[steamapi.ProductCC]int     `bson:"median_price"`
	MedianScore   float32                        `bson:"median_score"`
	MedianPlayers int                            `bson:"median_players"`
}

func (stat Stat) BSON() bson.D {
	return bson.D{
		{"_id", stat.GetKey()},
		{"type", stat.Type},
		{"id", stat.ID},
		{"name", stat.Name},
		{"apps", stat.Apps},
		{"apps_percent", stat.AppsPercnt},
		{"mean_price", stat.MeanPrice},
		{"mean_score", stat.MeanScore},
		{"mean_players", stat.MeanPlayers},
		{"median_price", stat.MedianPrice},
		{"median_score", stat.MedianScore},
		{"median_players", stat.MedianPlayers},
	}
}

func (stat Stat) GetKey() string {
	return string(stat.Type) + "-" + strconv.Itoa(stat.ID)
}

func (stat Stat) GetPath() string {
	return helpers.GetStatPath(stat.Type.MongoCol(), stat.ID, stat.Name)
}

// Types
type StatsType string

const (
	StatsTypeCategories StatsType = "c"
	StatsTypeDevelopers StatsType = "d"
	StatsTypeGenres     StatsType = "g"
	StatsTypePublishers StatsType = "p"
	StatsTypeTags       StatsType = "t"
)

func (st StatsType) String() string {
	return string(st)
}

func (st StatsType) MongoCol() string {
	switch st {
	case StatsTypeCategories:
		return "categories"
	case StatsTypeDevelopers:
		return "developers"
	case StatsTypeGenres:
		return "genres"
	case StatsTypePublishers:
		return "publishers"
	case StatsTypeTags:
		return "tags"
	default:
		log.Warn("invalid stats type", zap.String("type", string(st)))
		return ""
	}
}

func (st StatsType) Title() string {
	switch st {
	case StatsTypeCategories:
		return "Category"
	case StatsTypeDevelopers:
		return "Developer"
	case StatsTypeGenres:
		return "Genre"
	case StatsTypePublishers:
		return "Publisher"
	case StatsTypeTags:
		return "Tag"
	default:
		log.Warn("invalid stats type", zap.String("type", string(st)))
		return ""
	}
}

func (st StatsType) Plural() string {
	switch st {
	case StatsTypeCategories:
		return "Categories"
	case StatsTypeDevelopers:
		return "Developers"
	case StatsTypeGenres:
		return "Genres"
	case StatsTypePublishers:
		return "Publishers"
	case StatsTypeTags:
		return "Tags"
	default:
		log.Warn("invalid stats type", zap.String("type", string(st)))
		return ""
	}
}

//
func GetStat(typex StatsType, id int) (stat Stat, err error) {

	var item = memcache.MemcacheStat(string(typex), id)

	err = memcache.GetSetInterface(item.Key, item.Expiration, &stat, func() (interface{}, error) {

		stat.Type = typex
		stat.ID = id

		err = FindOne(CollectionStats, bson.D{{"_id", stat.GetKey()}}, nil, nil, &stat)
		return stat, err
	})

	return stat, err
}

//
func GetStats(offset int64, limit int64, filter bson.D, sort bson.D) (offers []Stat, err error) {

	cur, ctx, err := Find(CollectionStats, offset, limit, sort, filter, nil, options.Find())
	if err != nil {
		return offers, err
	}

	defer close(cur, ctx)

	for cur.Next(ctx) {

		var stat Stat
		err := cur.Decode(&stat)
		if err != nil {
			log.ErrS(err)
		} else {
			offers = append(offers, stat)
		}
	}

	return offers, cur.Err()
}

func GetStatsByID(typex StatsType, ids []int) (apps []Stat, err error) {

	if len(ids) < 1 {
		return apps, nil
	}

	a := bson.A{}
	for _, v := range ids {
		a = append(a, v)
	}

	return GetStats(0, 0, bson.D{{"type", typex}, {"id", bson.M{"$in": a}}}, nil)
}

func GetStatsForSelect(typex StatsType) (stats []Stat, err error) {

	var item = memcache.MemcacheStatsForSelect(string(typex))

	err = memcache.GetSetInterface(item.Key, item.Expiration, &stats, func() (interface{}, error) {

		stats, err = GetStats(0, 200, bson.D{{"type", typex}}, bson.D{{"mean_score", -1}})

		sort.Slice(stats, func(i, j int) bool {
			return stats[i].Name < stats[j].Name
		})

		return stats, err
	})

	return stats, err
}

func BatchStats(typex StatsType, callback func(stats []Stat)) (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		stats, err := GetStats(offset, limit, bson.D{{"type", typex}}, bson.D{{"name", 1}})
		if err != nil {
			return err
		}

		callback(stats)

		if int64(len(stats)) != limit {
			break
		}

		offset += limit
	}

	return nil
}

var (
	existingTagNames = map[StatsType]map[string]int{}
	tagLock          sync.Mutex
)

func FindOrCreateStatsByName(typex StatsType, names []string) (IDs []int, err error) {

	tagLock.Lock()
	defer tagLock.Unlock()

	if _, ok := existingTagNames[typex]; !ok {
		existingTagNames[typex] = map[string]int{}
	}

	for _, name := range names {

		name = strings.TrimSpace(name)

		if val, ok := existingTagNames[typex][name]; ok {
			IDs = append(IDs, val)
			continue
		}

		// Check if it exists
		existing := Stat{}
		err = FindOne(CollectionStats, bson.D{{"type", typex}, {"name", name}}, nil, nil, &existing)
		if err == ErrNoDocuments {

			// Get highest ID to increment
			highest := Stat{}
			err = FindOne(CollectionStats, bson.D{{"type", typex}}, bson.D{{"id", -1}}, nil, &highest)
			if err != nil {
				return nil, err
			}

			newStat := Stat{}
			newStat.Type = typex
			newStat.ID = highest.ID
			newStat.Name = name

			operation := func() (err error) {
				newStat.ID++
				_, err = InsertOne(CollectionStats, newStat)
				return err
			}

			policy := backoff.NewExponentialBackOff()
			policy.InitialInterval = time.Second * 1

			err = backoff.RetryNotify(operation, backoff.WithMaxRetries(policy, 5), func(err error, t time.Duration) { log.InfoS(err) })
			if err != nil {
				return nil, err
			}

			existingTagNames[typex][name] = newStat.ID
			IDs = append(IDs, newStat.ID)
			continue

		} else if err != nil {
			return nil, err
		}

		existingTagNames[typex][name] = existing.ID
		IDs = append(IDs, existing.ID)
		continue
	}

	return IDs, nil
}

var existingTagIDs = map[StatsType]map[int]bool{}

func EnsureStat(typex StatsType, ids []int, names []string) (err error) {

	if _, ok := existingTagIDs[typex]; !ok {
		existingTagIDs[typex] = map[int]bool{}
	}

	if len(ids) != len(names) {
		return errors.New("invalid stats")
	}

	var docs []Document
	for k, v := range ids {

		if _, ok := existingTagIDs[typex][v]; ok {
			continue
		}

		docs = append(docs, Stat{
			Type: typex,
			ID:   v,
			Name: names[k],
			Apps: 1,
		})
	}

	_, err = InsertMany(CollectionStats, docs)
	return err
}
