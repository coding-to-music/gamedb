package queue

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/social"
)

type RabbitMessageProduct struct {
	ID           int                           `json:"ID"`
	ChangeNumber int                           `json:"ChangeNumber"`
	MissingToken bool                          `json:"MissingToken"`
	SHAHash      string                        `json:"SHAHash"`
	KeyValues    RabbitMessageProductKeyValues `json:"KeyValues"`
	OnlyPublic   bool                          `json:"OnlyPublic"`
	UseHTTP      bool                          `json:"UseHttp"`
	HTTPURI      interface{}                   `json:"HttpUri"`
}

type RabbitMessageProductKeyValues struct {
	Name     string                          `json:"Name"`
	Value    interface{}                     `json:"Value"`
	Children []RabbitMessageProductKeyValues `json:"Children"`
}

func (i RabbitMessageProductKeyValues) String() string {

	if i.Value == nil {
		b, err := json.Marshal(i.ToNestedMaps())
		if err != nil {
			log.Err(err)
			return ""
		}
		return string(b)
	} else {
		return i.Value.(string)
	}
}

func (i RabbitMessageProductKeyValues) GetChildrenAsSlice() (ret []string) {
	for _, v := range i.Children {
		ret = append(ret, v.Value.(string))
	}
	return ret
}

func (i RabbitMessageProductKeyValues) GetChildrenAsMap() (ret map[string]string) {
	ret = map[string]string{}
	for _, v := range i.Children {
		ret[v.Name] = v.Value.(string)
	}
	return ret
}

// Turns it into nested maps
func (i RabbitMessageProductKeyValues) ToNestedMaps() (ret map[string]interface{}) {

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

func (i RabbitMessageProductKeyValues) GetExtended() (extended db.PICSExtended) {

	extended = db.PICSExtended{}
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

func (i RabbitMessageProductKeyValues) GetAppConfig() (config db.PICSAppConfig, launch []db.PICSAppConfigLaunchItem) {

	config = db.PICSAppConfig{}
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

func (i RabbitMessageProductKeyValues) GetAppDepots() (depots db.PICSDepots) {

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

		depot := db.PICSAppDepotItem{}
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
				logWarning(log.SeverityWarning, "GetAppDepots missing case: "+vv.Name)
			}
		}

		depots.Depots = append(depots.Depots, depot)
	}

	return depots
}

func (i RabbitMessageProductKeyValues) GetAppDepotBranches() (branches []db.PICSAppDepotBranches) {

	for _, v := range i.Children {

		branch := db.PICSAppDepotBranches{}
		branch.Name = v.Name

		for _, vv := range v.Children {

			switch vv.Name {
			case "buildid":
				buildID, err := strconv.Atoi(vv.Value.(string))
				logError(err)
				branch.BuildID = buildID
			case "timeupdated":
				time, err := strconv.ParseInt(vv.Value.(string), 10, 64)
				logError(err)
				branch.TimeUpdated = time
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
				logWarning(log.SeverityWarning, "GetAppDepotBranches missing case: "+vv.Name)
			}
		}

		branches = append(branches, branch)
	}

	return branches
}

func (i RabbitMessageProductKeyValues) GetAppLaunch() (items []db.PICSAppConfigLaunchItem) {

	for _, v := range i.Children {

		order, err := strconv.Atoi(v.Name)
		logError(err)

		item := db.PICSAppConfigLaunchItem{}
		item.Order = order

		v.getAppLaunchItem(&item)

		items = append(items, item)
	}

	return items
}

func (i RabbitMessageProductKeyValues) getAppLaunchItem(launchItem *db.PICSAppConfigLaunchItem) {

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
			v.getAppLaunchItem(launchItem)
		default:
			logWarning(log.SeverityWarning, "getAppLaunchItem missing case: "+v.Name)
		}
	}
}

// Save prices
func savePriceChanges(before db.ProductInterface, after db.ProductInterface) (err error) {

	var prices db.ProductPrices
	var price db.ProductPriceStruct
	var kinds []db.Kind
	for code := range steam.Countries {

		var oldPrice, newPrice int

		prices, err = before.GetPrices()
		if err == nil {
			price, err = prices.Get(code)
			if err == nil {
				oldPrice = price.Final
			} else {
				continue // Only compare if there is an old price to compare to
			}
		}

		prices, err = after.GetPrices()
		if err == nil {
			price, err = prices.Get(code)
			if err == nil {
				newPrice = price.Final
			} else {
				continue // Only compare if there is a new price to compare to
			}
		}

		if oldPrice != newPrice {
			kinds = append(kinds, db.CreateProductPrice(after, code, oldPrice, newPrice))
		}

		// Tweet free US products
		if code == steam.CountryUS && oldPrice > 0 && newPrice == 0 {

			twitter := social.GetTwitter()

			_, _, err = twitter.Statuses.Update(before.GetName()+" is now free! gamedb.online"+before.GetPath()+" #freegame #steam", nil)
			if err != nil {
				log.Critical(err)
			}
		}
	}

	return db.BulkSaveKinds(kinds, db.KindProductPrice, true)
}
