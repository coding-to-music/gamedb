package queue

import (
	"encoding/json"
	"strconv"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
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
			bytes, err := json.Marshal(v.GetChildrenAsSlice())
			log.Log(err)
			extended[v.Name] = string(bytes)
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
			bytes, err := json.Marshal(v.ToNestedMaps())
			log.Log(err)
			config[v.Name] = string(bytes)
		} else {
			config[v.Name] = v.Value.(string)
		}
	}

	return config, launch
}

func (i RabbitMessageProductKeyValues) GetAppDepots() (depots db.PicsDepots) {

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
				bytes, err := json.Marshal(v.ToNestedMaps())
				log.Log(err)
				depots.Extra[v.Name] = string(bytes)
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
				manifests, err := json.Marshal(vv.ToNestedMaps())
				log.Log(err)
				depot.EncryptedManifests = string(manifests)
			case "maxsize":
				maxSize, err := strconv.ParseInt(vv.Value.(string), 10, 64)
				log.Log(err)
				depot.MaxSize = maxSize
			case "dlcappid":
				appID, err := strconv.Atoi(vv.Value.(string))
				log.Log(err)
				depot.DLCApp = appID
			case "depotfromapp":
				app, err := strconv.Atoi(vv.Value.(string))
				log.Log(err)
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
			default:
				log.Log(log.SeverityInfo, "GetAppDepots missing case: "+vv.Name)
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
				log.Log(err)
				branch.BuildID = buildID
			case "timeupdated":
				time, err := strconv.ParseInt(vv.Value.(string), 10, 64)
				log.Log(err)
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
				log.Log(log.SeverityInfo, "GetAppDepotBranches missing case: "+vv.Name)
			}
		}

		branches = append(branches, branch)
	}

	return branches
}

func (i RabbitMessageProductKeyValues) GetAppLaunch() (items []db.PICSAppConfigLaunchItem) {

	for _, v := range i.Children {

		order, err := strconv.Atoi(v.Name)
		log.Log(err)

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
		case "ownsdlc":
			dlc, err := strconv.Atoi(v.Value.(string))
			log.Log(err)
			launchItem.OwnsDLC = dlc
		case "config":
			v.getAppLaunchItem(launchItem)
		default:
			log.Log(log.SeverityInfo, "getAppLaunchItem missing case: "+v.Name)
		}
	}
}

// Save prices
func savePriceChanges(before db.ProductInterface, after db.ProductInterface) (err error) {

	var prices db.ProductPrices
	var price db.ProductPriceCache
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
	}

	return db.BulkSaveKinds(kinds, db.KindProductPrice, true)
}
