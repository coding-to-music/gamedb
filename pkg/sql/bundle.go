package sql

import (
	"strconv"
	"time"

	"github.com/gamedb/website/pkg/config"
	"github.com/gamedb/website/pkg/helpers"
	"github.com/gamedb/website/pkg/log"
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
)

type Bundle struct {
	ID         int
	CreatedAt  time.Time `gorm:"not null;column:created_at;type:datetime"`
	UpdatedAt  time.Time `gorm:"not null;column:updated_at;type:datetime"`
	Name       string
	Discount   int
	AppIDs     string
	PackageIDs string
	Image      string
}

func (bundle *Bundle) BeforeSave(scope *gorm.Scope) error {

	if bundle.AppIDs == "" {
		bundle.AppIDs = "[]"
	}
	if bundle.PackageIDs == "" {
		bundle.PackageIDs = "[]"
	}

	return nil
}

func (bundle Bundle) GetPath() string {
	return "/bundles/" + strconv.Itoa(bundle.ID) + "/" + slug.Make(bundle.Name)
}

func (bundle Bundle) GetStoreLink() string {
	name := config.Config.GameDBShortName.Get()
	return "https://store.steampowered.com/bundle/" + strconv.Itoa(bundle.ID) + "?utm_source=" + name + "&utm_medium=link&utm_campaign=" + name
}

func (bundle Bundle) GetUpdatedNice() string {
	return bundle.UpdatedAt.Format(helpers.DateYearTime)
}

func (bundle Bundle) GetAppIDs() (ids []int, err error) {

	err = helpers.Unmarshal([]byte(bundle.AppIDs), &ids)
	return ids, err
}

func (bundle Bundle) AppsCount() int {

	apps, err := bundle.GetAppIDs()
	log.Err(err)
	return len(apps)
}

func (bundle Bundle) GetPackageIDs() (ids []int, err error) {

	err = helpers.Unmarshal([]byte(bundle.PackageIDs), &ids)
	return ids, err
}

func (bundle Bundle) PackagesCount() int {

	packages, err := bundle.GetPackageIDs()
	log.Err(err)
	return len(packages)
}

func (bundle Bundle) OutputForJSON() (output []interface{}) {

	return []interface{}{
		bundle.ID,
		bundle.Name,
		bundle.GetPath(),
		strconv.FormatInt(bundle.UpdatedAt.Unix(), 10),
		bundle.Discount,
		bundle.AppsCount(),
		bundle.PackagesCount(),
	}
}

func GetBundle(id int, columns []string) (bundle Bundle, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return bundle, err
	}

	db = db.First(&bundle, id)
	if db.Error != nil {
		return bundle, db.Error
	}

	if len(columns) > 0 {
		db = db.Select(columns)
		if db.Error != nil {
			return bundle, db.Error
		}
	}

	if bundle.ID == 0 {
		return bundle, ErrRecordNotFound
	}

	return bundle, nil
}

func CountBundles() (count int, err error) {

	var item = helpers.MemcacheBundlesCount

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		var count int

		db, err := GetMySQLClient()
		if err != nil {
			return count, err
		}

		db.Model(&Bundle{}).Count(&count)

		return count, db.Error
	})

	return count, err
}
