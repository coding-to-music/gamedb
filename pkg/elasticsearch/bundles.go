package elasticsearch

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/olivere/elastic/v7"
)

type Bundle struct {
	ID              int                        `json:"id"`
	UpdatedAt       time.Time                  `json:"updated_at"`
	Name            string                     `json:"name"`
	Discount        int                        `json:"discount"`
	SaleDiscount    int                        `json:"sale_discount"`
	HighestDiscount int                        `json:"highest_discount"`
	Apps            int                        `json:"apps"`
	Packages        int                        `json:"packages"`
	Icon            string                     `json:"icon"`
	Prices          map[steamapi.ProductCC]int `json:"prices"`
	SalePrices      map[steamapi.ProductCC]int `json:"sale_prices"`
	Type            string                     `json:"type"`
	NameMarked      string                     `json:"-"`
	Score           float64                    `json:"-"`
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
	return bundle.GetDiscountHighest()
}

func (bundle Bundle) GetPrices() map[steamapi.ProductCC]int {
	return bundle.Prices
}

func (bundle Bundle) GetScore() float64 {
	return bundle.Score
}

func (bundle Bundle) GetApps() int {
	return bundle.Apps
}

func (bundle Bundle) GetPackages() int {
	return bundle.Discount
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

func (bundle Bundle) OutputForJSON() (output []interface{}) {
	return helpers.OutputBundleForJSON(bundle)
}

func SearchBundles(offset int, limit int, search string, sorters []elastic.Sorter, boolQuery *elastic.BoolQuery) (bundles []Bundle, total int64, err error) {

	client, ctx, err := GetElastic()
	if err != nil {
		return bundles, 0, err
	}

	searchService := client.Search().
		Index(IndexBundles).
		From(offset).
		Size(limit).
		SortBy(sorters...)

	if boolQuery == nil {
		boolQuery = elastic.NewBoolQuery()
	}

	search = strings.TrimSpace(search)
	if search != "" {

		boolQuery.Must(
			elastic.NewBoolQuery().MinimumNumberShouldMatch(1).Should(
				elastic.NewTermQuery("id", search).Boost(5),
				elastic.NewMatchQuery("name", search),
			),
		)

		searchService.Highlight(elastic.NewHighlight().Field("name").PreTags("<mark>").PostTags("</mark>"))
	}

	searchService.Query(boolQuery)
	searchService.TrackTotalHits(true)

	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return bundles, 0, err
	}

	for _, hit := range searchResult.Hits.Hits {

		var bundle = Bundle{}

		err := json.Unmarshal(hit.Source, &bundle)
		if err != nil {
			log.ErrS(err)
			continue
		}

		if hit.Score != nil {
			bundle.Score = *hit.Score
		}

		bundle.NameMarked = bundle.Name
		if val, ok := hit.Highlight["name"]; ok {
			if len(val) > 0 {
				bundle.NameMarked = val[0]
			}
		}

		bundles = append(bundles, bundle)
	}

	return bundles, searchResult.TotalHits(), err
}

func IndexBundle(bundle Bundle) error {
	return indexDocument(IndexBundles, strconv.Itoa(bundle.ID), bundle)
}

//noinspection GoUnusedExportedFunction
func DeleteAndRebuildBundlesIndex() {

	var priceProperties = map[steamapi.ProductCC]interface{}{}
	for _, prodCC := range i18n.GetProdCCs(true) {
		priceProperties[prodCC.ProductCode] = fieldTypeInt32
	}

	var mapping = map[string]interface{}{
		"settings": settings,
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id":               fieldTypeInt32,
				"updated_at":       fieldTypeInt64,
				"name":             fieldTypeText,
				"discount":         fieldTypeInt32,
				"sale_discount":    fieldTypeInt32,
				"highest_discount": fieldTypeInt32,
				"apps":             fieldTypeInt32,
				"packages":         fieldTypeInt32,
				"icon":             fieldTypeDisabled,
				"prices":           map[string]interface{}{"type": "object", "properties": priceProperties},
				"sale_prices":      map[string]interface{}{"type": "object", "properties": priceProperties},
				"type":             fieldTypeKeyword,
				"giftable":         fieldTypeBool,
				"on_sale":          fieldTypeBool,
			},
		},
	}

	rebuildIndex(IndexBundles, mapping)
}
