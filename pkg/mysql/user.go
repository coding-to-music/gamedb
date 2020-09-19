package mysql

import (
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"golang.org/x/crypto/bcrypt"
)

const (
	UserLevel0 = 0 // Guest
	UserLevel1 = 1 // Free
	UserLevel2 = 2 // Level 1
	UserLevel3 = 3 // Level 2
	UserLevel4 = 4 // Level 3

	// Pages
	UserLevelLimit0 = 5   // Guest
	UserLevelLimit1 = 10  // Free
	UserLevelLimit2 = 10  // Level 1
	UserLevelLimit3 = 100 // Level 2
	UserLevelLimit4 = 0   // Level 3
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
		return UserLevelLimit3 * limit
	case UserLevel4:
		return UserLevelLimit4
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
	ID            int                `gorm:"not null;column:id;primary_key"`
	CreatedAt     time.Time          `gorm:"not null;column:created_at"`
	UpdatedAt     time.Time          `gorm:"not null;column:updated_at"`
	LoggedInAt    time.Time          `gorm:"not null;column:logged_in_at;type:datetime"`
	Email         string             `gorm:"not null;column:email;unique_index"`
	EmailVerified bool               `gorm:"not null;column:email_verified"`
	Password      string             `gorm:"not null;column:password"`
	SteamID       sql.NullString     `gorm:"not null;column:steam_id"`
	PatreonID     sql.NullString     `gorm:"not null;column:patreon_id"`
	GoogleID      sql.NullString     `gorm:"not null;column:google_id"`
	DiscordID     sql.NullString     `gorm:"not null;column:discord_id"`
	GitHubID      sql.NullString     `gorm:"not null;column:github_id"`
	Level         int8               `gorm:"not null;column:level"` // Patreon
	HideProfile   bool               `gorm:"not null;column:hide_profile"`
	ShowAlerts    bool               `gorm:"not null;column:show_alerts"`
	ProductCC     steamapi.ProductCC `gorm:"not null;column:country_code"`
	APIKey        string             `gorm:"not null;column:api_key"`
}

func (user User) GetSteamID() (ret int64) {

	if user.SteamID.Valid {
		i, err := strconv.ParseInt(user.SteamID.String, 10, 64)
		if err != nil {
			log.ErrS(err)
		} else {
			return i
		}
	}
	return 0
}

func (user *User) SetAPIKey() {
	// Must match api validation regex
	user.APIKey = helpers.RandString(20, helpers.Numbers+helpers.LettersCaps)
}

func (user User) Save() error {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	db = db.Save(&user)
	return db.Error
}

func NewUser(email, password string, prodCC steamapi.ProductCC, verified bool) (user User, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return user, err
	}

	var passwordBytes []byte
	if password != "" {
		passwordBytes, err = bcrypt.GenerateFromPassword([]byte(password), 14)
		if err != nil {
			return user, err
		}
	}

	user = User{
		Email:         email,
		EmailVerified: verified,
		Password:      string(passwordBytes),
		ProductCC:     prodCC,
		Level:         UserLevel1,
		LoggedInAt:    time.Unix(0, 0), // Fixes a gorm bug
	}

	user.SetAPIKey()

	db = db.Create(&user)
	return user, db.Error
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

	db = db.Model(&user).Updates(map[string]interface{}{
		column: value,
	})
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

func GetUserByEmail(email string) (user User, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return user, err
	}

	db = db.Where("email = ?", email).First(&user)
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

func GetUserByAPIKey(key string) (user User, err error) {

	var item = memcache.MemcacheUserByAPIKey(key)

	err = memcache.GetSetInterface(item.Key, item.Expiration, &user, func() (interface{}, error) {

		return GetUserByKey("api_key", key, 0)
	})

	return user, err
}
