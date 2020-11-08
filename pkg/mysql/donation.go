package mysql

import (
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
)

type DonationSource string

const (
	DonationSourcePatreon = "patreon"
	DonationSourceManual  = "manual"
)

type Donation struct {
	ID               int       `gorm:"not null;column:id;primary_key;auto_increment"`
	CreatedAt        time.Time `gorm:"not null;column:created_at"`
	UserID           int       `gorm:"not null;column:user_id"`
	PlayerID         int64     `gorm:"not null;column:player_id"`
	Email            string    `gorm:"not null;column:email"`
	AmountUSD        int       `gorm:"not null;column:amount_usd"`
	OriginalCurrency string    `gorm:"not null;column:original_currency"`
	OriginalAmount   int       `gorm:"not null;column:original_amount"`
	Source           string    `gorm:"not null;column:source"`
	Anon             bool      `gorm:"not null;column:anon"`
	PatreonRef       string    `gorm:"column:patreon_ref"` // Nullable, indexed
}

func (d Donation) Format() string {
	return helpers.FloatToString(float64(d.AmountUSD)/100, 2)
}

func GetDonationsByUser(userID int, offset int) (donations []Donation, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return donations, err
	}

	err = db.
		Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(100).
		Offset(offset).
		Find(&donations).Error

	return donations, err
}

func GetDonationCountByUser(userID int) (count int, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return count, err
	}

	err = db.Model(&Donation{}).Where("user_id = ?", userID).Count(&count).Error

	return count, err
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

	return donations, db.Error
}

type GroupedDonation struct {
	PlayerID  int64 `json:"player_id"`
	Donations int   `json:"donations"`
}

func (d GroupedDonation) Format() string {
	return helpers.FloatToString(float64(d.Donations)/100, 2)
}
