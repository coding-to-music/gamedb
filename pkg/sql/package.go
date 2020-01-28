package sql

import (
	"errors"
	"html/template"
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/Jleagle/unmarshal-go/ctypes"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql/pics"
	"github.com/jinzhu/gorm"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	ErrInvalidPackageID = errors.New("invalid id")
)

type Package struct {
	AppIDs           string    `gorm:"not null;column:apps"`                             // []string
	AppItems         string    `gorm:"not null;column:app_items"`                        // map[string]string
	AppsCount        int       `gorm:"not null;column:apps_count"`                       //
	BundleIDs        string    `gorm:"not null;column:bundle_ids"`                       // []int
	BillingType      int8      `gorm:"not null;column:billing_type"`                     //
	ChangeNumber     int       `gorm:"not null;column:change_id"`                        //
	ChangeNumberDate time.Time `gorm:"not null;column:change_number_date;type:datetime"` //
	ComingSoon       bool      `gorm:"not null;column:coming_soon"`                      //
	Controller       string    `gorm:"not null;column:controller"`                       // JSON (TEXT)
	CreatedAt        time.Time `gorm:"not null;column:created_at;type:datetime"`         //
	DepotIDs         string    `gorm:"not null;column:depot_ids"`                        // []string
	Extended         string    `gorm:"not null;column:extended"`                         // PICSExtended
	Icon             string    `gorm:"not null;column:icon"`                             //
	ID               int       `gorm:"not null;column:id;PRIMARY_KEY"`                   //
	ImageLogo        string    `gorm:"not null;column:image_logo"`                       //
	ImagePage        string    `gorm:"not null;column:image_page"`                       //
	InStore          bool      `gorm:"not null;column:in_store"`                         // todo
	LicenseType      int8      `gorm:"not null;column:license_type"`                     //
	Name             string    `gorm:"not null;column:name"`                             //
	Platforms        string    `gorm:"not null;column:platforms"`                        // []string
	Prices           string    `gorm:"not null;column:prices"`                           // ProductPrices
	PurchaseText     string    `gorm:"not null;column:purchase_text"`                    //
	ReleaseDate      string    `gorm:"not null;column:release_date"`                     //
	ReleaseDateUnix  int64     `gorm:"not null;column:release_date_unix"`                //
	Status           int8      `gorm:"not null;column:status"`                           //
	UpdatedAt        time.Time `gorm:"not null;column:updated_at;type:datetime"`         //
}

func (pack *Package) BeforeCreate(scope *gorm.Scope) error {
	return pack.UpdateJSON(scope)
}

func (pack *Package) BeforeSave(scope *gorm.Scope) error {
	return pack.UpdateJSON(scope)
}

func (pack *Package) UpdateJSON(scope *gorm.Scope) error {

	if pack.AppIDs == "" {
		pack.AppIDs = "[]"
	}
	if pack.AppItems == "" {
		pack.AppItems = "{}"
	}
	if pack.BundleIDs == "" || pack.BundleIDs == "null" {
		pack.BundleIDs = "[]"
	}
	if pack.ChangeNumberDate.IsZero() {
		pack.ChangeNumberDate = time.Now()
	}
	if pack.Controller == "" {
		pack.Controller = "{}"
	}
	if pack.DepotIDs == "" {
		pack.DepotIDs = "[]"
	}
	if pack.Extended == "" {
		pack.Extended = "{}"
	}
	if pack.Platforms == "" {
		pack.Platforms = "[]"
	}
	if pack.Prices == "" {
		pack.Prices = "{}"
	}

	mon := mongo.Package{
		AppIDs:           pack.GetAppIDs(),
		AppItems:         pack.GetAppItems(),
		AppsCount:        pack.AppsCount,
		BundleIDs:        pack.GetBundleIDs(),
		BillingType:      int(pack.BillingType),
		ChangeNumber:     0,
		ChangeNumberDate: time.Time{},
		ComingSoon:       false,
		Controller:       nil,
		CreatedAt:        time.Time{},
		DepotIDs:         nil,
		Extended:         nil,
		Icon:             "",
		ID:               0,
		ImageLogo:        "",
		ImagePage:        "",
		InStore:          false,
		LicenseType:      0,
		Name:             "",
		Platforms:        nil,
		Prices:           nil,
		PurchaseText:     "",
		ReleaseDate:      "",
		ReleaseDateUnix:  0,
		Status:           0,
		UpdatedAt:        time.Time{},
	}

	_, err := mongo.UpdateOne(mongo.CollectionPackages, bson.D{{"_id", pack.ID}}, mon.BSON())
	return err
}

func (pack Package) GetPath() string {
	return helpers.GetPackagePath(pack.ID, pack.GetName())
}

func (pack Package) StoreLink() string {
	if !pack.InStore {
		return ""
	}
	return "https://store.steampowered.com/sub/" + strconv.Itoa(pack.ID) + "/?curator_clanid=&utm_source=GameDB" // todo curator_clanid
}

func (pack Package) GetID() int {
	return pack.ID
}

// For an interface
func (pack Package) GetType() string {
	return "Package"
}

func (pack Package) GetIcon() string {
	if pack.Icon == "" {
		return helpers.DefaultAppIcon
	}
	return pack.Icon
}

func (pack Package) GetProductType() helpers.ProductType {
	return helpers.ProductTypePackage
}

func (pack Package) GetName() (name string) {

	if (pack.Name == "") || (pack.Name == strconv.Itoa(pack.ID)) {
		return "Package " + strconv.Itoa(pack.ID)
	}

	return pack.Name
}

func (pack Package) GetCreatedNice() string {
	return pack.CreatedAt.Format(helpers.DateYearTime)
}

func (pack Package) GetCreatedUnix() int64 {
	return pack.CreatedAt.Unix()
}

func (pack Package) GetUpdatedNice() string {
	return pack.UpdatedAt.Format(helpers.DateYearTime)
}

func (pack Package) GetPICSUpdatedNice() string {

	d := pack.ChangeNumberDate

	// Empty dates
	if d.IsZero() || d.Unix() == -62167219200 {
		return "-"
	}
	return d.Format(helpers.DateYearTime)
}

func (pack Package) GetUpdatedUnix() int64 {
	return pack.UpdatedAt.Unix()
}

func (pack Package) GetReleaseDateNice() string {

	if pack.ReleaseDateUnix == 0 {
		return pack.ReleaseDate
	}

	return time.Unix(pack.ReleaseDateUnix, 0).Format(helpers.DateYear)
}

func (pack Package) GetBillingType() string {

	switch pack.BillingType {
	case 0:
		return "No Cost"
	case 1:
		return "Store"
	case 2:
		return "Bill Monthly"
	case 3:
		return "CD Key"
	case 4:
		return "Guest Pass"
	case 5:
		return "Hardware Promo"
	case 6:
		return "Gift"
	case 7:
		return "Free Weekend"
	case 8:
		return "OEM Ticket"
	case 9:
		return "Recurring Option"
	case 10:
		return "Store or CD Key"
	case 11:
		return "Repurchaseable"
	case 12:
		return "Free on Demand"
	case 13:
		return "Rental"
	case 14:
		return "Commercial License"
	case 15:
		return "Free Commercial License"
	default:
		return "Unknown"
	}
}

func (pack Package) GetLicenseType() string {

	switch pack.LicenseType {
	case 0:
		return "No License"
	case 1:
		return "Single Purchase"
	case 2:
		return "Single Purchase (Limited Use)"
	case 3:
		return "Recurring Charge"
	case 6:
		return "Recurring"
	case 7:
		return "Limited Use Delayed Activation"
	default:
		return "Unknown"
	}
}

func (pack Package) GetStatus() string {

	switch pack.Status {
	case 0:
		return "Available"
	case 2:
		return "Unavailable"
	default:
		return "Unknown"
	}
}

func (pack Package) GetComingSoon() string {

	switch pack.ComingSoon {
	case true:
		return "Yes"
	case false:
		return "No"
	default:
		return "Unknown"
	}
}

func (pack Package) GetAppsCountString() string {

	if pack.AppsCount == 0 {
		return "Unknown"
	}
	return strconv.Itoa(pack.AppsCount)
}

func (pack Package) GetAppIDs() (apps []int) {

	err := helpers.Unmarshal([]byte(pack.AppIDs), &apps)
	log.Err(err)

	return apps
}

func (pack Package) GetBundleIDs() (apps []int) {

	err := helpers.Unmarshal([]byte(pack.BundleIDs), &apps)
	log.Err(err)

	return apps
}

func (pack Package) GetAppItems() (apps map[int]int) {

	cApps := map[ctypes.Int]ctypes.Int{}

	err := helpers.Unmarshal([]byte(pack.AppItems), &cApps)
	log.Err(err)

	apps = map[int]int{}
	for k, v := range cApps {
		apps[int(k)] = int(v)
	}

	return apps
}

func (pack Package) GetDepotIDs() (depots []int) {

	err := helpers.Unmarshal([]byte(pack.DepotIDs), &depots)
	log.Err(err)

	return depots
}

func (pack Package) GetPrices() (prices helpers.ProductPrices) {

	err := helpers.Unmarshal([]byte(pack.Prices), &prices)
	log.Err(err)

	return prices
}

func (pack Package) GetPrice(code steam.ProductCC) (price helpers.ProductPrice) {

	return pack.GetPrices().Get(code)
}

func (pack Package) GetExtended() (extended pics.PICSKeyValues) {

	extended = pics.PICSKeyValues{}

	err := helpers.Unmarshal([]byte(pack.Extended), &extended)
	log.Err(err)

	return extended
}

func (pack Package) GetController() (controller pics.PICSController) {

	controller = pics.PICSController{}

	err := helpers.Unmarshal([]byte(pack.Controller), &controller)
	log.Err(err)

	return controller
}

func (pack Package) GetPlatforms() (platforms []string) {

	err := helpers.Unmarshal([]byte(pack.Platforms), &platforms)
	log.Err(err)

	return platforms
}

func (pack Package) GetPlatformImages() (ret template.HTML) {

	for _, v := range pack.GetPlatforms() {
		if v == "macos" {
			ret = ret + `<i class="fab fa-apple"></i>`
		} else if v == "windows" {
			ret = ret + `<i class="fab fa-windows"></i>`
		} else if v == "linux" {
			ret = ret + `<i class="fab fa-linux"></i>`
		}
	}

	return ret
}

func (pack Package) GetMetaImage() string {
	return pack.ImageLogo
}

func (pack Package) OutputForJSON(code steam.ProductCC) (output []interface{}) {

	return []interface{}{
		pack.ID,                        // 0
		pack.GetPath(),                 // 1
		pack.GetName(),                 // 2
		pack.GetComingSoon(),           // 3
		pack.AppsCount,                 // 4
		pack.GetPrice(code).GetFinal(), // 5
		pack.ChangeNumberDate.Unix(),   // 6
		pack.ChangeNumberDate.Format(helpers.DateYearTime), // 7
		pack.GetIcon(),                           // 8
		pack.GetPrice(code).GetDiscountPercent(), // 9
		pack.StoreLink(),                         // 10
	}
}

func (pack Package) GetDaysToRelease() string {

	return helpers.GetDaysToRelease(pack.ReleaseDateUnix)
}

func (pack Package) GetBundles() (bundles []Bundle, err error) {

	var item = memcache.MemcachePackageBundles(pack.ID)

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &bundles, func() (interface{}, error) {

		db, err := GetMySQLClient()
		if err != nil {
			return bundles, err
		}

		var bundles []Bundle

		db = db.Where("JSON_CONTAINS(package_ids, '[" + strconv.Itoa(pack.ID) + "]')")
		db = db.Find(&bundles)

		return bundles, db.Error
	})

	if len(bundles) == 0 {
		bundles = []Bundle{} // Needed for marshalling into slice type
	}

	return bundles, err
}

func (pack *Package) SetName(name string, force bool) {
	if (pack.Name == "" || force) && name != "" {
		pack.Name = name
	}
}

func IsValidPackageID(id int) bool {
	return id != 0
}

func GetPackage(id int) (pack Package, err error) {

	var item = memcache.MemcachePackage(id)

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &pack, func() (interface{}, error) {

		var pack Package

		db, err := GetMySQLClient()
		if err != nil {
			return pack, err
		}

		db = db.First(&pack, id)
		if db.Error != nil {
			return pack, db.Error
		}

		if pack.ID == 0 {
			return pack, ErrRecordNotFound
		}

		return pack, nil
	})

	return pack, err
}

func GetPackages(ids []int, columns []string) (packages []Package, err error) {

	if len(ids) == 0 {
		return packages, err
	}

	db, err := GetMySQLClient()
	if err != nil {
		return packages, err
	}

	if len(columns) > 0 {
		db = db.Select(columns)
	}

	db.Where("id IN (?)", ids).Find(&packages)

	return packages, db.Error
}

func CountPackages() (count int, err error) {

	var item = memcache.MemcachePackagesCount

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		var count int

		db, err := GetMySQLClient()
		if err != nil {
			return count, err
		}

		db.Model(&Package{}).Count(&count)
		return count, db.Error
	})

	return count, err
}
