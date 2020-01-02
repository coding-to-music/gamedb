package mongo

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AppItem struct {
	AppID            int        `bson:"app_id"`
	Bundle           string     `bson:"bundle"`
	Commodity        bool       `bson:"commodity"`
	DateCreated      string     `bson:"date_created"`
	Description      string     `bson:"description"`
	DisplayType      string     `bson:"display_type"`
	DropInterval     int        `bson:"drop_interval"`
	DropMaxPerWindow int        `bson:"drop_max_per_window"`
	Exchange         []string   `bson:"exchange"`
	Hash             string     `bson:"hash"`
	IconURL          string     `bson:"icon_url"`
	IconURLLarge     string     `bson:"icon_url_large"`
	ItemDefID        int        `bson:"item_def_id"`
	ItemQuality      string     `bson:"item_quality"`
	Marketable       bool       `bson:"marketable"`
	Modified         string     `bson:"modified"`
	Name             string     `bson:"name"`
	Price            string     `bson:"price"`
	Promo            string     `bson:"promo"`
	Quantity         int        `bson:"quantity"`
	Tags             [][]string `bson:"tags"`
	Timestamp        string     `bson:"timestamp"`
	Tradable         bool       `bson:"tradable"`
	Type             string     `bson:"type"`
	WorkshopID       int64      `bson:"workshop_id"`
}

func (item AppItem) BSON() bson.D {

	return bson.D{
		{"_id", item.getKey()},
		{"app_id", item.AppID},
		{"bundle", item.Bundle},
		{"commodity", item.Commodity},
		{"date_created", item.DateCreated},
		{"description", item.Description},
		{"display_type", item.DisplayType},
		{"drop_interval", item.DropInterval},
		{"drop_max_per_window", item.DropMaxPerWindow},
		{"exchange", item.Exchange},
		{"hash", item.Hash},
		{"icon_url", item.IconURL},
		{"icon_url_large", item.IconURLLarge},
		{"item_def_id", item.ItemDefID},
		{"item_quality", item.ItemQuality},
		{"marketable", item.Marketable},
		{"modified", item.Modified},
		{"name", item.Name},
		{"price", item.Price},
		{"promo", item.Promo},
		{"quantity", item.Quantity},
		{"tags", item.Tags},
		{"timestamp", item.Timestamp},
		{"tradable", item.Tradable},
		{"type", item.Type},
		{"workshop_id", item.WorkshopID},
	}
}

func (item AppItem) getKey() string {
	return strconv.Itoa(item.AppID) + "-" + strconv.Itoa(item.ItemDefID)
}

func (item AppItem) GetType() string {

	switch item.Type {
	default:
		return strings.Title(item.Type)
	}
}

func (item *AppItem) SetTags(tagsString string) {
	tagsString = strings.TrimSpace(tagsString)
	if tagsString != "" {
		tags := strings.Split(tagsString, ";")
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tagKeyVal := strings.Split(tag, ":")
				if len(tagKeyVal) == 1 {
					item.Tags = append(item.Tags, []string{tagKeyVal[0], ""})
				} else if len(tagKeyVal) == 2 {
					item.Tags = append(item.Tags, []string{tagKeyVal[0], tagKeyVal[1]})
				} else {
					log.Warning(item.AppID, "Weird tags")
				}
			}
		}
	}
}

func (item *AppItem) SetExchange(exchange string) {

	item.Exchange = []string{}

	split := strings.Split(exchange, ";")
	item.Exchange = append(item.Exchange, split...)
}

func (item *AppItem) Link() string {
	if !item.Marketable {
		return ""
	}
	return "https://steamcommunity.com/market/listings/" + strconv.Itoa(item.AppID) + "/" + url.PathEscape(item.Name)
}

func (item AppItem) ShortDescription() string {
	return helpers.TruncateString(item.Description, 150, "...")
}

func (item *AppItem) Image(size int, crop bool) string {

	if item.IconURL == "" {
		return ""
	}

	params := url.Values{}
	params.Set("url", item.IconURL)
	params.Set("w", strconv.Itoa(size))
	params.Set("h", strconv.Itoa(size))
	if crop {
		params.Set("t", "square")
	}

	return "https://images.weserv.nl?" + params.Encode()
}

func GetAppItems(offset int64, limit int64, filter bson.D, projection bson.M) (items []AppItem, err error) {

	var sort = bson.D{{"item_def_id", 1}}

	cur, ctx, err := Find(CollectionAppItems, offset, limit, sort, filter, projection, nil)
	if err != nil {
		return items, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		item := AppItem{}
		err := cur.Decode(&item)
		if err != nil {
			log.Err(err, item.getKey())
		} else {
			items = append(items, item)
		}
	}

	return items, cur.Err()
}

func UpdateAppItems(items []AppItem) (err error) {

	if len(items) < 1 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	var writes []mongo.WriteModel
	for _, item := range items {

		write := mongo.NewReplaceOneModel()
		write.SetFilter(bson.M{"_id": item.getKey()})
		write.SetReplacement(item.BSON())
		write.SetUpsert(true)

		writes = append(writes, write)
	}

	collection := client.Database(MongoDatabase).Collection(CollectionAppItems.String())
	_, err = collection.BulkWrite(ctx, writes, options.BulkWrite())
	return err
}

func DeleteAppItems(appID int, items []int) (err error) {

	if len(items) < 1 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	keys := bson.A{}
	for _, itemID := range items {

		item := AppItem{}
		item.ItemDefID = itemID
		item.AppID = appID

		keys = append(keys, item.getKey())
	}

	collection := client.Database(MongoDatabase).Collection(CollectionAppItems.String())
	_, err = collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": keys}})

	return err
}
