package queue

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/steam-go/steamvdf"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql/pics"
	"github.com/gamedb/gamedb/pkg/websockets"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func getAppConfig(kv steamvdf.KeyValue) (config pics.PICSKeyValues, launch []pics.PICSAppConfigLaunchItem) {

	config = pics.PICSKeyValues{}
	for _, v := range kv.Children {
		if v.Key == "launch" {
			launch = getAppLaunch(v)
		} else if len(v.Children) > 0 {
			b, err := json.Marshal(v.ToMapOuter())
			if err != nil {
				log.ErrS(err)
			}
			config[v.Key] = string(b)
		} else {
			config[v.Key] = v.Value
		}
	}

	return config, launch
}

func getAppDepots(kv steamvdf.KeyValue) (depots pics.Depots) {

	depots.Extra = map[string]string{}

	// Loop depots
	for _, v := range kv.Children {

		if v.Key == "branches" {
			depots.Branches = getAppDepotBranches(v)
			continue
		}

		id, err := strconv.Atoi(v.Key)
		if err != nil {
			if v.Children == nil {
				depots.Extra[v.Key] = v.Value
			} else {
				b, err := json.Marshal(v.ToMapOuter())
				if err != nil {
					log.ErrS(err)
				}
				depots.Extra[v.Key] = string(b)
			}

			continue
		}

		depot := pics.AppDepotItem{}
		depot.ID = id

		for _, vv := range v.Children {

			switch vv.Key {
			case "name":
				depot.Name = vv.Value
			case "config":
				depot.Configs = vv.GetChildrenAsMap()
			case "manifests":
				depot.Manifests = vv.GetChildrenAsMap()
			case "encryptedmanifests":
				b, err := json.Marshal(vv.ToMapOuter())
				if err != nil {
					log.ErrS(err)
				}
				depot.EncryptedManifests = string(b)
			case "maxsize":
				maxSize, err := strconv.ParseUint(vv.Value, 10, 64)
				if err != nil {
					log.ErrS(err)
				}
				depot.MaxSize = maxSize
			case "dlcappid":
				appID, err := strconv.Atoi(vv.Value)
				if err != nil {
					log.ErrS(err)
				}
				depot.DLCApp = appID
			case "depotfromapp":
				id := helpers.RegexNonInts.ReplaceAllString(vv.Value, "")
				app, err := strconv.Atoi(id)
				if err != nil {
					log.ErrS(err)
				}
				depot.App = app
			case "systemdefined":
				if vv.Value == "1" {
					depot.SystemDefined = true
				}
			case "optional":
				if vv.Value == "1" {
					depot.Optional = true
				}
			case "sharedinstall":
				if vv.Value == "1" {
					depot.SharedInstall = true
				}
			case "shareddepottype":
				if vv.Value == "1" {
					depot.SharedDepotType = true
				}
			case "lvcache":
				if vv.Value == "1" {
					depot.LVCache = true
				}
			case "allowaddremovewhilerunning":
				if vv.Value == "1" {
					depot.AllowAddRemoveWhileRunning = true
				}
			default:
				log.WarnS("GetAppDepots missing case: " + vv.Key)
			}
		}

		depots.Depots = append(depots.Depots, depot)
	}

	return depots
}

func getAppDepotBranches(kv steamvdf.KeyValue) (branches []pics.AppDepotBranches) {

	for _, v := range kv.Children {

		branch := pics.AppDepotBranches{}
		branch.Name = v.Key

		for _, vv := range v.Children {

			switch vv.Key {
			case "buildid":
				buildID, err := strconv.Atoi(vv.Value)
				if err != nil {
					log.ErrS(err)
				}
				branch.BuildID = buildID
			case "timeupdated":
				t, err := strconv.ParseInt(vv.Value, 10, 64)
				if err != nil {
					log.ErrS(err)
				}
				branch.TimeUpdated = t
			case "defaultforsubs":
				branch.DefaultForSubs = vv.Value
			case "unlockforsubs":
				branch.UnlockForSubs = vv.Value
			case "description":
				branch.Description = vv.Value
			case "pwdrequired":
				if vv.Value == "1" {
					branch.PasswordRequired = true
				}
			case "lcsrequired":
				if vv.Value == "1" {
					branch.LCSRequired = true
				}
			default:
				log.WarnS("GetAppDepotBranches missing case: " + vv.Key)
			}
		}

		branches = append(branches, branch)
	}

	return branches
}

func getAppLaunch(kv steamvdf.KeyValue) (items []pics.PICSAppConfigLaunchItem) {

	for _, v := range kv.Children {

		item := pics.PICSAppConfigLaunchItem{}
		item.Order = v.Key

		setAppLaunchItem(v, &item)

		items = append(items, item)
	}

	return items
}

func setAppLaunchItem(kv steamvdf.KeyValue, launchItem *pics.PICSAppConfigLaunchItem) {

	for _, child := range kv.Children {

		switch child.Key {
		case "executable":
			launchItem.Executable = child.Value
		case "arguments":
			launchItem.Arguments = child.Value
		case "description":
			launchItem.Description = child.Value
		case "type":
			launchItem.Typex = child.Value
		case "oslist":
			launchItem.OSList = child.Value
		case "osarch":
			launchItem.OSArch = child.Value
		case "betakey":
			launchItem.BetaKey = child.Value
		case "vacmodulefilename":
			launchItem.VACModuleFilename = child.Value
		case "workingdir":
			launchItem.WorkingDir = child.Value
		case "vrmode":
			launchItem.VRMode = child.Value
		case "ownsdlc":
			dlcSlice := strings.Split(child.Value, ",")
			for _, v := range dlcSlice {
				var trimmed = strings.TrimSpace(v)
				if trimmed != "" {
					launchItem.OwnsDLCs = append(launchItem.OwnsDLCs, trimmed)
				}
			}
		case "config":
			setAppLaunchItem(child, launchItem)
		default:
			log.WarnS("setAppLaunchItem missing case: " + child.Key)
		}
	}
}

func saveProductPricesToMongo(before helpers.ProductInterface, after helpers.ProductInterface) (err error) {

	var prices helpers.ProductPrices
	var price helpers.ProductPrice
	var documents []mongo.Document

	for _, productCC := range i18n.GetProdCCs(true) {

		var oldPrice, newPrice int

		// Before price
		prices = before.GetPrices()
		price = prices.Get(productCC.ProductCode)
		if !price.Exists {
			continue // Only compare if there is an old price to compare to
		}

		oldPrice = price.Final

		// After price
		prices = after.GetPrices()
		price = prices.Get(productCC.ProductCode)
		if !price.Exists {
			continue // Only compare if there is a new price to compare to
		}

		newPrice = price.Final

		//
		if oldPrice != newPrice {

			price := mongo.ProductPrice{}

			if after.GetProductType() == helpers.ProductTypeApp {
				price.AppID = after.GetID()
			} else if after.GetProductType() == helpers.ProductTypePackage {
				price.PackageID = after.GetID()
			} else {
				return errors.New("invalid productType")
			}

			price.Name = after.GetName()
			price.Icon = after.GetIcon()
			price.CreatedAt = time.Now()
			price.Currency = productCC.CurrencyCode
			price.ProdCC = productCC.ProductCode
			price.PriceBefore = oldPrice
			price.PriceAfter = newPrice
			price.Difference = newPrice - oldPrice
			if oldPrice == 0 {
				price.DifferencePercent = 0
			} else {
				price.DifferencePercent = (float64(newPrice-oldPrice) / float64(oldPrice)) * 100
			}

			documents = append(documents, price)
		}

		// Tweet / Post to Reddit
		// var percentIncrease = helpers.PercentageChange(oldPrice, newPrice)
		//
		// if productCC.ProductCode == steam.ProductCCUS &&
		// 	before.GetProductType() == helpers.ProductTypeApp &&
		// 	helpers.SliceHasString([]string{"Game", "Package"}, before.GetType()) &&
		// 	percentIncrease <= -80 &&
		// 	oldPrice > newPrice && // Incase it goes from -90% to -80%
		// 	newPrice > 0 { // Free games are usually just removed from the store
		//
		// 	appBefore, ok := before.(sql.App)
		// 	if ok && appBefore.IsOnSale() {
		//
		// 		// Twitter
		// 		_, _, err = twitter.GetTwitter().Statuses.Update("["+helpers.FloatToString(percentIncrease, 0)+"%] ($"+helpers.FloatToString(float64(newPrice)/100, 2)+") gamedb.online/games/"+strconv.Itoa(before.GetID())+" #freegame #steam "+helpers.GetHashTag(before.GetName()), nil)
		// 		if err != nil {
		// 			if !strings.Contains(err.Error(), "Status is a duplicate") {
		// 				log.FatalS(err)
		// 			}
		// 		}
		//
		// 		// Reddit
		// 		err = reddit.PostToReddit("["+helpers.FloatToString(percentIncrease, 0)+"%] "+before.GetName()+" ($"+helpers.FloatToString(float64(newPrice)/100, 2)+")", "https://gamedb.online"+before.GetPath())
		// 		if err != nil {
		// 			log.FatalS(err)
		// 		}
		//
		// 		// Slack message
		// 		err = slack.PostWebhook(config.Config.SlackSocialWebhook, &slack.WebhookMessage{
		// 			Text: config.Config.GameDBDomain + before.GetPath(),
		// 		})
		// if err != nil {
		// 		log.ErrS(err)
		// }
		// 	}
		// }
	}

	result, err := mongo.InsertMany(mongo.CollectionProductPrices, documents)

	// Send websockets to prices page
	if err == nil && result != nil {

		var priceIDs []string

		for _, v := range result.InsertedIDs {
			if s, ok := v.(primitive.ObjectID); ok {
				priceIDs = append(priceIDs, s.Hex())
			}
		}

		if len(priceIDs) > 0 {

			wsPayload := StringsPayload{IDs: priceIDs}
			err2 := ProduceWebsocket(wsPayload, websockets.PagePrices)
			if err2 != nil {
				log.ErrS(err2)
			}
		}
	}
	return err
}
