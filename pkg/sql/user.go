package sql

import (
	"time"
)

type User struct {
	ID            int       `gorm:"not null;column:id;primary_key"`
	CreatedAt     time.Time `gorm:"not null;column:created_at"`
	UpdatedAt     time.Time `gorm:"not null;column:updated_at"`
	Email         string    `gorm:"not null;column:email;unique_index"`
	EmailVerified bool      `gorm:"not null;column:email_verified"`
	Password      string    `gorm:"not null;column:password"`
	SteamID       int64     `gorm:"not null;column:steam_id"`
	PatreonID     int64     `gorm:"not null;column:patreon_id"`
	PatreonLevel  int8      `gorm:"not null;column:patreon_level"`
	HideProfile   bool      `gorm:"not null;column:hide_profile"`
	ShowAlerts    bool      `gorm:"not null;column:show_alerts"`
	CountryCode   string    `gorm:"not null;column:country_code"`
}

func GetUserByID(id int) (user User, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return user, err
	}

	db = db.Where("id = ?", id).First(&user)
	return user, db.Error

}

func GetUserByEmail(email string) (user User, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return user, err
	}

	db = db.Where("email = ?", email).First(&user)
	return user, db.Error

}

func DeleteUser(id int64) (err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	db = db.Where("player_id = ?", id).Delete(&User{})

	return db.Error
}
