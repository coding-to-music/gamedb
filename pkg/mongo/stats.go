package mongo

import (
	"time"
)

type Stat struct {
	ID          int                `bson:"_id"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
	Name        string             `bson:"name"`
	Apps        int                `bson:"apps"`
	MeanPrice   map[string]float32 `bson:"mean_price"`
	MeanScore   float32            `bson:"mean_score"`
	MeanPlayers int                `bson:"mean_players"`
}

func (t Stat) BSON() (ret interface{}) {

	t.UpdatedAt = time.Now()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}

	return M{
		"_id":          t.ID,
		"created_at":   t.CreatedAt,
		"updated_at":   t.UpdatedAt,
		"name":         t.Name,
		"apps":         t.Apps,
		"mean_price":   t.MeanPrice,
		"mean_score":   t.MeanScore,
		"mean_players": t.MeanPlayers,
	}
}

func GetStats(c collection, offset int64, limit int64) (stats []Stat) {
	return stats
}

func UpdateStats(c collection, stats []Stat) {

	// batch upsert

}

func FindOrCreate(c collection, name string) (id int) {
	return 0
}
