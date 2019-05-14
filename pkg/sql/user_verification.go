package sql

import (
	"errors"
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
)

type UserVerification struct {
	CreatedAt time.Time `gorm:"not null;column:created_at"`
	Code      string    `gorm:"not null;column:code;primary_key"`
	UserID    int       `gorm:"not null;column:user_id"`
	Expires   time.Time `gorm:"not null;column:expires"`
}

func GetUserVerification(code string) (userVerification UserVerification, err error) {

	if len(code) != 10 {
		return userVerification, errors.New("invalid code: " + code)
	}

	db, err := GetMySQLClient()
	if err != nil {
		return userVerification, err
	}

	db = db.Where("code = ?", code).Find(&userVerification)
	return userVerification, db.Error
}

func CreateUserVerification(userID int) (userVerification UserVerification, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return userVerification, err
	}

	//
	userVerification.Code = helpers.RandString(10, helpers.Letters)
	userVerification.UserID = userID
	userVerification.Expires = time.Now().AddDate(0, 0, 7)

	//
	db = db.Create(&userVerification)
	if db.Error != nil && strings.HasPrefix(db.Error.Error(), "Error 1062: Duplicate entry") {
		time.Sleep(time.Second)
		return CreateUserVerification(userID)
	}

	return userVerification, db.Error
}

func DeleteUserVerification(code string) (err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	userVerification := UserVerification{}
	userVerification.Code = code

	return db.Delete(&userVerification).Error
}
