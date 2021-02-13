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
	Apps            []int                      `json:"apps"`
	CreatedAt       int64                      `json:"created_at"`
	Discount        int                        `json:"discount"`
	DiscountHighest int                        `json:"discount_highest"`
	DiscountLowest  int                        `json:"discount_lowest"`
	DiscountSale    int                        `json:"discount_sale"`
	Giftable        bool                       `json:"giftable"`
	Icon            string                     `json:"icon"`
	ID              int                        `json:"id"`
	Image           string                     `json:"image"`
	Name            string                     `json:"name"`
	OnSale          bool                       `json:"on_sale"`
	Packages        []int                      `json:"packages"`
	Prices          map[steamapi.ProductCC]int `json:"prices"`
	PricesSale      map[steamapi.ProductCC]int `json:"prices_sale"`
	Type            string                     `json:"type"`
	UpdatedAt       int64                      `json:"updated_at"`
	NameMarked      string                     `json:"-"`
	Score           float64                    `json:"-"`
}

func (bundle Bundle) GetID() int {
	return bundle.ID
}

func (bundle Bundle) GetUpdated() time.Time {
	return time.Unix(bundle.UpdatedAt, 0)
}

func (bundle Bundle) GetDiscount() int {
	return bundle.Discount
}

func (bundle Bundle) GetDiscountHighest() int {
	return bundle.DiscountHighest
}

func (bundle Bundle) GetPrices() map[steamapi.ProductCC]int {
	return bundle.Prices
}

func (bundle Bundle) GetPricesFormatted() (ret map[steamapi.ProductCC]string) {

	ret = map[steamapi.ProductCC]string{}

	for k, v := range bundle.GetPrices() {
		ret[k] = i18n.FormatPrice(i18n.GetProdCC(k).CurrencyCode, v)
	}

	return ret
}

func (bundle Bundle) GetScore() float64 {
	return bundle.Score
}

func (bundle Bundle) GetApps() int {
	return len(bundle.Apps)
}

func (bundle Bundle) GetPackages() int {
	return len(bundle.Packages)
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

func (bundle Bundle) IsGiftable() bool {
	return bundle.Giftable
}

func SearchBundles(offset int, limit int, search string, sorters []elastic.Sorter, boolQuery *elastic.BoolQuery) (bundles []Bundle, total int64, err error) {

	client, ctx, err := GetElastic()
	if err != nil {
		return bundles, 0, err
	}

	searchService := client.Search().
		Index(IndexBundles).
		From(offset).
		Size(limit)

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
	} else {
		searchService.SortBy(sorters...)
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
				"apps":             fieldTypeInt32,
				"created_at":       fieldTypeInt64,
				"discount":         fieldTypeInt32,
				"discount_highest": fieldTypeInt32,
				"discount_lowest":  fieldTypeInt32,
				"discount_sale":    fieldTypeInt32,
				"giftable":         fieldTypeBool,
				"icon":             fieldTypeDisabled,
				"id":               fieldTypeInt32,
				"image":            fieldTypeDisabled,
				"name":             fieldTypeText,
				"on_sale":          fieldTypeBool,
				"packages":         fieldTypeInt32,
				"prices":           map[string]interface{}{"type": "object", "properties": priceProperties},
				"prices_sale":      map[string]interface{}{"type": "object", "properties": priceProperties},
				"type":             fieldTypeKeyword,
				"updated_at":       fieldTypeInt64,
			},
		},
	}

	rebuildIndex(IndexBundles, mapping)
}
