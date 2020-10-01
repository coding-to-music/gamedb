package mysql

import (
	"net/http"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/email_providers"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/geo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/oauth"
	influx "github.com/influxdata/influxdb1-client"
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
	Level         int8               `gorm:"not null;column:level"` // Patreon
	ProductCC     steamapi.ProductCC `gorm:"not null;column:country_code"`
	APIKey        string             `gorm:"not null;column:api_key"`
}

func (user *User) SetAPIKey() {
	// Must match api validation regex
	user.APIKey = helpers.RandString(20, helpers.Numbers+helpers.LettersCaps)
}

func (user User) TouchLoggedInTime() error {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	update := map[string]interface{}{
		"logged_in_at": time.Now(),
	}

	return db.Model(&user).Updates(update).Error
}

func (user User) SetPassword(b []byte) error {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	update := map[string]interface{}{
		"password": string(b),
	}

	return db.Model(&user).Updates(update).Error
}

func (user User) SetProdCC(cc steamapi.ProductCC) error {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	update := map[string]interface{}{
		"country_code": cc,
	}

	return db.Model(&user).Updates(update).Error
}

func NewUser(r *http.Request, email, password string, prodCC steamapi.ProductCC, verified bool) (user User, err error) {

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
	if db.Error != nil {
		return user, db.Error
	}

	if !verified {
		err = SendUserVerification(r, user.ID, email)
		if err != nil {
			return user, err
		}
	}

	// Create event
	err = mongo.NewEvent(r, user.ID, mongo.EventSignup)
	if err != nil {
		log.ErrS(err)
	}

	// Influx
	fields := map[string]interface{}{
		"signup": 1,
	}

	if verified {
		fields["validate"] = 1
	}

	point := influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementSignups),
		Fields:      fields,
		Time:        time.Now(),
		Precision:   "s",
	}

	_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, point)
	if err != nil {
		log.ErrS(err)
	}

	return user, nil
}

func SendUserVerification(r *http.Request, userID int, email string) error {

	// Create verification code
	code, err := CreateUserVerification(userID)
	if err != nil {
		return err
	}

	// Send email
	body := "Please click the below link to verify your email address<br />" +
		config.C.GameDBDomain + "/signup/verify?code=" + code.Code +
		"<br><br>Thanks, Jleagle." +
		"<br><br>From IP: " + geo.GetFirstIP(r.RemoteAddr)

	return email_providers.GetSender().Send(
		email,
		email,
		"",
		"",
		"Game DB Email Verification",
		body,
	)
}

func VerifyUser(userID int) error {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	update := map[string]interface{}{
		"email_verified": true,
	}

	user := User{}
	user.ID = userID

	return db.Model(&user).Updates(update).Error
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

func GetUserByAPIKey(key string) (user User, err error) {

	var item = memcache.MemcacheUserByAPIKey(key)

	err = memcache.GetSetInterface(item.Key, item.Expiration, &user, func() (interface{}, error) {

		db, err := GetMySQLClient()
		if err != nil {
			return user, err
		}

		db = db.Where("api_key = ?", key)
		db = db.First(&user)

		return user, db.Error
	})

	return user, err
}

func GetUserByProviderID(provider oauth.ProviderEnum, providerID string) (user User, err error) {

	userProvider, err := GetUserProviderByProviderID(provider, providerID)
	if err != nil {
		return user, err
	}

	return GetUserByID(userProvider.UserID)
}
