package mysql

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/jinzhu/gorm"
)

type Publisher struct {
	ID        int        `gorm:"not null;primary_key;AUTO_INCREMENT"`
	CreatedAt time.Time  `gorm:"not null"`
	UpdatedAt time.Time  `gorm:"not null"`
	DeletedAt *time.Time `gorm:""`
	Name      string     `gorm:"not null;index:name"`
	Apps      int        `gorm:"not null"`
	MeanPrice string     `gorm:"not null"` // map[steamapi.CountryCode]float64
	MeanScore float64    `gorm:"not null"`
}

func (p *Publisher) BeforeCreate(scope *gorm.Scope) error {
	return p.Before(scope)
}

func (p *Publisher) BeforeSave(scope *gorm.Scope) error {
	return p.Before(scope)
}

func (p *Publisher) Before(scope *gorm.Scope) error {

	p.Name = TrimPublisherName(p.Name)

	return nil
}

func (p Publisher) GetPath() string {
	return "/games?publishers=" + strconv.Itoa(p.ID)
}

func (p Publisher) GetName() (name string) {
	if p.Name == "" {
		return "Publisher " + humanize.Comma(int64(p.ID))
	}

	return p.Name
}

func (p Publisher) GetMeanPrice(code steamapi.ProductCC) (string, error) {
	return GetMeanPrice(code, p.MeanPrice)
}

func (p Publisher) GetMeanScore() string {
	return helpers.FloatToString(p.MeanScore, 2) + "%"
}

func TrimPublisherName(name string) string {

	if len(name) > 255 {
		name = name[:255]
	}

	return strings.TrimSpace(name)
}

func GetPublisher(id int) (publisher Publisher, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return publisher, err
	}

	db = db.Where("id = ?", id)
	db = db.Limit(1)
	db = db.Find(&publisher)

	db = db.First(&publisher, id)

	return publisher, db.Error
}

func GetPublishersByID(ids []int, columns []string) (publishers []Publisher, err error) {

	if len(ids) == 0 {
		return publishers, err
	}

	db, err := GetMySQLClient()
	if err != nil {
		return publishers, err
	}

	if len(columns) > 0 {
		db = db.Select(columns)
	}

	db = db.Where("id IN (?)", ids)
	db = db.Order("name ASC")
	db = db.Limit(100)
	db = db.Find(&publishers)

	return publishers, db.Error
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

	var item = memcache.MemcachePublisherKeyNames

	err = memcache.GetSetInterface(item.Key, item.Expiration, &pubs, func() (interface{}, error) {

		var pubs []Publisher

		db, err := GetMySQLClient()
		if err != nil {
			return pubs, err
		}

		db = db.Select([]string{"id", "name"}).Order("apps DESC").Limit(200).Find(&pubs)
		if db.Error != nil {
			return pubs, db.Error
		}

		sort.Slice(pubs, func(i, j int) bool {
			return pubs[i].Name < pubs[j].Name
		})

		return pubs, err
	})

	return pubs, err
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
