package sql

import (
	"time"

	"github.com/gamedb/website/pkg/helpers"
)

type Donation struct {
	CreatedAt time.Time `gorm:"created_at"`
	PlayerID  int64     `gorm:"player_id"`
	Amount    int       `gorm:"amount"`
	AmountUSD int       `gorm:"amount_usd"`
	Currency  string    `gorm:"currency"`
}

func (d Donation) GetCreatedNice() (ret string) {
	return d.CreatedAt.Format(helpers.DateYear)
}

func (d Donation) GetCreatedUnix() int64 {
	return d.CreatedAt.Unix()
}
