package queue

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/sql/pics"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/nlopes/slack"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type rabbitMessageProduct struct {
	ID           int                           `json:"ID"`
	ChangeNumber int                           `json:"ChangeNumber"`
	MissingToken bool                          `json:"MissingToken"`
	SHAHash      string                        `json:"SHAHash"`
	KeyValues    rabbitMessageProductKeyValues `json:"KeyValues"`
	OnlyPublic   bool                          `json:"OnlyPublic"`
	UseHTTP      bool                          `json:"UseHttp"`
	HTTPURI      interface{}                   `json:"HttpUri"`
}

type rabbitMessageProductKeyValues struct {
	Name     string                          `json:"Name"`
	Value    interface{}                     `json:"Value"`
	Children []rabbitMessageProductKeyValues `json:"Children"`
}

func (i rabbitMessageProductKeyValues) String() string {

	if i.Value == nil {
		b, err := json.Marshal(i.ToNestedMaps())
		if err != nil {
			logError(err)
			return ""
		}
		return string(b)
	}

	return i.Value.(string)
}

func (i rabbitMessageProductKeyValues) GetChildrenAsSlice() (ret []string) {
	for _, v := range i.Children {
		ret = append(ret, v.Value.(string))
	}
	return ret
}

func (i rabbitMessageProductKeyValues) GetChildrenAsMap() (ret map[string]string) {
	ret = map[string]string{}
	for _, v := range i.Children {
		ret[v.Name] = v.Value.(string)
	}
	return ret
}

// Turns it into nested maps
func (i rabbitMessageProductKeyValues) ToNestedMaps() (ret map[string]interface{}) {

	m := map[string]interface{}{}

	for _, v := range i.Children {

		if v.Value == nil {
			m[v.Name] = v.ToNestedMaps()
		} else {
			m[v.Name] = v.Value
		}
	}

	return m
}

func (i rabbitMessageProductKeyValues) GetExtended() (extended pics.PICSKeyValues) {

	extended = pics.PICSKeyValues{}
	for _, v := range i.Children {
		if v.Value == nil {
			b, err := json.Marshal(v.ToNestedMaps())
			logError(err)
			extended[v.Name] = string(b)
		} else {
			extended[v.Name] = v.Value.(string)
		}
	}
	return extended
}

func (i rabbitMessageProductKeyValues) GetAppConfig() (config pics.PICSKeyValues, launch []pics.PICSAppConfigLaunchItem) {

	config = pics.PICSKeyValues{}
	for _, v := range i.Children {
		if v.Name == "launch" {
			launch = v.GetAppLaunch()
		} else if v.Value == nil {
			b, err := json.Marshal(v.ToNestedMaps())
			logError(err)
			config[v.Name] = string(b)
		} else {
			config[v.Name] = v.Value.(string)
		}
	}

	return config, launch
}

func (i rabbitMessageProductKeyValues) GetAppDepots() (depots pics.PICSDepots) {

	depots.Extra = map[string]string{}

	// Loop depots
	for _, v := range i.Children {

		if v.Name == "branches" {
			depots.Branches = v.GetAppDepotBranches()
			continue
		}

		id, err := strconv.Atoi(v.Name)
		if err != nil {
			if v.Children == nil {
				depots.Extra[v.Name] = v.Value.(string)
			} else {
				b, err := json.Marshal(v.ToNestedMaps())
				logError(err)
				depots.Extra[v.Name] = string(b)
			}

			continue
		}

		depot := pics.PICSAppDepotItem{}
		depot.ID = id

		for _, vv := range v.Children {

			switch vv.Name {
			case "name":
				depot.Name = vv.Value.(string)
			case "config":
				depot.Configs = vv.GetChildrenAsMap()
			case "manifests":
				depot.Manifests = vv.GetChildrenAsMap()
			case "encryptedmanifests":
				b, err := json.Marshal(vv.ToNestedMaps())
				logError(err)
				depot.EncryptedManifests = string(b)
			case "maxsize":
				maxSize, err := strconv.ParseInt(vv.Value.(string), 10, 64)
				logError(err)
				depot.MaxSize = maxSize
			case "dlcappid":
				appID, err := strconv.Atoi(vv.Value.(string))
				logError(err)
				depot.DLCApp = appID
			case "depotfromapp":
				app, err := strconv.Atoi(vv.Value.(string))
				logError(err)
				depot.App = app
			case "systemdefined":
				if vv.Value.(string) == "1" {
					depot.SystemDefined = true
				}
			case "optional":
				if vv.Value.(string) == "1" {
					depot.Optional = true
				}
			case "sharedinstall":
				if vv.Value.(string) == "1" {
					depot.SharedInstall = true
				}
			case "shareddepottype":
				if vv.Value.(string) == "1" {
					depot.SharedDepotType = true
				}
			case "lvcache":
				if vv.Value.(string) == "1" {
					depot.LVCache = true
				}
			case "allowaddremovewhilerunning":
				if vv.Value.(string) == "1" {
					depot.AllowAddRemoveWhileRunning = true
				}
			default:
				logWarning("GetAppDepots missing case: " + vv.Name)
			}
		}

		depots.Depots = append(depots.Depots, depot)
	}

	return depots
}

func (i rabbitMessageProductKeyValues) GetAppDepotBranches() (branches []pics.PICSAppDepotBranches) {

	for _, v := range i.Children {

		branch := pics.PICSAppDepotBranches{}
		branch.Name = v.Name

		for _, vv := range v.Children {

			switch vv.Name {
			case "buildid":
				buildID, err := strconv.Atoi(vv.Value.(string))
				logError(err)
				branch.BuildID = buildID
			case "timeupdated":
				t, err := strconv.ParseInt(vv.Value.(string), 10, 64)
				logError(err)
				branch.TimeUpdated = t
			case "defaultforsubs":
				branch.DefaultForSubs = vv.Value.(string)
			case "unlockforsubs":
				branch.UnlockForSubs = vv.Value.(string)
			case "description":
				branch.Description = vv.Value.(string)
			case "pwdrequired":
				if vv.Value.(string) == "1" {
					branch.PasswordRequired = true
				}
			case "lcsrequired":
				if vv.Value.(string) == "1" {
					branch.LCSRequired = true
				}
			default:
				logWarning("GetAppDepotBranches missing case: " + vv.Name)
			}
		}

		branches = append(branches, branch)
	}

	return branches
}

func (i rabbitMessageProductKeyValues) GetAppLaunch() (items []pics.PICSAppConfigLaunchItem) {

	for _, v := range i.Children {

		item := pics.PICSAppConfigLaunchItem{}
		item.Order = v.Name

		v.setAppLaunchItem(&item)

		items = append(items, item)
	}

	return items
}

func (i rabbitMessageProductKeyValues) setAppLaunchItem(launchItem *pics.PICSAppConfigLaunchItem) {

	for _, v := range i.Children {

		switch v.Name {
		case "executable":
			launchItem.Executable = v.Value.(string)
		case "arguments":
			launchItem.Arguments = v.Value.(string)
		case "description":
			launchItem.Description = v.Value.(string)
		case "type":
			launchItem.Typex = v.Value.(string)
		case "oslist":
			launchItem.OSList = v.Value.(string)
		case "osarch":
			launchItem.OSArch = v.Value.(string)
		case "betakey":
			launchItem.BetaKey = v.Value.(string)
		case "vacmodulefilename":
			launchItem.VACModuleFilename = v.Value.(string)
		case "workingdir":
			launchItem.WorkingDir = v.Value.(string)
		case "vrmode":
			launchItem.VRMode = v.Value.(string)
		case "ownsdlc":
			DLCSlice := strings.Split(v.Value.(string), ",")
			for _, v := range DLCSlice {
				var trimmed = strings.TrimSpace(v)
				if trimmed != "" {
					launchItem.OwnsDLCs = append(launchItem.OwnsDLCs, trimmed)
				}
			}
		case "config":
			v.setAppLaunchItem(launchItem)
		default:
			logWarning("setAppLaunchItem missing case: " + v.Name)
		}
	}
}

func savePriceChanges(before sql.ProductInterface, after sql.ProductInterface) (err error) {

	var prices sql.ProductPrices
	var price sql.ProductPrice
	var documents []mongo.Document

	for _, productCC := range helpers.GetProdCCs(true) {

		var oldPrice, newPrice int

		prices, err = before.GetPrices()
		if err == nil {

			price = prices.Get(productCC.ProductCode)
			if !price.Exists {
				continue // Only compare if there is an old price to compare to
			}

			oldPrice = price.Final
		}

		prices, err = after.GetPrices()
		if err == nil {

			price = prices.Get(productCC.ProductCode)
			if !price.Exists {
				continue // Only compare if there is a new price to compare to
			}

			newPrice = price.Final
		}

		if oldPrice != newPrice {

			price := mongo.ProductPrice{}

			if after.GetProductType() == helpers.ProductTypeApp {
				price.AppID = after.GetID()
			} else if after.GetProductType() == helpers.ProductTypePackage {
				price.PackageID = after.GetID()
			} else {
				panic("Invalid productType")
			}

			price.Name = after.GetName()
			price.Icon = after.GetIcon()
			price.CreatedAt = time.Now()
			price.Currency = productCC.CurrencyCode
			price.ProdCC = productCC.ProductCode
			price.PriceBefore = oldPrice
			price.PriceAfter = newPrice
			price.Difference = newPrice - oldPrice
			price.DifferencePercent = (float64(newPrice-oldPrice) / float64(oldPrice)) * 100

			documents = append(documents, price)
		}

		// Tweet / Post to Reddit
		if productCC.ProductCode == steam.ProductCCUS &&
			before.GetProductType() == helpers.ProductTypeApp &&
			helpers.SliceHasString([]string{"Game", "Package"}, before.GetType()) &&
			helpers.PercentageChange(oldPrice, newPrice) < -80 &&
			newPrice > 0 {

			appBefore, ok := before.(sql.App)
			if ok && appBefore.IsOnSale() {

				price := "Down from $" + helpers.FloatToString(float64(oldPrice)/100, 2)

				// Twitter
				_, _, err = helpers.GetTwitter().Statuses.Update("Free game! "+price+" gamedb.online/apps/"+strconv.Itoa(before.GetID())+" #freegame #steam "+helpers.GetHashTag(before.GetName()), nil)
				if err != nil {
					if !strings.Contains(err.Error(), "Status is a duplicate") {
						logCritical(err)
					}
				}

				// Reddit
				err = helpers.PostToReddit("[FREE] "+before.GetName()+" ("+price+")", "https://gamedb.online"+before.GetPath())
				if err != nil {
					logCritical(err)
				}

				// Slack message
				err = slack.PostWebhook(config.Config.SlackSocialWebhook.Get(), &slack.WebhookMessage{
					Text: "Free game: https://gamedb.online" + before.GetPath(),
				})
				log.Err(err)
			}
		}
	}

	result, err := mongo.InsertDocuments(mongo.CollectionProductPrices, documents)

	// Send websockets to prices page
	if err == nil && result != nil {

		var priceIDs []string

		for _, v := range result.InsertedIDs {
			if s, ok := v.(primitive.ObjectID); ok {
				priceIDs = append(priceIDs, string(s[:]))
			}
		}

		if len(priceIDs) > 0 {

			log.Debug("price_ids", priceIDs)

			wsPayload := websockets.PubSubIDStringsPayload{}
			wsPayload.IDs = priceIDs
			wsPayload.Pages = []websockets.WebsocketPage{websockets.PagePrices}

			_, err2 := helpers.Publish(helpers.PubSubTopicWebsockets, wsPayload)
			if err2 != nil {
				logError(err2)
			}
		}
	}
	return err
}
