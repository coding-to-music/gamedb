package mongo

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
)

type StatsType string

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
		log.WarnS("invalid stats type")
		return ""
	}
}

const (
	StatsTypeCategories StatsType = "c"
	StatsTypeDevelopers StatsType = "d"
	StatsTypeGenres     StatsType = "g"
	StatsTypePublishers StatsType = "p"
	StatsTypeTags       StatsType = "t"
)

type Stat struct {
	Type        StatsType          `bson:"type"`
	ID          int                `bson:"id"`
	CreatedAt   time.Time          `bson:"created_at"`
	Name        string             `bson:"name"`
	Apps        int                `bson:"apps"`
	MeanPrice   map[string]float32 `bson:"mean_price"`
	MeanScore   float32            `bson:"mean_score"`
	MeanPlayers float64            `bson:"mean_players"`
}

func (stat Stat) BSON() bson.D {

	if stat.CreatedAt.IsZero() || stat.CreatedAt.Unix() == 0 {
		stat.CreatedAt = time.Now()
	}

	return bson.D{
		{"_id", stat.getKey()},
		{"type", stat.Type},
		{"id", stat.ID},
		{"created_at", stat.CreatedAt},
		{"name", stat.Name},
		{"apps", stat.Apps},
		{"mean_price", stat.MeanPrice},
		{"mean_score", stat.MeanScore},
		{"mean_players", stat.MeanPlayers},
	}
}

func (stat Stat) getKey() string {
	return string(stat.Type) + "-" + strconv.Itoa(stat.ID)
}

func GetStats(typex StatsType, offset int64, limit int64) (offers []Stat, err error) {

	sort := bson.D{{"name", 1}}
	filter := bson.D{{"type", typex}}

	cur, ctx, err := Find(CollectionStats, offset, limit, sort, filter, nil, nil)
	if err != nil {
		return offers, err
	}

	defer func() {
		err = cur.Close(ctx)
		if err != nil {
			log.ErrS(err)
		}
	}()

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

func BatchStats(typex StatsType, callback func(stats []Stat)) (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		stats, err := GetStats(typex, offset, limit)
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

func FindOrCreateStatsByName(typex StatsType, names []string) (IDs []int, err error) {

	for _, v := range names {

		v = strings.TrimSpace(v)

		stat := Stat{}
		err = FindOne(CollectionStats, bson.D{{"type", typex}, {"name", v}}, nil, bson.M{"_id": 1}, &stat)
		if err == ErrNoDocuments {

			stat.Type = typex
			stat.Name = v

			// Get highest ID to increment
			err = FindOne(CollectionStats, bson.D{{"type", typex}}, nil, bson.M{"_id": 1}, &stat)

			resp, err := InsertOne(CollectionStats, stat)
			if err != nil {
				return nil, err
			}

			var ok bool
			stat.ID, ok = resp.InsertedID.(int)
			stat.ID++
			if !ok {
				return nil, errors.New("invalid casting stat id")
			}

		} else if err != nil {
			return nil, err
		}

		IDs = append(IDs, stat.ID)
	}

	return IDs, nil
}

func FindOrCreateStatsByID(typex StatsType, names []string, ids []int) (IDs []int, err error) {

	for _, v := range names {

		v = strings.TrimSpace(v)

		stat := Stat{}
		err = FindOne(CollectionStats, bson.D{{"type", typex}, {"name", v}}, nil, bson.M{"_id": 1}, &stat)
		if err == ErrNoDocuments {

			stat.Type = typex
			stat.Name = v

			// Get highest ID to increment
			err = FindOne(CollectionStats, bson.D{{"type", typex}}, nil, bson.M{"_id": 1}, &stat)

			resp, err := InsertOne(CollectionStats, stat)
			if err != nil {
				return nil, err
			}

			var ok bool
			stat.ID, ok = resp.InsertedID.(int)
			if !ok {
				return nil, errors.New("invalid casting stat id")
			}

		} else if err != nil {
			return nil, err
		}

		IDs = append(IDs, stat.ID)
	}

	return IDs, nil
}
