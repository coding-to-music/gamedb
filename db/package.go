package db

import (
	"html/template"
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/helpers"
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
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
	ImageHeader      string    `gorm:"not null;column:image_header"`                     //
	ImageLogo        string    `gorm:"not null;column:image_logo"`                       //
	ImagePage        string    `gorm:"not null;column:image_page"`                       //
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

	return nil
}

func (pack Package) GetPath() string {
	return GetPackagePath(pack.ID, pack.GetName())
}

func (pack Package) GetID() int {
	return pack.ID
}

func (pack Package) GetIcon() string {
	if pack.Icon == "" {
		return DefaultAppIcon
	}
	return pack.Icon
}

func (pack Package) GetProductType() ProductType {
	return ProductTypePackage
}

func (pack Package) GetName() (name string) {

	var IDString = strconv.FormatInt(int64(pack.ID), 10)

	if (pack.Name == "") || (pack.Name == IDString) {
		return "Package " + IDString
	}

	return pack.Name
}

func (pack Package) HasDefaultName() bool {
	return pack.Name == strconv.FormatInt(int64(pack.ID), 10)
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

func (pack Package) GetAppIDs() (apps []int, err error) {

	err = helpers.Unmarshal([]byte(pack.AppIDs), &apps)
	return apps, err
}

func (pack Package) GetDepotIDs() (depots []int, err error) {

	err = helpers.Unmarshal([]byte(pack.DepotIDs), &depots)
	return depots, err
}

func (pack Package) GetPrices() (prices ProductPrices, err error) {

	err = helpers.Unmarshal([]byte(pack.Prices), &prices)
	return prices, err
}

func (pack Package) GetPrice(code steam.CountryCode) (price ProductPriceStruct, err error) {

	prices, err := pack.GetPrices()
	if err != nil {
		return price, err
	}

	return prices.Get(code)
}

func (pack Package) GetExtended() (extended PICSExtended, err error) {

	extended = PICSExtended{}

	err = helpers.Unmarshal([]byte(pack.Extended), &extended)
	return extended, err
}

func (pack Package) GetController() (controller PICSController, err error) {

	controller = PICSController{}

	err = helpers.Unmarshal([]byte(pack.Controller), &controller)
	return controller, err
}

func (pack Package) GetPlatforms() (platforms []string, err error) {

	err = helpers.Unmarshal([]byte(pack.Platforms), &platforms)
	return platforms, err
}

func (pack Package) GetPlatformImages() (ret template.HTML, err error) {

	platforms, err := pack.GetPlatforms()
	if err != nil {
		return ret, err
	}

	for _, v := range platforms {
		if v == "macos" {
			ret = ret + `<i class="fab fa-apple"></i>`
		} else if v == "windows" {
			ret = ret + `<i class="fab fa-windows"></i>`
		} else if v == "linux" {
			ret = ret + `<i class="fab fa-linux"></i>`
		}
	}

	return ret, nil
}

func (pack Package) GetMetaImage() string {

	if pack.ImageHeader != "" {
		return pack.ImageHeader
	}

	if pack.ImageLogo != "" {
		return pack.ImageLogo
	}

	return ""
}

func (pack Package) OutputForJSON(code steam.CountryCode) (output []interface{}) {

	return []interface{}{
		pack.ID,
		pack.GetPath(),
		pack.GetName(),
		pack.GetComingSoon(),
		pack.AppsCount,
		GetPriceFormatted(pack, code).Final,
		pack.ChangeNumberDate.Unix(),
		pack.ChangeNumberDate.Format(helpers.DateYearTime),
		pack.GetIcon(),
	}
}

// Must be the same as app OutputForJSONUpcoming
func (pack Package) OutputForJSONUpcoming(code steam.CountryCode) (output []interface{}) {

	return []interface{}{
		pack.ID,
		pack.GetName(),
		pack.GetIcon(),
		pack.GetPath(),
		pack.AppsCount,
		GetPriceFormatted(pack, code).Final,
		pack.GetDaysToRelease(),
		pack.GetReleaseDateNice(),
	}
}

func (pack Package) GetDaysToRelease() string {

	return helpers.GetDaysToRelease(pack.ReleaseDateUnix)
}

func IsValidPackageID(id int) bool {
	return id != 0
}

func GetPackagePath(id int, name string) string {

	path := "/packages/" + strconv.Itoa(id)
	if name == "" {
		return path
	}
	return path + "/" + slug.Make(name)
}

func GetPackage(id int, columns []string) (pack Package, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return pack, err
	}

	db = db.First(&pack, id)
	if db.Error != nil {
		return pack, db.Error
	}

	if len(columns) > 0 {
		db = db.Select(columns)
		if db.Error != nil {
			return pack, db.Error
		}
	}

	if pack.ID == 0 {
		return pack, ErrRecordNotFound
	}

	return pack, nil
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

func GetPackagesAppIsIn(appID int) (packages []Package, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return packages, err
	}

	db = db.Where("JSON_CONTAINS(apps, '[" + strconv.Itoa(appID) + "]')").Order("id DESC").Find(&packages)

	if db.Error != nil {
		return packages, db.Error
	}

	return packages, nil
}

func CountPackages() (count int, err error) {

	var item = helpers.MemcachePackagesCount

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

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
