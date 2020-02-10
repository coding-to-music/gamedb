package pics

import (
	"html/template"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
)

const (
	picsTypeBool = iota + 1
	picsTypeBytes
	picsTypeCustom
	picsTypeImage
	picsTypeJSON
	picsTypeLink
	picsTypeMap
	picsTypeNumber
	picsTypeNumberListJSON     // From JSON object
	picsTypeNumberListJSONKeys // From JSON object keys
	picsTypeNumberListString   // From comma string
	picsTypeStringListJSON     // From JSON bject
	picsTypeTextListString     // From comma string
	picsTypeTimestamp
	picsTypeTitle
	picsTypeTooLong
	picsTypePercent
)

var CommonKeys = map[string]PicsKey{
	"app_retired_publisher_request": {FormatType: picsTypeBool},
	"associations":                  {FormatType: picsTypeCustom},
	"category":                      {FormatType: picsTypeCustom},
	"clienticns":                    {FormatType: picsTypeLink, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.icns"},
	"clienticon":                    {FormatType: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.ico"},
	"clienttga":                     {FormatType: picsTypeLink, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.tga"},
	"community_hub_visible":         {FormatType: picsTypeBool},
	"community_visible_stats":       {FormatType: picsTypeBool},
	"controller_support":            {FormatType: picsTypeTitle},
	"controllervr":                  {FormatType: picsTypeNumberListJSONKeys},
	"eulas":                         {FormatType: picsTypeCustom},
	"exfgls":                        {FormatType: picsTypeBool, Description: "Exclude from game library sharing"},
	"gameid":                        {FormatType: picsTypeLink, Link: "/apps/$val$"},
	"genres":                        {FormatType: picsTypeNumberListJSON, Link: "/apps?genres=$val$"},
	"has_adult_content":             {FormatType: picsTypeBool},
	"header_image":                  {FormatType: picsTypeMap},
	"icon":                          {FormatType: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.jpg"},
	"languages":                     {FormatType: picsTypeCustom},
	"library_assets":                {FormatType: picsTypeJSON},
	"linuxclienticon":               {FormatType: picsTypeLink, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.zip"},
	"logo":                          {FormatType: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.jpg"},
	"logo_small":                    {FormatType: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.jpg"},
	"market_presence":               {FormatType: picsTypeBool},
	"metacritic_fullurl":            {FormatType: picsTypeLink, Link: "$val$"},
	"metacritic_score":              {FormatType: picsTypePercent},
	"name_localized":                {FormatType: picsTypeStringListJSON},
	"onlyvrsupport":                 {FormatType: picsTypeBool},
	"openvr_controller_bindings":    {FormatType: picsTypeJSON},
	"openvrsupport":                 {FormatType: picsTypeBool},
	"original_release_date":         {FormatType: picsTypeTimestamp},
	"oslist":                        {FormatType: picsTypeTextListString, Link: "/apps?platforms=$val$"},
	"parent":                        {FormatType: picsTypeLink, Link: "/apps/$val$"},
	"playareavr":                    {FormatType: picsTypeJSON},
	"primary_genre":                 {FormatType: picsTypeLink, Link: "/apps?genres=$val$"},
	"releasestate":                  {FormatType: picsTypeTitle},
	"requireskbmouse":               {FormatType: picsTypeBool},
	"review_percentage":             {FormatType: picsTypePercent},
	"small_capsule":                 {FormatType: picsTypeMap},
	"steam_release_date":            {FormatType: picsTypeTimestamp},
	"store_asset_mtime":             {FormatType: picsTypeTimestamp},
	"store_tags":                    {FormatType: picsTypeNumberListJSON, Link: "/apps?tags=$val$"},
	"supported_languages":           {FormatType: picsTypeCustom},
	"type":                          {FormatType: picsTypeTitle},
	"workshop_visible":              {FormatType: picsTypeBool},
}

var ExtendedKeys = map[string]PicsKey{
	"allowcrossregiontradingandgifting":    {FormatType: picsTypeBool},
	"allowpurchasefromrestrictedcountries": {FormatType: picsTypeBool},
	"anti_cheat_support_url":               {FormatType: picsTypeLink, Link: "$val$"},
	"curatorconnect":                       {FormatType: picsTypeBool},
	"developer_url":                        {FormatType: picsTypeLink, Link: "$val$"},
	"disableoverlayinjection":              {FormatType: picsTypeBool},
	"dlcavailableonstore":                  {FormatType: picsTypeBool},
	"gamemanualurl":                        {FormatType: picsTypeLink, Link: "$val$"},
	"homepage":                             {FormatType: picsTypeLink, Link: "$val$"},
	"isconverteddlc":                       {FormatType: picsTypeBool},
	"isfreeapp":                            {FormatType: picsTypeBool},
	"languages":                            {FormatType: picsTypeTextListString},
	"languages_macos":                      {FormatType: picsTypeTextListString},
	"listofdlc":                            {FormatType: picsTypeNumberListString, Link: "/apps/$val$"},
	"loadallbeforelaunch":                  {FormatType: picsTypeBool},
	"musicalbumforappid":                   {FormatType: picsTypeLink, Link: "/apps/$val$"},
	"noservers":                            {FormatType: picsTypeBool},
	"onlyallowrestrictedcountries":         {FormatType: picsTypeBool},
	"requiressse":                          {FormatType: picsTypeBool},
	"showcdkeyinmenu":                      {FormatType: picsTypeBool},
	"showcdkeyonlaunch":                    {FormatType: picsTypeBool},
	"sourcegame":                           {FormatType: picsTypeBool},
	"supportscdkeycopytoclipboard":         {FormatType: picsTypeBool},
	"vacmacmodulecache":                    {FormatType: picsTypeLink, Link: "/apps/$val$"},
	"vacmodulecache":                       {FormatType: picsTypeLink, Link: "/apps/$val$"},
	"validoslist":                          {FormatType: picsTypeTextListString, Link: "/apps?platforms=$val$"},
	"visibleonlywheninstalled":             {FormatType: picsTypeBool},
	"visibleonlywhensubscribed":            {FormatType: picsTypeBool},
	"vrheadsetstreaming":                   {FormatType: picsTypeBool},
}

var ConfigKeys = map[string]PicsKey{
	"cegpublickey":                      {FormatType: picsTypeTooLong},
	"checkforupdatesbeforelaunch":       {FormatType: picsTypeBool},
	"enabletextfiltering":               {FormatType: picsTypeBool},
	"installscriptoverride":             {FormatType: picsTypeBool},
	"installscriptsignature":            {FormatType: picsTypeTooLong},
	"launchwithoutworkshopupdates":      {FormatType: picsTypeBool},
	"matchmaking_uptodate":              {FormatType: picsTypeBool},
	"signaturescheckedonlaunch":         {FormatType: picsTypeJSON},
	"signedfiles":                       {FormatType: picsTypeJSON},
	"steamcontrollerconfigdetails":      {FormatType: picsTypeJSON},
	"steamcontrollertemplateindex":      {FormatType: picsTypeBool},
	"steamcontrollertouchconfigdetails": {FormatType: picsTypeJSON},
	"steamcontrollertouchtemplateindex": {FormatType: picsTypeBool},
	"systemprofile":                     {FormatType: picsTypeBool},
	"uselaunchcommandline":              {FormatType: picsTypeBool},
	"usemms":                            {FormatType: picsTypeBool},
	"usesfrenemies":                     {FormatType: picsTypeBool},
	"verifyupdates":                     {FormatType: picsTypeBool},
	"vrcompositorsupport":               {FormatType: picsTypeBool},
}

var UFSKeys = map[string]PicsKey{
	"hidecloudui":         {FormatType: picsTypeBool},
	"ignoreexternalfiles": {FormatType: picsTypeBool},
	"maxnumfiles":         {FormatType: picsTypeNumber},
	"quota":               {FormatType: picsTypeBytes},
	"rootoverrides":       {FormatType: picsTypeJSON},
	"savefiles":           {FormatType: picsTypeCustom},
}

type PicsKey struct {
	FormatType  int
	Link        string
	Description string
}

func getType(key string, keys map[string]PicsKey) int {

	if val, ok := keys[key]; ok {
		return val.FormatType
	}
	return 0
}

func getDescription(key string, keys map[string]PicsKey) string {

	if val, ok := keys[key]; ok {
		return val.Description
	}
	return ""
}

func FormatVal(key string, val string, appID int, keys map[string]PicsKey) interface{} {

	if item, ok := keys[key]; ok {
		switch item.FormatType {
		case picsTypeBool:

			b, _ := strconv.ParseBool(val)
			if b || val == "yes" {
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

			return template.HTML("<a href=\"" + item.Link + "\" rel=\"nofollow\" target=\"_blank\"><img class=\"wide\" data-lazy=\"" + item.Link + "\" alt=\"\" data-lazy-alt=\"" + key + "\" /></a>")

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

			// Shrink keys
			j = regexp.MustCompile("([A-Z0-9]{31,})").ReplaceAllStringFunc(j, func(s string) string {
				return s[0:30] + "..."
			})

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

			err := helpers.Unmarshal([]byte(val), &idMap)
			if err != nil {
				log.Err(err, val)
			}

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

		case picsTypeStringListJSON:

			m := map[string]string{}

			err := helpers.Unmarshal([]byte(val), &m)
			if err != nil {
				log.Err(err, val)
			}

			var items []string
			for k, v := range m {
				items = append(items, "<li>"+k+": <span class=font-weight-bold>"+v+"</span></li>")
			}

			return template.HTML("<ul class='mb-0 pl-3'>" + strings.Join(items, "") + "</ul>")

		case picsTypeNumberListJSONKeys:

			idMap := map[string]string{}

			err := helpers.Unmarshal([]byte(val), &idMap)
			if err != nil {
				log.Err(err, val)
			}

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

		case picsTypeTooLong:

			val = regexp.MustCompile("([a-zA-Z0-9]{31,})").ReplaceAllStringFunc(val, func(s string) string {
				return s[0:30] + "..."
			})

		case picsTypeMap:

			if val != "" {

				m := map[string]string{}
				err := helpers.Unmarshal([]byte(val), &m)
				if err != nil {
					log.Err(err, val)
				}

				var items []string
				for k, v := range m {
					items = append(items, "<li>"+k+": <span class=font-weight-bold>"+v+"</span></li>")
				}

				return template.HTML("<ul class='mb-0 pl-3'>" + strings.Join(items, "") + "</ul>")
			}

		case picsTypePercent:

			return val + "%"

		case picsTypeCustom:

			switch key {
			case "eulas":

				if val != "" {

					eulas := EULAs{}
					_ = helpers.Unmarshal([]byte(val), &eulas)

					var items []string
					for _, eula := range eulas {
						if eula.Name == "" {
							eula.Name = "EULA"
						}
						items = append(items, `<li><a target="_blank" href="`+eula.URL+`">`+eula.Name+`</a></li>`)
					}

					return template.HTML("<ul class='mb-0 pl-3'>" + strings.Join(items, "") + "</ul>")
				}

			case "supported_languages":

				if val != "" {

					langs := SupportedLanguages{}
					err := helpers.Unmarshal([]byte(val), &langs)
					if err != nil {
						log.Err(err, val)
					}

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
					err := helpers.Unmarshal([]byte(val), &categories)
					if err != nil {
						log.Err(err, val)
					}

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
					err := helpers.Unmarshal([]byte(val), &languages)
					if err != nil {
						log.Err(err, val)
					}

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
					err := helpers.Unmarshal([]byte(val), &associations)
					if err != nil {
						log.Err(err, val)
					}

					var items []string
					for _, v := range associations {
						items = append(items, "<li>"+strings.Title(v.Type)+": <span class=font-weight-bold>"+v.Name+"</span></li>")
					}

					return template.HTML("<ul class='mb-0 pl-3'>" + strings.Join(items, "") + "</ul>")
				}

			case "savefiles":

				if val != "" {

					files := saveFiles{}
					err := helpers.Unmarshal([]byte(val), &files)
					if err != nil {
						log.Err(err, val)
					}

					var items []string
					for _, file := range files {

						if file.Path == "{}" {
							file.Path = ""
						}

						pieces := []string{
							`<strong>Path:</strong> ` + string(file.Path),
							`<strong>Pattern:</strong> ` + file.Pattern,
							`<strong>Root:</strong> ` + file.Root,
						}

						if file.Recursive != "" {
							pieces = append(pieces, `<strong>Recursive:</strong> `+file.Recursive)
						}

						if len(file.Platforms) > 0 {
							var platforms []string
							for _, v := range file.Platforms {
								platforms = append(platforms, v)
							}
							pieces = append(pieces, `<strong>Platforms:</strong> `+strings.Join(platforms, ", "))
						}

						items = append(items, "<li>"+strings.Join(pieces, ", ")+"</li>")
					}

					return template.HTML("<ul class='mb-0 pl-3'>" + strings.Join(items, "") + "</ul>")
				}

			}

			return val
		}
	}

	return val
}
