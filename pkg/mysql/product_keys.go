package mysql

import (
	"time"
)

const (
	ProductKeyTypeApp     = "app"
	ProductKeyTypePackage = "package"

	ProductKeyFieldCommon   = "common"
	ProductKeyFieldExtended = "extended"
	ProductKeyFieldUFS      = "ufs"
	ProductKeyFieldConfig   = "config"
)

type ProductKey struct {
	Type      string     `gorm:"not null;primary_key"`
	Field     string     `gorm:"not null;primary_key"`
	Key       string     `gorm:"not null;primary_key"`
	CreatedAt time.Time  `gorm:"not null"`
	UpdatedAt time.Time  `gorm:"not null"`
	DeletedAt *time.Time `gorm:""`
	Count     int        `gorm:"not null"`
}

func (key ProductKey) Save() error {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	db = db.Save(&key)
	return db.Error
}

func GetProductKeys() (keys []ProductKey, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return keys, err
	}

	db = db.Order("`key` ASC").Find(&keys)
	return keys, db.Error
}
