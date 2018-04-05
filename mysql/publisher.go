package mysql

import (
	"strconv"
	"time"
)

type Publisher struct {
	ID           int        `gorm:"not null;column:id;primary_key;AUTO_INCREMENT"`
	CreatedAt    *time.Time `gorm:"not null;column:created_at"`
	UpdatedAt    *time.Time `gorm:"not null;column:updated_at"`
	Name         string     `gorm:"not null;column:name"`
	Apps         int        `gorm:"not null;column:apps"`
	MeanPrice    float64    `gorm:"not null;column:mean_price"`
	MeanDiscount float64    `gorm:"not null;column:mean_discount"`
}

func (p Publisher) GetPath() string {
	return "/apps?publisher=" + strconv.Itoa(p.ID)
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

func SaveOrUpdatePublisher(id int, vals Publisher) (err error) {

	db, err := GetDB()
	if err != nil {
		return err
	}

	publisher := new(Publisher)
	db.Assign(vals).FirstOrCreate(publisher, Publisher{ID: id})
	if db.Error != nil {
		return db.Error
	}

	return nil
}
