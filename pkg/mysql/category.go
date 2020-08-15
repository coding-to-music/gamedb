package mysql

import (
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/memcache"
)

type Category struct {
	ID        int        `gorm:"not null;primary_key"`
	CreatedAt time.Time  `gorm:"not null"`
	UpdatedAt time.Time  `gorm:"not null"`
	DeletedAt *time.Time `gorm:""`
	Name      string     `gorm:"not null;index:name"`
	Apps      int        `gorm:"not null"`
	MeanPrice string     `gorm:"not null"` // map[steamapi.CountryCode]float64
	MeanScore float64    `gorm:"not null"`
}

func (category Category) GetPath() string {
	return "/games?categories=" + strconv.Itoa(category.ID)
}

func (category Category) GetName() (name string) {

	if category.Name == "" {
		return "Category " + humanize.Comma(int64(category.ID))
	}

	return category.Name
}

func (category Category) GetMeanPrice(code steamapi.ProductCC) (string, error) {
	return GetMeanPrice(code, category.MeanPrice)
}

func (category Category) GetMeanScore() string {
	return helpers.FloatToString(category.MeanScore, 2) + "%"
}

func GetAllCategories() (categories []Category, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return categories, err
	}

	db = db.Find(&categories)
	if db.Error != nil {
		return categories, db.Error
	}

	return categories, nil
}

func GetCategoriesByID(ids []int, columns []string) (categories []Category, err error) {

	if len(ids) == 0 {
		return categories, err
	}

	db, err := GetMySQLClient()
	if err != nil {
		return categories, err
	}

	if len(columns) > 0 {
		db = db.Select(columns)
	}

	db = db.Where("id IN (?)", ids)
	db = db.Order("name ASC")
	db = db.Limit(100)
	db = db.Find(&categories)

	return categories, db.Error
}

func GetCategoriesForSelect() (tags []Category, err error) {

	var item = memcache.MemcacheCategoryKeyNames

	err = memcache.GetSetInterface(item.Key, item.Expiration, &tags, func() (interface{}, error) {

		var cats []Category

		db, err := GetMySQLClient()
		if err != nil {
			return cats, err
		}

		db = db.Select([]string{"id", "name"}).Order("name ASC").Find(&cats)
		return cats, db.Error
	})

	return tags, err
}

// func DeleteTags(ids []int) (err error) {
//
// 	zap.S().Info("Deleteing " + strconv.Itoa(len(ids)) + " tags")
//
// 	if len(ids) == 0 {
// 		return nil
// 	}
//
// 	db, err := GetMySQLClient()
// 	if err != nil {
// 		return err
// 	}
//
// 	db.Where("id IN (?)", ids).Delete(Tag{})
//
// 	return db.Error
// }
