package sql

import (
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
)

type Donation struct {
	ID               int       `gorm:"not null;column:id;primary_key;AUTO_INCREMENT"`
	CreatedAt        time.Time `gorm:"not null;column:created_at"`
	PlayerID         int64     `gorm:"not null;column:player_id"`
	Email            string    `gorm:"not null;column:email"`
	AmountUSD        float64   `gorm:"not null;column:amount_usd"`
	OriginalCurrency string    `gorm:"not null;column:original_currency"`
	OriginalAmount   float64   `gorm:"not null;column:original_amount"`
}

func (d Donation) Format() string {
	return helpers.FloatToString(d.AmountUSD, 2)
}

func LatestDonations() (donations []Donation, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return donations, err
	}

	db = db.Where("amount_usd > ?", 0).Order("created_at desc").Limit(100).Find(&donations)
	if db.Error != nil {
		return donations, db.Error
	}

	return donations, nil
}

type GroupedDonation struct {
	PlayerID  int64   `json:"player_id"`
	Donations float64 `json:"donations"`
}

func (d GroupedDonation) Format() string {
	return helpers.FloatToString(d.Donations, 2)
}

func TopDonators() (donations []GroupedDonation, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return donations, err
	}

	db = db.Table("donations").
		Select("player_id, sum(amount_usd) as donations").
		Group("player_id").
		Order("donations desc").
		Scan(&donations)

	if db.Error != nil {
		return donations, db.Error
	}

	return donations, nil
}
