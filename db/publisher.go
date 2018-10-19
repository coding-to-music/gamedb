package db

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/memcache"
)

type Publisher struct {
	ID        int        `gorm:"not null;primary_key;AUTO_INCREMENT"`
	CreatedAt *time.Time `gorm:"not null"`
	UpdatedAt *time.Time `gorm:"not null"`
	DeletedAt *time.Time `gorm:""`
	Name      string     `gorm:"not null;index:name"`
	Apps      int        `gorm:"not null"`
	MeanPrice float64    `gorm:"not null"`
	MeanScore float64    `gorm:"not null"`
}

func (p Publisher) GetPath() string {
	return "/games?publisher=" + strconv.Itoa(p.ID)
}

func (p Publisher) GetName() (name string) {
	return p.Name
}

func (p Publisher) GetMeanPrice() float64 {
	return helpers.CentsFloat(p.MeanPrice)
}

func (p Publisher) GetMeanScore() float64 {
	return helpers.DollarsFloat(p.MeanScore)
}

func GetAllPublishers() (publishers []Publisher, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return publishers, err
	}

	db = db.Find(&publishers)
	if db.Error != nil {
		return publishers, db.Error
	}

	return publishers, nil
}

func GetPublishersForSelect() (pubs []Publisher, err error) {

	s, err := memcache.GetSetString(memcache.PublisherKeyNames, func() (s string, err error) {

		db, err := GetMySQLClient()
		if err != nil {
			return s, err
		}

		var pubs []Publisher
		db = db.Select([]string{"id", "name"}).Order("name ASC").Find(&pubs)
		if db.Error != nil {
			return s, db.Error
		}

		bytes, err := json.Marshal(pubs)
		return string(bytes), err
	})

	if err != nil {
		return pubs, err
	}

	err = json.Unmarshal([]byte(s), &pubs)
	return pubs, err
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

func DeletePublishers(ids []int) (err error) {

	if len(ids) == 0 {
		return nil
	}

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	db.Where("id IN (?)", ids).Delete(Publisher{})

	return db.Error
}
