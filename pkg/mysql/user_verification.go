package mysql

import (
	"errors"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	mysql2 "github.com/go-sql-driver/mysql"
)

type UserVerification struct {
	CreatedAt time.Time `gorm:"not null;column:created_at"`
	Code      string    `gorm:"not null;column:code;primary_key"`
	UserID    int       `gorm:"not null;column:user_id"`
	Expires   time.Time `gorm:"not null;column:expires"`
}

var ErrExpiredVerification = errors.New("verification code expired")

func GetUserVerification(code string) (userID int, err error) {

	if len(code) != 10 {
		return userID, errors.New("invalid code: " + code)
	}

	db, err := GetMySQLClient()
	if err != nil {
		return userID, err
	}

	row := UserVerification{}
	db = db.Where("code = ?", code).Find(&row)
	if db.Error != nil {
		return row.UserID, db.Error
	}

	// if row.Expires.Unix() < time.Now().Unix() {
	// 	return userID, ErrExpiredVerification
	// }

	return row.UserID, deleteUserVerification(code)
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

	if val, ok := db.Error.(*mysql2.MySQLError); ok && val.Number == 1062 {
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

	err = db.Where("code = ?", code).Delete(UserVerification{}).Error
	err = helpers.IgnoreErrors(err, ErrRecordNotFound)

	return err
}
