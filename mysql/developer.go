package mysql

import (
	"strconv"
	"time"
)

type Developer struct {
	ID        int        `gorm:"not null;column:id;primary_key;AUTO_INCREMENT"`
	CreatedAt *time.Time `gorm:"not null;column:created_at"`
	UpdatedAt *time.Time `gorm:"not null;column:updated_at"`
	Name      string     `gorm:"not null;column:name"`
}

func (d Developer) GetPath() string {
	return "/apps?developer=" + strconv.Itoa(d.ID)
}

func GetAllDevelopers() (developers []Developer, err error) {

	db, err := GetDB()
	if err != nil {
		return developers, err
	}

	db = db.Limit(1000).Order("name ASC").Find(&developers)
	if db.Error != nil {
		return developers, db.Error
	}

	return developers, nil
}
