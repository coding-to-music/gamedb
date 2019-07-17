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
	picsTypeBool             PicsItemType = "bool"
	picsTypeLink             PicsItemType = "link"
	picsTypeImage            PicsItemType = "image"
	picsTypeTimestamp        PicsItemType = "timestamp"
	picsTypeJSON             PicsItemType = "json"
	picsTypeBytes            PicsItemType = "bytes"
	picsTypeNumber           PicsItemType = "number"
	picsTypeNumberListJSON   PicsItemType = "number-list-json"   // From comma string
	picsTypeNumberListString PicsItemType = "number-list-string" // From JSON object
	picsTypeTextListString   PicsItemType = "text-list-string"   // From comma string
	picsTypeTitle            PicsItemType = "title"
)

var CommonKeys = map[string]PicsKey{
	"associations":            {Type: picsTypeJSON},
	"category":                {Type: picsTypeJSON},
	"clienticns":              {Type: picsTypeLink, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.icns"},
	"clienticon":              {Type: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.ico"},
	"clienttga":               {Type: picsTypeLink, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.tga"},
	"community_hub_visible":   {Type: picsTypeBool},
	"community_visible_stats": {Type: picsTypeBool},
	"controllervr":            {Type: picsTypeJSON},
	"eulas":                   {Type: picsTypeJSON},
	"exfgls":                  {Type: picsTypeBool, Description: "Exclude from game library sharing"},
	"gameid":                  {Type: picsTypeLink, Link: "/apps/$val$"},
	"genres":                  {Type: picsTypeNumberListJSON, Link: "/apps?genres=$val$"},
	"has_adult_content":       {Type: picsTypeBool},
	"header_image":            {Type: picsTypeJSON},
	"icon":                    {Type: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.jpg"},
	"languages":               {Type: picsTypeJSON},
	"library_assets":          {Type: picsTypeJSON},
	"linuxclienticon":         {Type: picsTypeLink, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.zip"},
	"logo":                    {Type: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.jpg"},
	"logo_small":              {Type: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.jpg"},
	"metacritic_fullurl":      {Type: picsTypeLink, Link: "$val$"},
	"original_release_date":   {Type: picsTypeTimestamp},
	"primary_genre":           {Type: picsTypeLink, Link: "/apps?genres=$val$"},
	"small_capsule":           {Type: picsTypeJSON},
	"steam_release_date":      {Type: picsTypeTimestamp},
	"store_asset_mtime":       {Type: picsTypeTimestamp},
	"store_tags":              {Type: picsTypeNumberListJSON, Link: "/apps?tags=$val$"},
	"supported_languages":     {Type: picsTypeJSON},
	"workshop_visible":        {Type: picsTypeBool},
	"releasestate":            {Type: picsTypeTitle},
	"type":                    {Type: picsTypeTitle},
	"controller_support":      {Type: picsTypeTitle},
	"oslist":                  {Type: picsTypeTextListString, Link: "/apps?platforms=$val$"},
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
	"validoslist":                          {Type: picsTypeTextListString},
	"languages":                            {Type: picsTypeTextListString},
	"visibleonlywheninstalled":             {Type: picsTypeBool},
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

type KeyValue struct {
	Key            string
	Value          string
	ValueFormatted interface{}
	Type           PicsItemType
	Description    string
}

func (kv KeyValue) TDClass() string {

	switch kv.Type {
	case picsTypeImage:
		return "img"
	default:
		return ""
	}
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

			return template.HTML(time.Unix(i, 0).Format(helpers.DateTime) + " <small>(" + val + ")</small>")

		case picsTypeBytes:

			i, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				return val
			}

			return humanize.Bytes(i) + " (" + val + " bytes)"

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

		}
	}

	return val
}
