package db

import (
	"strconv"
	"time"
)

type Category struct {
	ID        int        `gorm:"not null;primary_key"`
	CreatedAt time.Time  `gorm:"not null"`
	UpdatedAt time.Time  `gorm:"not null"`
	DeletedAt *time.Time `gorm:""`
	Name      string     `gorm:"not null;index:name"`
}

func (category Category) GetPath() string {
	return "/apps?categories=" + strconv.Itoa(category.ID)
}

func (category Category) GetName() (name string) {

	if category.Name == "" {
		return "Tag " + strconv.Itoa(category.ID)
	}

	return category.Name
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

// func GetTagsForSelect() (tags []Tag, err error) {
//
// 	var item = helpers.MemcacheTagKeyNames
//
// 	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &tags, func() (interface{}, error) {
//
// 		var tags []Tag
//
// 		db, err := GetMySQLClient()
// 		if err != nil {
// 			return tags, err
// 		}
//
// 		db = db.Select([]string{"id", "name"}).Order("name ASC").Find(&tags)
// 		return tags, db.Error
// 	})
//
// 	return tags, err
// }
//
// func GetTagsByID(ids []int, columns []string) (tags []Tag, err error) {
//
// 	if len(ids) == 0 {
// 		return tags, err
// 	}
//
// 	db, err := GetMySQLClient()
// 	if err != nil {
// 		return tags, err
// 	}
//
// 	if len(columns) > 0 {
// 		db = db.Select(columns)
// 	}
//
// 	db = db.Where("id IN (?)", ids)
// 	db = db.Order("name ASC")
// 	db = db.Limit(100)
// 	db = db.Find(&tags)
//
// 	return tags, db.Error
// }
//
// func DeleteTags(ids []int) (err error) {
//
// 	log.Info("Deleteing " + strconv.Itoa(len(ids)) + " tags")
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
