package db

import (
	"strconv"
	"time"

	"github.com/steam-authority/steam-authority/helpers"
)

type Publisher struct {
	ID           int        `gorm:"not null;primary_key;AUTO_INCREMENT"`
	CreatedAt    *time.Time `gorm:"not null"`
	UpdatedAt    *time.Time `gorm:"not null"`
	DeletedAt    *time.Time `gorm:""`
	Name         string     `gorm:"not null;index:name"`
	Apps         int        `gorm:"not null"`
	MeanPrice    float64    `gorm:"not null"`
	MeanDiscount float64    `gorm:"not null"`
}

func (p Publisher) GetPath() string {
	return "/games?publisher=" + strconv.Itoa(p.ID)
}

func (p Publisher) GetMeanPrice() float64 {
	return helpers.CentsFloat(p.MeanPrice)
}

func (p Publisher) GetMeanDiscount() float64 {
	return helpers.DollarsFloat(p.MeanDiscount)
}

func GetAllPublishers() (publishers []Publisher, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return publishers, err
	}

	db = db.Limit(1000).Find(&publishers)
	if db.Error != nil {
		return publishers, db.Error
	}

	return publishers, nil
}

func SaveOrUpdatePublisher(name string, vals Publisher) (err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	publisher := new(Publisher)
	publisher.DeletedAt = nil

	db.Assign(vals).FirstOrCreate(publisher, Publisher{Name: name})
	if db.Error != nil {
		return db.Error
	}

	return nil
}

func DeletePublisher(id int) (err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	publisher := new(Publisher)
	publisher.ID = id

	db.Delete(publisher)
	if db.Error != nil {
		return db.Error
	}

	return nil
}
