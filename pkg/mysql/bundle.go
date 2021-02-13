package mysql

import (
	"math"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/jinzhu/gorm"
)

type Bundle struct {
	AppIDs          string    `gorm:"not null;column:app_ids"` // JSON
	CreatedAt       time.Time `gorm:"not null;column:created_at;type:datetime"`
	Discount        int       `gorm:"not null;column:discount"`
	DiscountHighest int       `gorm:"not null;column:highest_discount"`
	DiscountLowest  int       `gorm:"not null;column:lowest_discount"`
	DiscountSale    int       `gorm:"not null;column:discount_sale"`
	Giftable        bool      `gorm:"not null;column:giftable"`
	Icon            string    `gorm:"not null;column:icon"`
	ID              int       `gorm:"not null;column:id"`
	Image           string    `gorm:"not null;column:image"`
	Name            string    `gorm:"not null;column:name"`
	OnSale          bool      `gorm:"not null;column:on_sale"`
	PackageIDs      string    `gorm:"not null;column:package_ids"` // JSON
	Prices          string    `gorm:"not null;column:prices"`      // JSON
	PricesSale      string    `gorm:"not null;column:sale_prices"` // JSON
	Type            string    `gorm:"not null;column:type"`
	UpdatedAt       time.Time `gorm:"not null;column:updated_at;type:datetime"`
}

func (bundle Bundle) GetID() int {
	return bundle.ID
}

func (bundle Bundle) GetUpdated() time.Time {
	return bundle.UpdatedAt
}

func (bundle Bundle) GetDiscount() int {
	return bundle.Discount
}

func (bundle Bundle) GetDiscountHighest() int {
	return bundle.DiscountHighest
}

func (bundle Bundle) GetScore() float64 {
	return 0
}

func (bundle Bundle) GetApps() int {
	return bundle.AppsCount()
}

func (bundle Bundle) GetPackages() int {
	return bundle.PackagesCount()
}

func (bundle Bundle) IsGiftable() bool {
	return bundle.Giftable
}

func (bundle *Bundle) BeforeSave(scope *gorm.Scope) error {

	bundle.Discount = int(math.Abs(float64(bundle.Discount)))
	bundle.DiscountHighest = int(math.Abs(float64(bundle.DiscountHighest)))
	bundle.DiscountLowest = int(math.Abs(float64(bundle.DiscountLowest)))
	bundle.DiscountSale = int(math.Abs(float64(bundle.DiscountSale)))

	if bundle.AppIDs == "" {
		bundle.AppIDs = "[]"
	}
	if bundle.PackageIDs == "" {
		bundle.PackageIDs = "[]"
	}

	bundle.OnSale = bundle.DiscountSale > bundle.Discount

	if bundle.DiscountSale < bundle.Discount {
		bundle.DiscountSale = bundle.Discount
	}

	if bundle.DiscountSale > bundle.DiscountHighest {
		bundle.DiscountHighest = bundle.DiscountSale
	}

	if bundle.Discount < bundle.DiscountLowest {
		bundle.DiscountLowest = bundle.Discount
	}

	return nil
}

func (bundle Bundle) GetPath() string {
	return helpers.GetBundlePath(bundle.ID, bundle.Name)
}

func (bundle Bundle) GetName() string {
	return helpers.GetBundleName(bundle.ID, bundle.Name)
}

func (bundle Bundle) GetStoreLink() string {
	return helpers.GetBundleStoreLink(bundle.ID)
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

func (bundle Bundle) GetPrices() (ret map[steamapi.ProductCC]int) {

	ret = map[steamapi.ProductCC]int{}

	if bundle.Prices == "" {
		return ret
	}

	err := helpers.Unmarshal([]byte(bundle.Prices), &ret)
	if err != nil {
		log.ErrS(err)
		return ret
	}

	return ret
}

func (bundle Bundle) GetPricesSale() (ret map[steamapi.ProductCC]int) {

	ret = map[steamapi.ProductCC]int{}

	if bundle.Prices == "" {
		return ret
	}

	err := helpers.Unmarshal([]byte(bundle.PricesSale), &ret)
	if err != nil {
		log.ErrS(err)
		return ret
	}

	return ret
}

func (bundle Bundle) GetPricesFormatted() (ret map[steamapi.ProductCC]string) {

	ret = map[steamapi.ProductCC]string{}

	for k, v := range bundle.GetPrices() {
		ret[k] = i18n.FormatPrice(i18n.GetProdCC(k).CurrencyCode, v)
	}

	return ret
}

func (bundle Bundle) AppsCount() int {

	apps, err := bundle.GetAppIDs()
	if err != nil {
		log.ErrS(err)
	}
	return len(apps)
}

func (bundle Bundle) PackagesCount() int {

	return len(bundle.GetPackageIDs())
}

func (bundle Bundle) GetPackageIDs() (ids []int) {

	err := helpers.Unmarshal([]byte(bundle.PackageIDs), &ids)
	if err != nil {
		log.ErrS(err)
	}

	return ids
}

func (bundle Bundle) OutputForJSON() (output []interface{}) {
	return helpers.OutputBundleForJSON(bundle)
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
