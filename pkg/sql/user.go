package sql

import (
	"time"

	"github.com/Jleagle/steam-go/steam"
)

type User struct {
	CreatedAt     time.Time `gorm:"not null;column:created_at"`
	UpdatedAt     time.Time `gorm:"not null;column:updated_at"`
	Email         string    `gorm:"not null;column:email;primary_key"`
	EmailVerified bool      `gorm:"not null;column:email_verified"`
	Password      string    `gorm:"not null;column:password"`
	SteamID       int64     `gorm:"not null;column:steam_id"`
	PatreonID     int64     `gorm:"not null;column:steam_id"`
	PatreonLevel  int8      `gorm:"not null;column:patreon_level"`
	HideProfile   int8      `gorm:"not null;column:hide_profile"`
	ShowAlerts    int8      `gorm:"not null;column:show_alerts"`
	CountryCode   string    `gorm:"not null;column:country_code"`
}

func GetUser(email string, mustBeVerified bool) (user User, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return user, err
	}

	db = db.Where("email = ?", email)

	if mustBeVerified {
		db = db.Where("email_verified = ?", true)
	}

	db = db.First(&user)

	return user, db.Error

}

func GetOrCreateUser(playerID int64) (user User, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return user, err
	}

	db.Attrs(User{CountryCode: string(steam.CountryUS)}).FirstOrCreate(&user, User{SteamID: playerID})

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
