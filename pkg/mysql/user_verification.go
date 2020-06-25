package mysql

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

func GetUserVerification(code string) (userID int, err error) {

	if len(code) != 10 {
		return userID, errors.New("invalid code: " + code)
	}

	db, err := GetMySQLClient()
	if err != nil {
		return userID, err
	}
	userVerification := UserVerification{}
	db = db.Where("code = ?", code).Find(&userVerification)
	if db.Error != nil {
		return userVerification.UserID, db.Error
	}

	return userVerification.UserID, deleteUserVerification(code)
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

func deleteUserVerification(code string) (err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	userVerification := UserVerification{}
	userVerification.Code = code

	err = db.Delete(&userVerification).Error
	err = helpers.IgnoreErrors(err, ErrRecordNotFound)

	return err
}
