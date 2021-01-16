package main

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/cmd/backend/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func (s Server) GetPackages(w http.ResponseWriter, _ *http.Request, params generated.GetPackagesParams) {

	var limit int64 = 10
	if params.Limit != nil && *params.Limit >= 1 && *params.Limit <= 1000 {
		limit = int64(*params.Limit)
	}

	var offset int64 = 0
	if params.Offset != nil {
		offset = int64(*params.Offset)
	}

	var sort = "_id"
	if params.Sort != nil {
		switch *params.Sort {
		case "id":
			sort = "_id"
		case "apps_count":
			sort = "_id"
		case "billing_type":
			sort = "_id"
		case "change_number_date":
			sort = "_id"
		case "license_type":
			sort = "_id"
		case "platforms":
			sort = "_id"
		case "status":
			sort = "_id"
		default:
			sort = "_id"
		}
	}

	var order = 1
	if params.Order != nil {
		switch *params.Sort {
		case "1", "asc", "ascending":
			order = 1
		case "0", "-1", "desc", "descending":
			order = -1
		default:
			order = 1
		}
	}

	filter := bson.D{}

	if params.Ids != nil {
		filter = append(filter, bson.E{Key: "_id", Value: *params.Ids})
	}
	if params.BillingType != nil {
		filter = append(filter, bson.E{Key: "billing_type", Value: *params.BillingType})
	}
	if params.LicenseType != nil {
		filter = append(filter, bson.E{Key: "license_type", Value: *params.LicenseType})
	}

	if params.Status != nil {
		filter = append(filter, bson.E{Key: "status", Value: *params.Status})
	}

	projection := bson.M{
		"apps":               1,
		"apps_count":         1,
		"bundle_ids":         1,
		"billing_type":       1,
		"change_id":          1,
		"change_number_date": 1,
		"coming_soon":        1,
		"depot_ids":          1,
		"icon":               1,
		"_id":                1,
		"image_logo":         1,
		"image_page":         1,
		"license_type":       1,
		"name":               1,
		"platforms":          1,
		"prices":             1,
		"release_date":       1,
		"release_date_unix":  1,
		"status":             1,
	}

	packages, err := mongo.GetPackages(offset, limit, bson.D{{sort, order}}, filter, projection)
	if err != nil {
		returnErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	total, err := mongo.CountDocuments(mongo.CollectionPackages, filter, 0)
	if err != nil {
		log.ErrS(err)
	}

	result := generated.PackagesResponse{}
	result.Pagination.Fill(offset, limit, total)

	for _, pack := range packages {

		newPackage := generated.PackageSchema{
			Apps:             helpers.IntsToInt32s(pack.Apps),
			AppsCount:        int32(pack.AppsCount),
			BillingType:      pack.GetBillingType(),
			Bundle:           helpers.IntsToInt32s(pack.Bundles),
			ChangeId:         int32(pack.ChangeNumber),
			ChangeNumberDate: pack.ChangeNumberDate.Unix(),
			ComingSoon:       pack.ComingSoon,
			DepotIds:         helpers.IntsToInt32s(pack.Depots),
			Icon:             pack.GetIcon(),
			Id:               int32(pack.GetID()),
			ImageLogo:        pack.ImageLogo,
			ImagePage:        pack.ImagePage,
			LicenseType:      pack.GetLicenseType(),
			Name:             pack.GetName(),
			Platforms:        pack.Platforms,
			Prices: generated.PackageSchema_Prices{
				AdditionalProperties: map[string]generated.ProductPriceSchema{},
			},
			ReleaseDate:     pack.ReleaseDate,
			ReleaseDateUnix: pack.ReleaseDateUnix,
			Status:          pack.GetStatus(),
		}

		for k, price := range pack.GetPrices() {
			newPackage.Prices.AdditionalProperties[string(k)] = generated.ProductPriceSchema{
				Currency:        string(price.Currency),
				DiscountPercent: int32(price.DiscountPercent),
				Final:           int32(price.Final),
				Free:            price.Free,
				Individual:      int32(price.Individual),
				Initial:         int32(price.Initial),
			}
		}

		result.Packages = append(result.Packages, newPackage)
	}

	returnResponse(w, http.StatusOK, result)
}
