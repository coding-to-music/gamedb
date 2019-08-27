package tasks

import (
	"encoding/json"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/log"
)

type statsRow struct {
	name       string
	count      int
	totalPrice map[steam.ProductCC]int
	totalScore float64
}

func (t statsRow) getMeanPrice() string {

	means := map[steam.ProductCC]float64{}

	for code, total := range t.totalPrice {
		means[code] = float64(total) / float64(t.count)
	}

	bytes, err := json.Marshal(means)
	log.Err(err)

	return string(bytes)
}

func (t statsRow) getMeanScore() float64 {
	return t.totalScore / float64(t.count)
}
