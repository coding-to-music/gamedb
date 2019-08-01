package sql

import (
	"errors"
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/helpers"
)

const (
	UserLevel0 = 0 // Free
	UserLevel1 = 1
	UserLevel2 = 2
	UserLevel3 = 3

	// Pages
	UserLevelLimit0 = 10 // Free
	UserLevelLimit1 = 10
	UserLevelLimit2 = 0
	UserLevelLimit3 = 0
)

type UserLevel int

func (ul UserLevel) MaxResults(limit int64) int64 {

	switch ul {
	default:
		return UserLevelLimit0 * limit
	case UserLevel1:
		return UserLevelLimit1 * limit
	case UserLevel2:
		return UserLevelLimit2 * limit
	case UserLevel3:
		return UserLevelLimit3
	}
}

func (ul UserLevel) MaxOffset(limit int64) int64 {

	results := ul.MaxResults(limit)
	if results == 0 {
		return 0
	}
	return results - limit
}

type User struct {
	ID            int             `gorm:"not null;column:id;primary_key"`
	CreatedAt     time.Time       `gorm:"not null;column:created_at"`
	UpdatedAt     time.Time       `gorm:"not null;column:updated_at"`
	Email         string          `gorm:"not null;column:email;unique_index"`
	EmailVerified bool            `gorm:"not null;column:email_verified"`
	Password      string          `gorm:"not null;column:password"`
	SteamID       int64           `gorm:"not null;column:steam_id"`
	PatreonID     string          `gorm:"not null;column:patreon_id"`
	GoogleID      string          `gorm:"not null;column:google_id"`
	DiscordID     string          `gorm:"not null;column:discord_id"`
	PatreonLevel  int8            `gorm:"not null;column:patreon_level"`
	HideProfile   bool            `gorm:"not null;column:hide_profile"`
	ShowAlerts    bool            `gorm:"not null;column:show_alerts"`
	ProductCC     steam.ProductCC `gorm:"not null;column:country_code"`
	APIKey        string          `gorm:"not null;column:api_key"`
}

func UpdateUserCol(userID int, column string, value interface{}) (err error) {

	if userID == 0 {
		return errors.New("invalid user id: " + strconv.Itoa(userID))
	}

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	var user = User{ID: userID}

	db = db.Model(&user).Update(column, value)
	return db.Error
}

func GetUserByID(id int) (user User, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return user, err
	}

	db = db.Where("id = ?", id).First(&user)
	return user, db.Error
}

func GetUserByKey(key string, value interface{}, excludeUserID int) (user User, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return user, err
	}

	db = db.Where(key+" = ?", value)
	if excludeUserID > 0 {
		db = db.Where("id != ?", excludeUserID)
	}
	db = db.First(&user)

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

func GetUserFromKeyCache(key string) (user User, err error) {

	var item = helpers.MemcacheUserLevelByKey(key)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &user, func() (interface{}, error) {

		return GetUserByKey("api_key", key, 0)
	})

	return user, err
}
