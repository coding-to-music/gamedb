package pics

import (
	"html/template"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
)

type PicsItemType string

const (
	picsTypeBool   PicsItemType = "bool"
	picsTypeLink   PicsItemType = "link"
	picsTypeImage  PicsItemType = "image"
	picsTypeTime   PicsItemType = "timestamp"
	picsTypeJSON   PicsItemType = "json"
	picsTypeBytes  PicsItemType = "bytes"
	picsTypeNumber PicsItemType = "number"
	picsTypeApps   PicsItemType = "apps" // By comma
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
	"genres":                  {Type: picsTypeJSON},
	"has_adult_content":       {Type: picsTypeBool},
	"header_image":            {Type: picsTypeJSON},
	"icon":                    {Type: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.jpg"},
	"languages":               {Type: picsTypeJSON},
	"library_assets":          {Type: picsTypeJSON},
	"linuxclienticon":         {Type: picsTypeLink, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.zip"},
	"logo":                    {Type: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.jpg"},
	"logo_small":              {Type: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.jpg"},
	"metacritic_fullurl":      {Type: picsTypeLink, Link: "$val$"},
	"original_release_date":   {Type: picsTypeTime},
	"primary_genre":           {Type: picsTypeLink, Link: "/apps?tags=$val$"},
	"small_capsule":           {Type: picsTypeJSON},
	"steam_release_date":      {Type: picsTypeTime},
	"store_asset_mtime":       {Type: picsTypeTime},
	"store_tags":              {Type: picsTypeJSON},
	"supported_languages":     {Type: picsTypeJSON},
	"workshop_visible":        {Type: picsTypeBool},
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
	"listofdlc":                            {Type: picsTypeApps},
	"dlcavailableonstore":                  {Type: picsTypeBool},
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
	"hidecloudui": {Type: picsTypeBool},
	"maxnumfiles": {Type: picsTypeNumber},
	"quota":       {Type: picsTypeBytes},
	"savefiles":   {Type: picsTypeJSON},
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
				return "Yes"
			} else {
				return "No"
			}
		case picsTypeLink:
			item.Link = strings.ReplaceAll(item.Link, "$val$", val)
			item.Link = strings.ReplaceAll(item.Link, "$app$", strconv.Itoa(appID))

			var blank string
			if !strings.HasPrefix(item.Link, "/") {
				blank = " rel=\"nofollow\" target=\"_blank\""
			}

			return template.HTML("<a href=\"" + item.Link + "\"" + blank + ">" + val + "</a>")
		case picsTypeImage:
			item.Link = strings.ReplaceAll(item.Link, "$val$", val)
			item.Link = strings.ReplaceAll(item.Link, "$app$", strconv.Itoa(appID))
			return template.HTML("<img src=\"" + item.Link + "\" /><span><a href=\"" + item.Link + "\" rel=\"nofollow\" target=\"_blank\">" + val + "</a></span>")
		case picsTypeTime:

			i, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return val
			}

			t := time.Unix(i, 0).Format(helpers.DateTime) + " (" + val + ")"

			return t

		case picsTypeBytes:

			i, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				return val
			}

			return humanize.IBytes(i) + " (" + val + " bytes)"

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

		case picsTypeApps:

			var ret []string

			ids := strings.Split(val, ",")
			for _, id := range ids {
				ret = append(ret, `<a href="/apps/`+id+`" rel=\"nofollow\">`+id+`</a>`)
			}

			return template.HTML(strings.Join(ret, ", "))

		default:
			return val
		}
	}

	return val
}
