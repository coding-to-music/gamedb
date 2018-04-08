package mysql

import (
	"strconv"
	"time"

	"github.com/steam-authority/steam-authority/helpers"
)

type Publisher struct {
	ID           int        `gorm:"not null;column:id;primary_key;AUTO_INCREMENT"`
	CreatedAt    *time.Time `gorm:"not null;column:created_at"`
	UpdatedAt    *time.Time `gorm:"not null;column:updated_at"`
	DeletedAt    *time.Time `gorm:"not null;column:deleted_at"`
	Name         string     `gorm:"not null;column:name"`
	Apps         int        `gorm:"not null;column:apps"`
	MeanPrice    float64    `gorm:"not null;column:mean_price"`
	MeanDiscount float64    `gorm:"not null;column:mean_discount"`
}

func (p Publisher) GetPath() string {
	return "/games?publisher=" + strconv.Itoa(p.ID)
}

func (p Publisher) GetMeanPrice() string {
	return helpers.CentsFloat(p.MeanPrice)
}

func (p Publisher) GetMeanDiscount() string {
	return helpers.DollarsFloat(p.MeanDiscount)
}

func GetAllPublishers() (publishers []Publisher, err error) {

	db, err := GetDB()
	if err != nil {
		return publishers, err
	}

	db = db.Limit(1000).Order("name ASC").Find(&publishers)
	if db.Error != nil {
		return publishers, db.Error
	}

	return publishers, nil
}

func SaveOrUpdatePublisher(name string, vals Publisher) (err error) {

	db, err := GetDB()
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

	db, err := GetDB()
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
