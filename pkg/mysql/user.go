package mysql

import (
	"net/http"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/email"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/geo"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/oauth"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserLevel uint8

const (
	UserLevelGuest UserLevel = iota // Guest
	UserLevelFree                   // Free
	UserLevel1                      // Level 1
	UserLevel2                      // Level 2
	UserLevel3                      // Level 3

	// Pages
	UserLevelLimitGuest = 5   // Guest
	UserLevelLimitFree  = 10  // Free
	UserLevelLimit1     = 10  // Level 1
	UserLevelLimit2     = 100 // Level 2
	UserLevelLimit3     = 0   // Level 3
)

func (ul UserLevel) MaxResults(limit int64) int64 {

	switch ul {
	default:
		return UserLevelLimitGuest * limit
	case UserLevelFree:
		return UserLevelLimitFree * limit
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
	ID             int                `gorm:"not null;column:id;primary_key"`
	CreatedAt      time.Time          `gorm:"not null;column:created_at"`
	UpdatedAt      time.Time          `gorm:"not null;column:updated_at"`
	LoggedInAt     *time.Time         `gorm:"column:logged_in_at;type:datetime"`
	Email          string             `gorm:"not null;column:email;unique_index"`
	EmailVerified  bool               `gorm:"not null;column:email_verified"`
	Password       string             `gorm:"not null;column:password"`
	Level          UserLevel          `gorm:"not null;column:level"`
	ProductCC      steamapi.ProductCC `gorm:"not null;column:country_code"`
	APIKey         string             `gorm:"not null;column:api_key"`
	DonatedPatreon int                `gorm:"not null;column:donated_patreon"`
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

func NewUser(r *http.Request, userEmail, password string, prodCC steamapi.ProductCC, verified bool) (user User, err error) {

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
		Email:         userEmail,
		EmailVerified: verified,
		Password:      string(passwordBytes),
		ProductCC:     prodCC,
		Level:         UserLevelFree,
	}

	user.SetAPIKey()

	db = db.Create(&user)
	if db.Error != nil {
		return user, db.Error
	}

	if verified {

		err = email.NewSignup(userEmail, r)
		if err != nil {
			return user, err
		}

	} else {

		err = SendUserVerification(user.ID, userEmail, r)
		if err != nil {
			return user, err
		}
	}

	// Create event
	go func() {
		if err = mongo.NewEvent(r, user.ID, mongo.EventSignup); err != nil {
			log.ErrS(err)
		}
	}()

	// Influx
	// go func() {
	//
	//	fields := map[string]interface{}{
	//		"signup": 1,
	//	}
	//
	//	if verified {
	//		fields["validate"] = 1
	//	}
	//
	//	point := influx.Point{
	//		Measurement: string(influxHelper.InfluxMeasurementSignups),
	//		Fields:      fields,
	//		Time:        time.Now(),
	//		Precision:   "s",
	//	}
	//
	//	if _, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, point); err != nil {
	//		log.ErrS(err)
	//	}
	// }()

	return user, nil
}

func SendUserVerification(userID int, userEmail string, r *http.Request) error {

	// Create verification code
	code, err := CreateUserVerification(userID)
	if err != nil {
		return err
	}

	// Send email
	return email.GetProvider().Send(
		userEmail,
		"",
		"",
		"Global Steam Email Verification",
		email.VerifyTemplate{
			Domain: config.C.GameDBDomain,
			Code:   code.Code,
			IP:     geo.GetFirstIP(r.RemoteAddr),
		},
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

	err = memcache.GetSetInterface(memcache.MemcacheUserByAPIKey(key), &user, func() (interface{}, error) {

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
