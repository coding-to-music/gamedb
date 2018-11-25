package queue

import (
	"encoding/json"
	"strconv"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/logging"
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
			logging.Error(err)
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
			bytes, err := json.Marshal(v.GetChildrenAsSlice())
			logging.Error(err)
			config[v.Name] = string(bytes)
		} else {
			config[v.Name] = v.Value.(string)
		}
	}

	return config, launch
}

func (i RabbitMessageProductKeyValues) GetAppDepots() (depots []db.PICSAppDepot, branches []db.PICSAppDepotBranches) {

	// Loop depots
	for _, v := range i.Children {

		if v.Name == "branches" {
			branches = v.GetAppDepotBranches()
			continue
		}

		id, err := strconv.Atoi(v.Name)
		logging.Error(err)

		depot := db.PICSAppDepot{}
		depot.ID = id

		for _, vv := range v.Children {

			var value = vv.Value.(string)

			switch vv.Name {
			case "name":
				depot.Name = value
			case "config":
				depot.Configs = vv.GetChildrenAsMap()
			case "manifests":
				depot.Manifests = vv.GetChildrenAsMap()
			case "encryptedmanifests":
				manifests, err := json.Marshal(vv.ToNestedMaps())
				logging.Error(err)
				depot.EncryptedManifests = string(manifests)
			case "maxsize":
				maxSize, err := strconv.ParseInt(value, 10, 64)
				logging.Error(err)
				depot.MaxSize = maxSize
			case "dlcappid":
				appID, err := strconv.Atoi(value)
				logging.Error(err)
				depot.DLCApp = appID
			case "depotfromapp":
				app, err := strconv.Atoi(value)
				logging.Error(err)
				depot.App = app
			case "systemdefined":
				if value == "1" {
					depot.SystemDefined = true
				}
			case "optional":
				if value == "1" {
					depot.Optional = true
				}
			default:
				logging.Info("GetAppDepots missing case: " + vv.Name)
			}
		}

		depots = append(depots, depot)
	}

	return depots, branches
}

func (i RabbitMessageProductKeyValues) GetAppDepotBranches() (branches []db.PICSAppDepotBranches) {

	for _, v := range i.Children {

		branch := db.PICSAppDepotBranches{}
		branch.Name = v.Name

		// 767390: hasdepotsindlc
		// 438100: baselanguages

		for _, vv := range v.Children {

			switch vv.Name {
			case "buildid":
				buildID, err := strconv.Atoi(vv.Value.(string))
				logging.Error(err)
				branch.BuildID = buildID
			case "timeupdated":
				time, err := strconv.ParseInt(vv.Value.(string), 10, 64)
				logging.Error(err)
				branch.TimeUpdated = time
			case "description":
				branch.Description = vv.Value.(string)
			case "pwdrequired":
				if vv.Value.(string) == "1" {
					branch.PasswordRequired = true
				}
			default:
				logging.Info("GetAppDepotBranches missing case: " + vv.Name)
			}
		}

		branches = append(branches, branch)
	}

	return branches
}

func (i RabbitMessageProductKeyValues) GetAppLaunch() (items []db.PICSAppConfigLaunchItem) {

	for _, v := range i.Children {

		order, err := strconv.Atoi(v.Name)
		logging.Error(err)

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
		case "ownsdlc":
			launchItem.OSArch = v.Value.(string)
		case "config":
			v.getAppLaunchItem(launchItem)
		default:
			logging.Info("getAppLaunchItem missing case: " + v.Name)
		}
	}
}
