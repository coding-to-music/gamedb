package pics

import (
	"encoding/json"
	"html/template"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
)

type PicsItemType string

const (
	picsTypeBool               PicsItemType = "bool"
	picsTypeLink               PicsItemType = "link"
	picsTypeImage              PicsItemType = "image"
	picsTypeTimestamp          PicsItemType = "timestamp"
	picsTypeJSON               PicsItemType = "json"
	picsTypeBytes              PicsItemType = "bytes"
	picsTypeNumber             PicsItemType = "number"
	picsTypeNumberListJSON     PicsItemType = "number-list-json"      // From JSON object
	picsTypeNumberListJSONKeys PicsItemType = "number-list-json-keys" // From JSON object keys
	picsTypeNumberListString   PicsItemType = "number-list-string"    // From comma string
	picsTypeTextListString     PicsItemType = "text-list-string"      // From comma string
	picsTypeTitle              PicsItemType = "title"
	picsTypeCustom             PicsItemType = "custom"
	picsTypeMap                PicsItemType = "map"
)

var CommonKeys = map[string]PicsKey{
	"associations":               {Type: picsTypeCustom},
	"category":                   {Type: picsTypeCustom},
	"clienticns":                 {Type: picsTypeLink, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.icns"},
	"clienticon":                 {Type: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.ico"},
	"clienttga":                  {Type: picsTypeLink, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.tga"},
	"community_hub_visible":      {Type: picsTypeBool},
	"community_visible_stats":    {Type: picsTypeBool},
	"controllervr":               {Type: picsTypeNumberListJSONKeys},
	"eulas":                      {Type: picsTypeJSON},
	"exfgls":                     {Type: picsTypeBool, Description: "Exclude from game library sharing"},
	"gameid":                     {Type: picsTypeLink, Link: "/apps/$val$"},
	"genres":                     {Type: picsTypeNumberListJSON, Link: "/apps?genres=$val$"},
	"has_adult_content":          {Type: picsTypeBool},
	"header_image":               {Type: picsTypeMap},
	"icon":                       {Type: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.jpg"},
	"languages":                  {Type: picsTypeCustom},
	"library_assets":             {Type: picsTypeJSON},
	"playareavr":                 {Type: picsTypeJSON},
	"openvr_controller_bindings": {Type: picsTypeJSON},
	"linuxclienticon":            {Type: picsTypeLink, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.zip"},
	"logo":                       {Type: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.jpg"},
	"logo_small":                 {Type: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.jpg"},
	"metacritic_fullurl":         {Type: picsTypeLink, Link: "$val$"},
	"original_release_date":      {Type: picsTypeTimestamp},
	"primary_genre":              {Type: picsTypeLink, Link: "/apps?genres=$val$"},
	"small_capsule":              {Type: picsTypeMap},
	"steam_release_date":         {Type: picsTypeTimestamp},
	"store_asset_mtime":          {Type: picsTypeTimestamp},
	"store_tags":                 {Type: picsTypeNumberListJSON, Link: "/apps?tags=$val$"},
	"supported_languages":        {Type: picsTypeCustom},
	"workshop_visible":           {Type: picsTypeBool},
	"releasestate":               {Type: picsTypeTitle},
	"type":                       {Type: picsTypeTitle},
	"controller_support":         {Type: picsTypeTitle},
	"oslist":                     {Type: picsTypeTextListString, Link: "/apps?platforms=$val$"},
	"metacritic_score":           {Type: picsTypeCustom},
	"onlyvrsupport":              {Type: picsTypeBool},
}

var ExtendedKeys = map[string]PicsKey{
	"anti_cheat_support_url":               {Type: picsTypeLink, Link: "$val$"},
	"developer_url":                        {Type: picsTypeLink, Link: "$val$"},
	"gamemanualurl":                        {Type: picsTypeLink, Link: "$val$"},
	"homepage":                             {Type: picsTypeLink, Link: "$val$"},
	"isfreeapp":                            {Type: picsTypeBool},
	"loadallbeforelaunch":                  {Type: picsTypeBool},
	"noservers":                            {Type: picsTypeBool},
	"requiressse":                          {Type: picsTypeBool},
	"sourcegame":                           {Type: picsTypeBool},
	"vacmacmodulecache":                    {Type: picsTypeLink, Link: "/apps/$val$"},
	"vacmodulecache":                       {Type: picsTypeLink, Link: "/apps/$val$"},
	"allowcrossregiontradingandgifting":    {Type: picsTypeBool},
	"allowpurchasefromrestrictedcountries": {Type: picsTypeBool},
	"listofdlc":                            {Type: picsTypeNumberListString, Link: "/apps/$val$"},
	"dlcavailableonstore":                  {Type: picsTypeBool},
	"validoslist":                          {Type: picsTypeTextListString, Link: "/apps?platforms=$val$"},
	"languages":                            {Type: picsTypeTextListString},
	"visibleonlywheninstalled":             {Type: picsTypeBool},
	"visibleonlywhensubscribed":            {Type: picsTypeBool},
	"vrheadsetstreaming":                   {Type: picsTypeBool},
}

var ConfigKeys = map[string]PicsKey{
	"checkforupdatesbeforelaunch":  {Type: picsTypeBool},
	"signedfiles":                  {Type: picsTypeJSON},
	"steamcontrollerconfigdetails": {Type: picsTypeJSON},
	"steamcontrollertemplateindex": {Type: picsTypeBool},
	"systemprofile":                {Type: picsTypeBool},
	"verifyupdates":                {Type: picsTypeBool},
	"vrcompositorsupport":          {Type: picsTypeBool},
	"launchwithoutworkshopupdates": {Type: picsTypeBool},
	"usemms":                       {Type: picsTypeBool},
}

var UFSKeys = map[string]PicsKey{
	"hidecloudui":   {Type: picsTypeBool},
	"maxnumfiles":   {Type: picsTypeNumber},
	"quota":         {Type: picsTypeBytes},
	"savefiles":     {Type: picsTypeJSON},
	"rootoverrides": {Type: picsTypeJSON},
}

type PicsKey struct {
	Type        PicsItemType
	Link        string
	Description string
}

func getType(key string, keys map[string]PicsKey) PicsItemType {

	if val, ok := keys[key]; ok {
		return val.Type
	}
	return ""
}

func getDescription(key string, keys map[string]PicsKey) string {

	if val, ok := keys[key]; ok {
		return val.Description
	}
	return ""
}

func FormatVal(key string, val string, appID int, keys map[string]PicsKey) interface{} {

	if item, ok := keys[key]; ok {
		switch item.Type {
		case picsTypeBool:

			b, _ := strconv.ParseBool(val)
			if b {
				return template.HTML("<i class=\"fas fa-check text-success\"></i>")
			}
			return template.HTML("<i class=\"fas fa-times text-danger\"></i>")

		case picsTypeLink:

			if val == "" {
				return ""
			}

			item.Link = strings.ReplaceAll(item.Link, "$val$", val)
			item.Link = strings.ReplaceAll(item.Link, "$app$", strconv.Itoa(appID))

			var blank string
			if !strings.HasPrefix(item.Link, "/") {
				blank = " rel=\"nofollow\" target=\"_blank\""
			}

			return template.HTML("<a href=\"" + item.Link + "\"" + blank + " rel=\"nofollow\">" + val + "</a>")

		case picsTypeImage:

			if val == "" {
				return ""
			}

			item.Link = strings.ReplaceAll(item.Link, "$val$", val)
			item.Link = strings.ReplaceAll(item.Link, "$app$", strconv.Itoa(appID))

			return template.HTML("<div class=\"icon-name\"><div class=\"icon\"><img class=\"wide\" src=\"" + item.Link + "\" /></div><div class=\"name\"><a href=\"" + item.Link + "\" rel=\"nofollow\" target=\"_blank\">" + val + "</a></div></div>")

		case picsTypeTimestamp:

			i, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return val
			}

			return time.Unix(i, 0).Format(helpers.DateTime)

		case picsTypeBytes:

			i, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				return val
			}

			return humanize.Bytes(i)

		case picsTypeNumber:

			i, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return val
			}

			return humanize.Comma(i)

		case picsTypeJSON:

			j, err := helpers.FormatJSON(val)
			if err != nil {
				return val
			}

			return template.HTML("<div class=\"json\">" + j + "</div>")

		case picsTypeNumberListString:

			var idSlice []string

			ids := strings.Split(val, ",")
			for _, id := range ids {
				id = strings.TrimSpace(id)
				idSlice = append(idSlice, id)
			}

			sort.Slice(idSlice, func(i, j int) bool {
				a, _ := strconv.Atoi(idSlice[i])
				b, _ := strconv.Atoi(idSlice[j])
				return a < b
			})

			if item.Link != "" {
				for k, id := range idSlice {
					idSlice[k] = "<a href=\"" + strings.ReplaceAll(item.Link, "$val$", id) + "\" rel=\"nofollow\">" + id + "</a>"
				}
			}

			return template.HTML(strings.Join(idSlice, ", "))

		case picsTypeTitle:

			return strings.Title(val)

		case picsTypeNumberListJSON:

			idMap := map[string]string{}

			err := json.Unmarshal([]byte(val), &idMap)
			log.Err(err)

			// Check for missing fields
			go log.Err(helpers.UnmarshalStrict([]byte(val), &idMap))

			var idSlice []string

			for _, id := range idMap {
				idSlice = append(idSlice, id)
			}

			sort.Slice(idSlice, func(i, j int) bool {
				a, _ := strconv.Atoi(idSlice[i])
				b, _ := strconv.Atoi(idSlice[j])
				return a < b
			})

			if item.Link != "" {
				for k, id := range idSlice {
					idSlice[k] = "<a href=\"" + strings.ReplaceAll(item.Link, "$val$", id) + "\" rel=\"nofollow\">" + id + "</a>"
				}
			}

			return template.HTML(strings.Join(idSlice, ", "))

		case picsTypeNumberListJSONKeys:

			idMap := map[string]string{}

			err := json.Unmarshal([]byte(val), &idMap)
			log.Err(err)

			// Check for missing fields
			go log.Err(helpers.UnmarshalStrict([]byte(val), &idMap))

			var idSlice []string

			for k := range idMap {
				idSlice = append(idSlice, k)
			}

			sort.Slice(idSlice, func(i, j int) bool {
				return idSlice[i] < idSlice[j]
			})

			if item.Link != "" {
				for k, id := range idSlice {
					idSlice[k] = "<a href=\"" + strings.ReplaceAll(item.Link, "$val$", id) + "\" rel=\"nofollow\">" + id + "</a>"
				}
			}

			return strings.Join(idSlice, ", ")

		case picsTypeTextListString:

			var idSlice []string

			ids := strings.Split(val, ",")
			for _, id := range ids {
				id = strings.TrimSpace(id)
				idSlice = append(idSlice, id)
			}

			sort.Slice(idSlice, func(i, j int) bool {
				return idSlice[i] < idSlice[j]
			})

			if item.Link != "" {
				for k, id := range idSlice {
					idSlice[k] = "<a href=\"" + strings.ReplaceAll(item.Link, "$val$", id) + "\" rel=\"nofollow\">" + id + "</a>"
				}
			}

			return template.HTML(strings.Join(idSlice, ", "))

		case picsTypeMap:

			if val != "" {

				m := map[string]string{}
				err := helpers.Unmarshal([]byte(val), &m)
				log.Err(err)

				var items []string
				for k, v := range m {
					items = append(items, "<li>"+k+": <span class=font-weight-bold>"+v+"</span></li>")
				}

				return template.HTML("<ul class='mb-0 pl-3'>" + strings.Join(items, "") + "</ul>")
			}

		case picsTypeCustom:

			switch key {
			case "supported_languages":

				if val != "" {

					langs := SupportedLanguages{}
					err := json.Unmarshal([]byte(val), &langs)
					log.Err(err)

					// Check for missing fields
					go log.Err(helpers.UnmarshalStrict([]byte(val), &langs))

					var items []string
					for code, lang := range langs {

						var item = code.Title()
						var features []string

						// if lang.Supported {
						// 	item += " <i class=\"fas fa-check text-success\"></i>"
						// } else {
						// 	item += " <i class=\"fas fa-times text-danger\"></i>"
						// }

						if lang.FullAudio {
							features = append(features, "Full Audio")
						}
						if lang.Subtitles {
							features = append(features, "Subtitles")
						}

						if len(features) > 0 {
							item += " + " + strings.Join(features, ", ")
						}

						items = append(items, item)
					}

					sort.Slice(items, func(i, j int) bool {
						return items[i] < items[j]
					})

					return template.HTML(strings.Join(items, ", "))
				}

			case "category":

				if val != "" {

					categories := map[string]string{}
					err := json.Unmarshal([]byte(val), &categories)
					log.Err(err)

					// Check for missing fields
					go log.Err(helpers.UnmarshalStrict([]byte(val), &categories))

					var items []int
					for k := range categories {

						i, err := strconv.Atoi(strings.Replace(k, "category_", "", 1))
						if err == nil {
							items = append(items, i)
						}
					}

					sort.Slice(items, func(i, j int) bool {
						return items[i] < items[j]
					})

					return helpers.JoinInts(items, ", ")
				}

			case "languages":

				if val != "" {

					languages := map[string]string{}
					err := json.Unmarshal([]byte(val), &languages)
					log.Err(err)

					// Check for missing fields
					go log.Err(helpers.UnmarshalStrict([]byte(val), &languages))

					var items []string
					for k, v := range languages {
						if v == "1" {
							items = append(items, k)
						}
					}

					sort.Slice(items, func(i, j int) bool {
						return items[i] < items[j]
					})

					return strings.Join(items, ", ")
				}

			case "associations":

				if val != "" {

					associations := Associations{}
					err := json.Unmarshal([]byte(val), &associations)
					log.Err(err)

					// Check for missing fields
					go log.Err(helpers.UnmarshalStrict([]byte(val), &associations))

					var items []string
					for _, v := range associations {
						items = append(items, "<li>"+strings.Title(v.Type)+": <span class=font-weight-bold>"+v.Name+"</span></li>")
					}

					return template.HTML("<ul class='mb-0 pl-3'>" + strings.Join(items, "") + "</ul>")
				}

			case "metacritic_score":
				return template.HTML(val + "<small>/100</small>")
			}

			return val
		}
	}

	return val
}
