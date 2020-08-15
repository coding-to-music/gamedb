package mongo

import (
	"errors"
	"html/template"
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/Philipp15b/go-steam/protocol/steamlang"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mysql/pics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

var (
	ErrInvalidPackageID = errors.New("invalid package id")
)

type Package struct {
	Apps             []int                 `bson:"apps"`
	AppItems         map[int]int           `bson:"app_items"`
	AppsCount        int                   `bson:"apps_count"`
	Bundles          []int                 `bson:"bundle_ids"`
	BillingType      int32                 `bson:"billing_type"`
	ChangeNumber     int                   `bson:"change_id"`
	ChangeNumberDate time.Time             `bson:"change_number_date"`
	ComingSoon       bool                  `bson:"coming_soon"`
	Controller       pics.PICSController   `bson:"controller"`
	CreatedAt        time.Time             `bson:"created_at"`
	Depots           []int                 `bson:"depot_ids"`
	Extended         pics.PICSKeyValues    `bson:"extended"`
	Icon             string                `bson:"icon"`
	ID               int                   `bson:"_id" json:"id"`
	ImageLogo        string                `bson:"image_logo"`
	ImagePage        string                `bson:"image_page"`
	InStore          bool                  `bson:"in_store"` // todo
	LicenseType      int32                 `bson:"license_type"`
	Name             string                `bson:"name"`
	Platforms        []string              `bson:"platforms"`
	Prices           helpers.ProductPrices `bson:"prices"`
	PurchaseText     string                `bson:"purchase_text"`
	ReleaseDate      string                `bson:"release_date"`
	ReleaseDateUnix  int64                 `bson:"release_date_unix"`
	Status           int8                  `bson:"status"`
	UpdatedAt        time.Time             `bson:"updated_at"`
}

func (pack Package) BSON() bson.D {

	pack.UpdatedAt = time.Now()

	return bson.D{
		{"apps", pack.Apps},
		{"apps_count", pack.AppsCount},
		{"app_items", pack.AppItems},
		{"bundless", pack.Bundles},
		{"billing_type", pack.BillingType},
		{"change_number", pack.ChangeNumber},
		{"change_number_date", pack.ChangeNumberDate},
		{"coming_soon", pack.ComingSoon},
		{"controller", pack.Controller},
		{"created_at", pack.CreatedAt},
		{"depots", pack.Depots},
		{"extended", pack.Extended},
		{"icon", pack.Icon},
		{"_id", pack.ID},
		{"image_logo", pack.ImageLogo},
		{"image_page", pack.ImagePage},
		{"in_store", pack.InStore},
		{"license_type", pack.LicenseType},
		{"name", pack.Name},
		{"platforms", pack.Platforms},
		{"prices", pack.Prices},
		{"purchase_text", pack.PurchaseText},
		{"release_date", pack.ReleaseDate},
		{"release_date_unix", pack.ReleaseDateUnix},
		{"status", pack.Status},
		{"updated_at", pack.UpdatedAt},
	}
}

//noinspection GoUnusedExportedFunction
func CreatePackageIndexes() {

	var indexModels []mongo.IndexModel

	var cols = []string{
		"apps_count",
		"billing_type",
		"change_number_date",
		"license_type",
		"platforms",
		"status",
	}

	// Price fields
	for _, v := range i18n.GetProdCCs(true) {
		cols = append(cols, "prices."+string(v.ProductCode)+".final")
		cols = append(cols, "prices."+string(v.ProductCode)+".discount_percent")
	}

	//
	for _, v := range cols {
		indexModels = append(indexModels, mongo.IndexModel{
			Keys: bson.D{{v, 1}},
		}, mongo.IndexModel{
			Keys: bson.D{{v, -1}},
		})
	}

	//
	client, ctx, err := getMongo()
	if err != nil {
		zap.S().Error(err)
		return
	}

	_, err = client.Database(MongoDatabase).Collection(CollectionPackages.String()).Indexes().CreateMany(ctx, indexModels)
	zap.S().Error(err)
}

func (pack Package) GetID() int {
	return pack.ID
}

func (pack Package) GetProductType() helpers.ProductType {
	return helpers.ProductTypePackage
}

func (pack Package) GetType() string {
	return "Package"
}

func (pack Package) GetName() (name string) {
	return helpers.GetPackageName(pack.ID, pack.Name)
}

func (pack Package) GetPath() string {
	return helpers.GetPackagePath(pack.ID, pack.GetName())
}

func (pack Package) GetIcon() string {
	if pack.Icon == "" {
		return helpers.DefaultAppIcon
	}
	return pack.Icon
}

func (pack Package) GetMetaImage() string {
	return pack.ImageLogo
}

func (pack Package) StoreLink() string {
	if !pack.InStore {
		return ""
	}
	return "https://store.steampowered.com/sub/" + strconv.Itoa(pack.ID) + "/?curator_clanid=&utm_source=GameDB" // todo curator_clanid
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

func (pack Package) GetBillingType() string {

	switch steamlang.EBillingType(pack.BillingType) {
	case steamlang.EBillingType_NoCost:
		return "No Cost"
	case steamlang.EBillingType_BillOnceOnly:
		return "Store"
	case steamlang.EBillingType_BillMonthly:
		return "Bill Monthly"
	case steamlang.EBillingType_ProofOfPrepurchaseOnly:
		return "CD Key"
	case steamlang.EBillingType_GuestPass:
		return "Guest Pass"
	case steamlang.EBillingType_HardwarePromo:
		return "Hardware Promo"
	case steamlang.EBillingType_Gift:
		return "Gift"
	case steamlang.EBillingType_AutoGrant:
		return "Free Weekend"
	case steamlang.EBillingType_OEMTicket:
		return "OEM Ticket"
	case steamlang.EBillingType_RecurringOption:
		return "Recurring Option"
	case steamlang.EBillingType_BillOnceOrCDKey:
		return "Store or CD Key"
	case steamlang.EBillingType_Repurchaseable:
		return "Repurchaseable"
	case steamlang.EBillingType_FreeOnDemand:
		return "Free on Demand"
	case steamlang.EBillingType_Rental:
		return "Rental"
	case steamlang.EBillingType_CommercialLicense:
		return "Commercial License"
	case steamlang.EBillingType_FreeCommercialLicense:
		return "Free Commercial License"
	case steamlang.EBillingType_NumBillingTypes:
		return "Num"
	default:
		zap.S().Warn("New billing type", pack.BillingType)
		return "Unknown"
	}
}

func (pack Package) GetLicenseType() string {

	switch steamlang.ELicenseType(pack.LicenseType) {
	case steamlang.ELicenseType_NoLicense:
		return "No License"
	case steamlang.ELicenseType_SinglePurchase:
		return "Single Purchase"
	case steamlang.ELicenseType_SinglePurchaseLimitedUse:
		return "Single Purchase (Limited Use)"
	case steamlang.ELicenseType_RecurringCharge:
		return "Recurring Charge"
	case steamlang.ELicenseType_RecurringChargeLimitedUse:
		return "Recurring Charge"
	case steamlang.ELicenseType_RecurringChargeLimitedUseWithOverages:
		return "Recurring Charge"
	case steamlang.ELicenseType_RecurringOption:
		return "Recurring"
	case steamlang.ELicenseType_LimitedUseDelayedActivation:
		return "Limited Use Delayed Activation"
	default:
		zap.S().Warn("New license type", pack.LicenseType)
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

func (pack Package) GetPlatformImages() (ret template.HTML) {

	for _, v := range pack.Platforms {
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

func (pack Package) GetPICSUpdatedNice() string {

	if pack.ChangeNumberDate.IsZero() || pack.ChangeNumberDate.Unix() == 0 {
		return "-"
	}
	return pack.ChangeNumberDate.Format(helpers.DateYearTime)
}

func (pack Package) GetUpdatedNice() string {
	return pack.UpdatedAt.Format(helpers.DateYearTime)
}

func (pack Package) GetPrices() (prices helpers.ProductPrices) {
	return pack.Prices
}

var PackageOutputForJSON = bson.M{"id": 1, "name": 1, "apps_count": 1, "prices": 1, "change_number_date": 1, "icon": 1, "billing_type": 1}

func (pack Package) OutputForJSON(code steamapi.ProductCC) (output []interface{}) {

	var changeNumberDate = pack.ChangeNumberDate.Format(helpers.DateYearTime)
	var discount = pack.Prices.Get(code).GetDiscountPercent()

	return []interface{}{
		pack.ID,                          // 0
		pack.GetPath(),                   // 1
		pack.GetName(),                   // 2
		"",                               // 3
		pack.AppsCount,                   // 4
		pack.Prices.Get(code).GetFinal(), // 5
		pack.ChangeNumberDate.Unix(),     // 6
		changeNumberDate,                 // 7
		pack.GetIcon(),                   // 8
		discount,                         // 9
		pack.StoreLink(),                 // 10
		pack.GetBillingType(),            // 11
	}
}

func (pack Package) Save() (err error) {

	if pack.CreatedAt.Unix() < 1 {
		pack.CreatedAt = time.Now()
	}

	pack.UpdatedAt = time.Now()

	_, err = ReplaceOne(CollectionPackages, bson.D{{"_id", pack.ID}}, pack)
	return err
}

func (pack Package) ShouldUpdate() bool {

	return pack.UpdatedAt.Before(time.Now().Add(time.Hour * 24 * -1))
}

func (pack *Package) SetName(name string, force bool) {
	if (pack.Name == "" || force) && name != "" {
		pack.Name = name
	}
}

func GetPackage(id int) (pack Package, err error) {

	if !helpers.IsValidPackageID(id) {
		return pack, ErrInvalidPackageID
	}

	var item = memcache.MemcachePackage(id)

	err = memcache.GetSetInterface(item.Key, item.Expiration, &pack, func() (interface{}, error) {

		err := FindOne(CollectionPackages, bson.D{{"_id", id}}, nil, nil, &pack)
		return pack, err
	})

	return pack, err
}

func GetPackagesByID(ids []int, projection bson.M) (packages []Package, err error) {

	a := bson.A{}
	for _, v := range ids {
		a = append(a, v)
	}

	return GetPackages(0, 0, nil, bson.D{{"_id", bson.M{"$in": a}}}, projection)
}

func GetPackages(offset int64, limit int64, sort bson.D, filter bson.D, projection bson.M) (packages []Package, err error) {

	cur, ctx, err := Find(CollectionPackages, offset, limit, sort, filter, projection, nil)
	if err != nil {
		return packages, err
	}

	defer func() {
		err = cur.Close(ctx)
		zap.S().Error(err)
	}()

	for cur.Next(ctx) {

		var pack Package
		err := cur.Decode(&pack)
		if err != nil {
			zap.S().Error(err, pack.ID)
		} else {
			packages = append(packages, pack)
		}
	}

	return packages, cur.Err()
}
