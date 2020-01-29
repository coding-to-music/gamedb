package mongo

import (
	"errors"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql/pics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrInvalidPackageID = errors.New("invalid package id")
)

type Package struct {
	Apps             []int                 `bson:"apps"`
	AppItems         map[int]int           `bson:"app_items"`
	AppsCount        int                   `bson:"apps_count"`
	Bundles          []int                 `bson:"bundle_ids"`
	BillingType      int                   `bson:"billing_type"`
	ChangeNumber     int                   `bson:"change_id"`
	ChangeNumberDate time.Time             `bson:"change_number_date"`
	ComingSoon       bool                  `bson:"coming_soon"`
	Controller       map[string]bool       `bson:"controller"`
	CreatedAt        time.Time             `bson:"created_at"`
	Depots           []int                 `bson:"depot_ids"`
	Extended         pics.PICSKeyValues    `bson:"extended"`
	Icon             string                `bson:"icon"`
	ID               int                   `bson:"_id" json:"id"`
	ImageLogo        string                `bson:"image_logo"`
	ImagePage        string                `bson:"image_page"`
	InStore          bool                  `bson:"in_store"` // todo
	LicenseType      int8                  `bson:"license_type"`
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

func GetPackage(id int, projection bson.M) (pack Package, err error) {

	if !helpers.IsValidPackageID(id) {
		return pack, ErrInvalidPackageID
	}

	var item = memcache.MemcachePackageMongo(id, projection)

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &pack, func() (interface{}, error) {

		err := FindOne(CollectionPackages, bson.D{{"_id", id}}, nil, projection, &pack)
		if err != nil {
			return pack, err
		}
		if pack.ID == 0 {
			return pack, ErrNoDocuments
		}

		return pack, err
	})

	return pack, err
}

func GetPackagesByID(ids []int, projection bson.M) (packages []Package, err error) {

	a := bson.A{}
	for _, v := range ids {
		a = append(a, v)
	}

	return GetPackages(0, 0, nil, bson.D{{"_id", bson.M{"$in": a}}}, projection, nil)
}

func GetPackages(offset int64, limit int64, sort bson.D, filter bson.D, projection bson.M, ops *options.FindOptions) (packages []Package, err error) {

	cur, ctx, err := Find(CollectionPackages, offset, limit, sort, filter, projection, nil)
	if err != nil {
		return packages, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var pack Package
		err := cur.Decode(&pack)
		if err != nil {
			log.Err(err, pack.ID)
		} else {
			packages = append(packages, pack)
		}
	}

	return packages, cur.Err()
}
