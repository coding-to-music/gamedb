package mysql

import (
	"strconv"
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
)

type Bundle struct {
	ID              int       `gorm:"not null;column:id"`
	CreatedAt       time.Time `gorm:"not null;column:created_at;type:datetime"`
	UpdatedAt       time.Time `gorm:"not null;column:updated_at;type:datetime"`
	Name            string    `gorm:"not null;column:name"`
	Discount        int       `gorm:"not null;column:discount"`
	HighestDiscount int       `gorm:"not null;column:highest_discount"`
	LowestDiscount  int       `gorm:"not null;column:lowest_discount"`
	AppIDs          string    `gorm:"not null;column:app_ids"`     // JSON
	PackageIDs      string    `gorm:"not null;column:package_ids"` // JSON
	Image           string    `gorm:"not null;column:image"`
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

func (bundle *Bundle) SetDiscount(discount int) {

	bundle.Discount = discount

	if discount < bundle.HighestDiscount {
		bundle.HighestDiscount = discount
	}

	if discount > bundle.LowestDiscount || bundle.LowestDiscount == 0 {
		bundle.LowestDiscount = discount
	}
}

func (bundle Bundle) GetPath() string {
	return "/bundles/" + strconv.Itoa(bundle.ID) + "/" + slug.Make(bundle.GetName())
}

func (bundle Bundle) GetName() string {

	var name = strings.TrimSpace(bundle.Name)

	if name != "" {
		return name
	}
	return "Bundle " + strconv.Itoa(bundle.ID)
}

func (bundle Bundle) GetStoreLink() string {
	name := config.C.GameDBShortName
	return "https://store.steampowered.com/bundle/" + strconv.Itoa(bundle.ID) + "?utm_source=" + name + "&utm_medium=referral&utm_campaign=app-store-link"
}

func (bundle Bundle) GetCreatedNice() string {
	return bundle.CreatedAt.Format(helpers.DateYearTime)
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
	if err != nil {
		log.ErrS(err)
	}
	return len(apps)
}

func (bundle Bundle) GetPackageIDs() (ids []int) {

	err := helpers.Unmarshal([]byte(bundle.PackageIDs), &ids)
	if err != nil {
		log.ErrS(err)
	}

	return ids
}

func (bundle Bundle) OutputForJSON() (output []interface{}) {

	return []interface{}{
		bundle.ID,        // 0
		bundle.GetName(), // 1
		bundle.GetPath(), // 2
		strconv.FormatInt(bundle.UpdatedAt.Unix(), 10), // 3
		bundle.Discount,             // 4
		bundle.AppsCount(),          // 5
		len(bundle.GetPackageIDs()), // 6
		bundle.HighestDiscount == bundle.Discount && bundle.Discount != 0, // 7 Highest ever discount
		bundle.GetStoreLink(), // 8
	}
}

func (bundle Bundle) Save() error {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	db = db.Save(&bundle)
	return db.Error
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

func GetBundlesByID(ids []int, columns []string) (bundles []Bundle, err error) {

	if len(ids) == 0 {
		return bundles, nil
	}

	db, err := GetMySQLClient()
	if err != nil {
		return bundles, err
	}

	ids = helpers.UniqueInt(ids)

	chunks := helpers.ChunkInts(ids, 100)
	for _, chunk := range chunks {

		db = db.New()

		if len(columns) > 0 {
			db = db.Select(columns)
		}

		var bundlesChunk []Bundle
		db = db.Where("id IN (?)", chunk).Find(&bundlesChunk)
		if db.Error != nil {
			log.ErrS(db.Error)
			return bundles, db.Error
		}

		bundles = append(bundles, bundlesChunk...)
	}

	return bundles, nil
}

func CountBundles() (count int, err error) {

	err = memcache.GetSetInterface(memcache.ItemBundlesCount, &count, func() (interface{}, error) {

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
