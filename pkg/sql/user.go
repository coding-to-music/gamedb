package sql

import (
	"time"

	"github.com/Jleagle/steam-go/steam"
)

type User struct {
	CreatedAt    time.Time `gorm:"not null;column:created_at"`
	UpdatedAt    time.Time `gorm:"not null;column:updated_at"`
	PlayerID     int64     `gorm:"not null;column:player_id;primary_key"`
	Email        string    `gorm:"not null;column:email;index:email"`
	Verified     bool      `gorm:"not null;column:verified"`
	Password     string    `gorm:"not null;column:password"`
	HideProfile  int8      `gorm:"not null;column:hide_profile"`
	ShowAlerts   int8      `gorm:"not null;column:show_alerts"`
	CountryCode  string    `gorm:"not null;column:country_code"`
	PatreonLevel int8      `gorm:"not null;column:patreon_level"`
}

func GetUsersByEmail(email string) (users []User, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return users, err
	}

	db = db.Limit(100).Where("email = (?)", email).Order("created_at ASC").Find(&users)
	if db.Error != nil {
		return users, db.Error
	}

	return users, nil
}

func GetOrCreateUser(playerID int64) (user User, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return user, err
	}

	db.Attrs(User{CountryCode: string(steam.CountryUS)}).FirstOrCreate(&user, User{PlayerID: playerID})

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
