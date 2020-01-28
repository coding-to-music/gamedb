package mongo

import (
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/sql/pics"
	"go.mongodb.org/mongo-driver/bson"
)

type Package struct {
	AppIDs           []int                 `bson:"apps"`               //
	AppItems         map[int]int           `bson:"app_items"`          //
	AppsCount        int                   `bson:"apps_count"`         //
	BundleIDs        []int                 `bson:"bundle_ids"`         //
	BillingType      int                   `bson:"billing_type"`       //
	ChangeNumber     int                   `bson:"change_id"`          //
	ChangeNumberDate time.Time             `bson:"change_number_date"` //
	ComingSoon       bool                  `bson:"coming_soon"`        //
	Controller       map[string]bool       `bson:"controller"`         //
	CreatedAt        time.Time             `bson:"created_at"`         //
	DepotIDs         []int                 `bson:"depot_ids"`          //
	Extended         pics.PICSKeyValues    `bson:"extended"`           //
	Icon             string                `bson:"icon"`               //
	ID               int                   `bson:"_id" json:"id"`      //
	ImageLogo        string                `bson:"image_logo"`         //
	ImagePage        string                `bson:"image_page"`         //
	InStore          bool                  `bson:"in_store"`           // todo
	LicenseType      int8                  `bson:"license_type"`       //
	Name             string                `bson:"name"`               //
	Platforms        []string              `bson:"platforms"`          //
	Prices           helpers.ProductPrices `bson:"prices"`             //
	PurchaseText     string                `bson:"purchase_text"`      //
	ReleaseDate      string                `bson:"release_date"`       //
	ReleaseDateUnix  int64                 `bson:"release_date_unix"`  //
	Status           int8                  `bson:"status"`             //
	UpdatedAt        time.Time             `bson:"updated_at"`         //
}

func (pack Package) BSON() bson.D {

	pack.UpdatedAt = time.Now()

	return bson.D{

	}
}
