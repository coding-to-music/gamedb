package db

import (
	"encoding/json"
	"html/template"
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gosimple/slug"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logging"
	"github.com/steam-authority/steam-authority/memcache"
)

// todo, make column meta match table names
type Package struct {
	ID        int        `gorm:"not null;primary_key"` //
	CreatedAt *time.Time `gorm:"not null"`             //
	UpdatedAt *time.Time `gorm:"not null"`             //

	PICSName        string `gorm:"not null"`                          //
	PICSChangeID    int    `gorm:"not null"`                          //
	PICSBillingType int8   `gorm:"not null;column:billing_type"`      //
	PICSLicenseType int8   `gorm:"not null;column:license_type"`      //
	PICSStatus      int8   `gorm:"not null;column:status"`            //
	PICSExtended    string `gorm:"not null;default:'{}'"`             // JSON (TEXT)
	PICSAppIDs      string `gorm:"not null;default:'[]';column:apps"` // JSON
	PICSAppItems    string `gorm:"not null;default:'{}'"`             // JSON (TEXT)
	PICSDepotIDs    string `gorm:"not null;default:'[]'"`             // JSON
	PICSRaw         string `gorm:"not null;default:'{}'"`             // JSON (TEXT)

	AppsCount       int    `gorm:"not null"`              //
	ImagePage       string `gorm:"not null"`              //
	ImageHeader     string `gorm:"not null"`              //
	ImageLogo       string `gorm:"not null"`              //
	PurchaseText    string `gorm:"not null"`              //
	PriceInitial    int    `gorm:"not null"`              //
	PriceFinal      int    `gorm:"not null"`              //
	PriceDiscount   int    `gorm:"not null"`              //
	PriceIndividual int    `gorm:"not null"`              //
	Controller      string `gorm:"not null;default:'{}'"` // JSON (TEXT)
	ComingSoon      bool   `gorm:"not null"`              //
	ReleaseDate     string `gorm:"not null"`              //
	Platforms       string `gorm:"not null;default:'[]'"` // JSON
}

func GetDefaultPackageJSON() Package {
	return Package{
		//PICSAppIDs:   "[]",
		//PICSExtended: "{}",
		//Controller:   "{}",
		//Platforms:    "[]",
	}
}

func (pack Package) GetPath() string {

	s := "/packages/" + strconv.Itoa(pack.ID)

	if pack.PICSName != "" {
		s = s + "/" + slug.Make(pack.GetName())
	}

	return s
}

func (pack Package) GetName() (name string) {

	if pack.PICSName == "" {
		pack.PICSName = "Package " + strconv.FormatInt(int64(pack.ID), 10)
	}

	return pack.PICSName
}

func (pack Package) GetDefaultAvatar() string {
	return "/assets/img/no-app-image-square.jpg"
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

func (pack Package) GetUpdatedUnix() int64 {
	return pack.UpdatedAt.Unix()
}

func (pack Package) GetReleaseDateNice() string {

	return helpers.GetReleaseDateNice(pack.ReleaseDate)
}

func (pack Package) GetReleaseDateUnix() int64 {

	return helpers.GetReleaseDateUnix(pack.ReleaseDate)
}

func (pack Package) GetBillingType() string {

	switch pack.PICSBillingType {
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

	switch pack.PICSLicenseType {
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

	switch pack.PICSStatus {
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

	err = helpers.Unmarshal([]byte(pack.PICSAppIDs), &apps)
	return apps, err
}

func (pack *Package) SetAppIDs(apps []int) (err error) {

	bytes, err := json.Marshal(apps)
	if err != nil {

		pack.PICSAppIDs = string(bytes)
		pack.AppsCount = len(apps)
	}

	return err
}

func (pack *Package) SetDepotIDs(apps []int) (err error) {

	bytes, err := json.Marshal(apps)
	if err != nil {
		return err
	}

	pack.PICSDepotIDs = string(bytes)

	return nil
}

func (pack *Package) SetAppItems(items map[string]string) (err error) {

	bytes, err := json.Marshal(items)
	if err != nil {
		return err
	}

	pack.PICSAppItems = string(bytes)

	return nil
}

func (pack Package) GetPriceInitial() float64 {
	return helpers.CentsInt(pack.PriceInitial)
}

func (pack Package) GetPriceFinal() float64 {
	return helpers.CentsInt(pack.PriceFinal)
}

func (pack Package) GetPriceDiscount() float64 {
	return helpers.CentsInt(pack.PriceDiscount)
}

func (pack Package) GetPriceIndividual() float64 {
	return helpers.CentsInt(pack.PriceInitial)
}

type Extended map[string]string

func (pack *Package) SetExtended(extended Extended) (err error) {

	bytes, err := json.Marshal(extended)
	if err != nil {
		return err
	}

	pack.PICSExtended = string(bytes)

	return nil
}

func (pack Package) GetExtended() (extended map[string]interface{}, err error) {

	extended = make(map[string]interface{})

	err = helpers.Unmarshal([]byte(pack.PICSExtended), &extended)
	return extended, err
}

// Used in temmplate
func (pack Package) GetExtendedNice() (ret map[string]interface{}) {

	ret = make(map[string]interface{})

	extended, err := pack.GetExtended()
	if err != nil {
		logging.Error(err)
		return ret
	}

	for k, v := range extended {

		if val, ok := PackageExtendedKeys[k]; ok {
			ret[val] = v
		} else {
			logging.Info("Need to add " + k + " to extended map")
			ret[k] = v
		}
	}

	return ret
}

func (pack Package) GetController() (controller map[string]interface{}, err error) {

	controller = make(map[string]interface{})

	err = helpers.Unmarshal([]byte(pack.Controller), &controller)
	return controller, err
}

// Used in temmplate
func (pack Package) GetControllerNice() (ret map[string]interface{}) {

	ret = map[string]interface{}{}

	extended, err := pack.GetController()
	if err != nil {
		logging.Error(err)
		return ret
	}

	for k, v := range extended {

		if val, ok := PackageControllerKeys[k]; ok {
			ret[val] = v
		} else {
			logging.Info("Need to add " + k + " to controller map")
			ret[k] = v
		}
	}

	return ret
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

func GetPackage(id int) (pack Package, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return pack, err
	}

	db.First(&pack, id)
	if db.Error != nil {
		return pack, db.Error
	}

	if pack.ID == 0 {
		return pack, ErrNotFound
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

	return memcache.GetSetInt(memcache.PackagesCount, func() (count int, err error) {

		db, err := GetMySQLClient()
		if err != nil {
			return count, err
		}

		db.Model(&Package{}).Count(&count)
		return count, db.Error
	})
}

// GORM callback
func (pack *Package) Update() (errs []error) {

	var wg sync.WaitGroup

	// Get package details
	wg.Add(1)
	go func(pack *Package) {

		// Get app details
		// Get data
		response, _, err := helpers.GetSteam().GetPackageDetails(pack.ID)
		if err != nil {

			if err == steam.ErrNullResponse {
				errs = append(errs, err)
			}
		}

		// Controller
		controllerString, err := json.Marshal(response.Data.Controller)
		if err != nil {
			errs = append(errs, err)
		}

		// Platforms
		var platforms []string
		if response.Data.Platforms.Linux {
			platforms = append(platforms, "linux")
		}
		if response.Data.Platforms.Windows {
			platforms = append(platforms, "windows")
		}
		if response.Data.Platforms.Windows {
			platforms = append(platforms, "macos")
		}

		platformsString, err := json.Marshal(platforms)
		if err != nil {
			errs = append(errs, err)
		}

		//
		pack.ImageHeader = response.Data.HeaderImage
		pack.ImageLogo = response.Data.SmallLogo
		pack.ImageHeader = response.Data.HeaderImage
		// pack.PICSAppIDs = string(appsString) // Can get from PICS
		pack.PriceInitial = response.Data.Price.Initial
		pack.PriceFinal = response.Data.Price.Final
		pack.PriceDiscount = response.Data.Price.DiscountPercent
		pack.PriceIndividual = response.Data.Price.Individual
		pack.Platforms = string(platformsString)
		pack.Controller = string(controllerString)
		pack.ReleaseDate = response.Data.ReleaseDate.Date
		pack.ComingSoon = response.Data.ReleaseDate.ComingSoon

		wg.Done()
	}(pack)

	// Default JSON values
	if pack.PICSAppIDs == "" || pack.PICSAppIDs == "null" {
		pack.PICSAppIDs = "[]"
	}

	if pack.PICSExtended == "" || pack.PICSExtended == "null" {
		pack.PICSExtended = "{}"
	}

	if pack.Controller == "" || pack.Controller == "null" {
		pack.Controller = "{}"
	}

	if pack.Platforms == "" || pack.Platforms == "null" {
		pack.Platforms = "[]"
	}

	return errs
}

var PackageExtendedKeys = map[string]string{
	"allowcrossregiontradingandgifting":     "Allow Cross Region Trading & Gifting",
	"allowpurchasefromretrictedcountries":   "Allow Purchase From Restricted Countries",
	"allowpurchasefromrestrictedcountries":  "Allow Purchase From Restricted Countries",
	"allowpurchaseinrestrictedcountries":    "Allow Purchase In Restricted Countries",
	"allowpurchaserestrictedcountries":      "Allow Purchase Restricted Countries",
	"allowrunincountries":                   "Allow Run Inc Cuntries",
	"alwayscountasowned":                    "Always Count As Owned",
	"alwayscountsasowned":                   "Always Counts As Owned",
	"alwayscountsasunowned":                 "Always Counts As Unowned",
	"appid":                                 "App ID",
	"appidownedrequired":                    "App ID Owned Required",
	"billingagreementtype":                  "Billing Agreement Type",
	"blah":                                  "Blah",
	"canbegrantedfromexternal":              "Can Be Granted From External",
	"cantownapptopurchase":                  "Cant Own App To Purchase",
	"complimentarypackagegrant":             "Complimentary Package Grant",
	"complimentarypackagegrants":            "Complimentary Package Grants",
	"curatorconnect":                        "Curator Connect",
	"devcomp":                               "Devcomp",
	"dontallowrunincountries":               "Dont Allow Run In Countries",
	"dontgrantifappidowned":                 "Dont Grant If App ID Owned",
	"enforceintraeeaactivationrestrictions": "Enforce Intraeeaactivation Restrictions",
	"excludefromsharing":                    "Exclude From Sharing",
	"exfgls":                                "Exclude From Game Library Sharing",
	"expirytime":                            "Expiry Time",
	"extended":                              "Extended",
	"fakechange":                            "Fake Change",
	"foo":                                   "Foo",
	"freeondemand":                          "Free On Demand",
	"freeweekend":                           "Free Weekend",
	"full_gamepad":                          "Full Gamepad",
	"giftsaredeletable":                     "Gifts Are Deletable",
	"giftsaremarketable":                    "Gifts Are Marketable",
	"giftsaretradable":                      "Gifts Are Tradable",
	"grantexpirationdays":                   "Grant Expiration Days",
	"grantguestpasspackage":                 "Grant Guest Pass Package",
	"grantpassescount":                      "Grant Passes Count",
	"hardwarepromotype":                     "Hardware Promo Type",
	"ignorepurchasedateforrefunds":          "Ignore Purchase Date For Refunds",
	"initialperiod":                         "Initial Period",
	"initialtimeunit":                       "Initial Time Unit",
	"iploginrestriction":                    "IP Login Restriction",
	"languages":                             "Languages",
	"launcheula":                            "Launch EULA",
	"legacygamekeyappid":                    "Legacy Game Key App ID",
	"lowviolenceinrestrictedcountries":      "Low Violence In Restricted Countries",
	"martinotest":                           "Martino Test",
	"mustownapptopurchase":                  "Must Own App To Purchase",
	"onactivateguestpassmsg":                "On Activate Guest Pass Message",
	"onexpiredmsg":                          "On Expired Message",
	"ongrantguestpassmsg":                   "On Grant Guest Pass Message",
	"onlyallowincountries":                  "Only Allow In Countries",
	"onlyallowrestrictedcountries":          "Only Allow Restricted Countries",
	"onlyallowrunincountries":               "Only Allow Run In Countries",
	"onpurchasegrantguestpasspackage":       "On Purchase Grant Guest Pass Package",
	"onpurchasegrantguestpasspackage0":      "On Purchase Grant Guest Pass Package 0",
	"onpurchasegrantguestpasspackage1":      "On Purchase Grant Guest Pass Package 1",
	"onpurchasegrantguestpasspackage2":      "On Purchase Grant Guest Pass Package 2",
	"onpurchasegrantguestpasspackage3":      "On Purchase Grant Guest Pass Package 3",
	"onpurchasegrantguestpasspackage4":      "On Purchase Grant Guest Pass Package 4",
	"onpurchasegrantguestpasspackage5":      "On Purchase Grant Guest Pass Package 5",
	"onpurchasegrantguestpasspackage6":      "On Purchase Grant Guest Pass Package 6",
	"onpurchasegrantguestpasspackage7":      "On Purchase Grant Guest Pass Package 7",
	"onpurchasegrantguestpasspackage8":      "On Purchase Grant Guest Pass Package 8",
	"onpurchasegrantguestpasspackage9":      "On Purchase Grant Guest Pass Package 9",
	"onpurchasegrantguestpasspackage10":     "On Purchase Grant Guest Pass Package 10",
	"onpurchasegrantguestpasspackage11":     "On Purchase Grant Guest Pass Package 11",
	"onpurchasegrantguestpasspackage12":     "On Purchase Grant Guest Pass Package 12",
	"onpurchasegrantguestpasspackage13":     "On Purchase Grant Guest Pass Package 13",
	"onpurchasegrantguestpasspackage14":     "On Purchase Grant Guest Pass Package 14",
	"onpurchasegrantguestpasspackage15":     "On Purchase Grant Guest Pass Package 15",
	"onpurchasegrantguestpasspackage16":     "On Purchase Grant Guest Pass Package 16",
	"onpurchasegrantguestpasspackage17":     "On Purchase Grant Guest Pass Package 17",
	"onpurchasegrantguestpasspackage18":     "On Purchase Grant Guest Pass Package 18",
	"onpurchasegrantguestpasspackage19":     "On Purchase Grant Guest Pass Package 19",
	"onpurchasegrantguestpasspackage20":     "On Purchase Grant Guest Pass Package 20",
	"onpurchasegrantguestpasspackage21":     "On Purchase Grant Guest Pass Package 21",
	"onpurchasegrantguestpasspackage22":     "On Purchase Grant Guest Pass Package 22",
	"onquitguestpassmsg":                    "On Quit Guest Pass Message",
	"overridetaxtype":                       "Override Tax Type",
	"permitrunincountries":                  "Permit Run In Countries",
	"prohibitrunincountries":                "Prohibit Run In Countries",
	"purchaserestrictedcountries":           "Purchase Restricted Countries",
	"purchaseretrictedcountries":            "Purchase Restricted Countries",
	"recurringoptions":                      "Recurring Options",
	"recurringpackageoption":                "Recurring Package Option",
	"releaseoverride":                       "Release Override",
	"releasestatecountries":                 "Release State Countries",
	"releasestateoverride":                  "Release State Override",
	"releasestateoverridecountries":         "Release State Override Countries",
	"relesestateoverride":                   "Release State Override",
	"renewalperiod":                         "Renewal Period",
	"renewaltimeunit":                       "Renewal Time Unit",
	"requiredps3apploginforpurchase":        "Required PS3 App Login For Purchase",
	"requirespreapproval":                   "Requires Preapproval",
	"restrictedcountries":                   "Restricted Countries",
	"runrestrictedcountries":                "Run Restricted Countries",
	"shippableitem":                         "Shippable Item",
	"skipownsallappsinpackagecheck":         "Skip Owns All Apps In Package Check",
	"starttime":                             "Start Time",
	"state":                                 "State",
	"test":                                  "Test",
	"testchange":                            "Test Change",
	"trading_card_drops":                    "Trading Card Drops",
	"violencerestrictedcountries":           "Violence Restricted Countries",
	"violencerestrictedterritorycodes":      "Violence Restricted Territory Codes",
	"virtualitemreward":                     "Virtual Item Reward",
}

var PackageControllerKeys = map[string]string{
	"full_gamepad":                         "Full Gamepad",
	"allowpurchasefromrestrictedcountries": "Allow Purchase From Restricted Countries",
}
