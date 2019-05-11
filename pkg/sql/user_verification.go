package sql

import (
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
)

type UserVerification struct {
	CreatedAt time.Time `gorm:"not null;column:created_at"`
	Email     string    `gorm:"not null;column:email"`
	Code      string    `gorm:"not null;column:code;primary_key"`
	Expires   time.Time `gorm:"not null;column:expires"`
}

func CreateUserVerification(email string) (userVerification UserVerification, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return userVerification, err
	}

	//
	userVerification.Code = helpers.RandString(10, helpers.Letters)
	userVerification.Email = email
	userVerification.Expires = time.Now().AddDate(0, 0, 7)

	//
	db = db.Create(&userVerification)
	if db.Error != nil && strings.HasPrefix(db.Error.Error(), "Error 1062: Duplicate entry") {
		time.Sleep(time.Second)
		return CreateUserVerification(email)
	}

	return userVerification, db.Error
}
