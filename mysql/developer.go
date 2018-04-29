package mysql

import (
	"strconv"
	"time"

	"github.com/steam-authority/steam-authority/helpers"
)

type Developer struct {
	ID           int        `gorm:"not null;column:id;primary_key;AUTO_INCREMENT"`
	CreatedAt    *time.Time `gorm:"not null;column:created_at"`
	UpdatedAt    *time.Time `gorm:"not null;column:updated_at"`
	DeletedAt    *time.Time `gorm:"not null;column:deleted_at"`
	Name         string     `gorm:"not null;column:name"`
	Apps         int        `gorm:"not null;column:apps"`
	MeanPrice    float64    `gorm:"not null;column:mean_price"`
	MeanDiscount float64    `gorm:"not null;column:mean_discount"`
}

func (d Developer) GetPath() string {
	return "/games?developer=" + strconv.Itoa(d.ID)
}

func (d Developer) GetMeanPrice() float64 {
	return helpers.CentsFloat(d.MeanPrice)
}

func (d Developer) GetMeanDiscount() float64 {
	return helpers.DollarsFloat(d.MeanDiscount)
}

func GetAllDevelopers() (developers []Developer, err error) {

	db, err := GetDB()
	if err != nil {
		return developers, err
	}

	db = db.Limit(1000).Find(&developers)
	if db.Error != nil {
		return developers, db.Error
	}

	return developers, nil
}

func SaveOrUpdateDeveloper(name string, vals Developer) (err error) {

	db, err := GetDB()
	if err != nil {
		return err
	}

	developer := new(Developer)
	developer.DeletedAt = nil

	db.Assign(vals).FirstOrCreate(developer, Developer{Name: name})
	if db.Error != nil {
		return db.Error
	}

	return nil
}

func DeleteDeveloper(id int) (err error) {

	db, err := GetDB()
	if err != nil {
		return err
	}

	developer := new(Developer)
	developer.ID = id

	db.Delete(developer)
	if db.Error != nil {
		return db.Error
	}

	return nil
}
