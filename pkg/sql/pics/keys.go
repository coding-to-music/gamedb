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
	picsTypeJson   PicsItemType = "json"
	picsTypeBytes  PicsItemType = "bytes"
	picsTypeNumber PicsItemType = "number"
)

var CommonKeys = map[string]PicsKey{
	"associations":            {Type: picsTypeJson},
	"category":                {Type: picsTypeJson},
	"clienticns":              {Type: picsTypeLink, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.icns"},
	"clienticon":              {Type: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.ico"},
	"clienttga":               {Type: picsTypeLink, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.tga"},
	"community_hub_visible":   {Type: picsTypeBool},
	"community_visible_stats": {Type: picsTypeBool},
	"eulas":                   {Type: picsTypeJson},
	"exfgls":                  {Type: picsTypeBool},
	"gameid":                  {Type: picsTypeLink, Link: "/apps/$val$"},
	"genres":                  {Type: picsTypeJson},
	"has_adult_content":       {Type: picsTypeBool},
	"header_image":            {Type: picsTypeJson},
	"icon":                    {Type: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.jpg"},
	"languages":               {Type: picsTypeJson},
	"library_assets":          {Type: picsTypeJson},
	"linuxclienticon":         {Type: picsTypeLink, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.zip"},
	"logo":                    {Type: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.jpg"},
	"logo_small":              {Type: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.jpg"},
	"metacritic_fullurl":      {Type: picsTypeLink, Link: "$val$"},
	"original_release_date":   {Type: picsTypeTime},
	"primary_genre":           {Type: picsTypeLink, Link: "/apps?tags=$val$"},
	"small_capsule":           {Type: picsTypeJson},
	"steam_release_date":      {Type: picsTypeTime},
	"store_asset_mtime":       {Type: picsTypeTime},
	"store_tags":              {Type: picsTypeJson},
	"supported_languages":     {Type: picsTypeJson},
	"workshop_visible":        {Type: picsTypeBool},
}

var ExtendedKeys = map[string]PicsKey{
	"anti_cheat_support_url": {Type: picsTypeLink, Link: "$val$"},
	"developer_url":          {Type: picsTypeLink, Link: "$val$"},
	"gamemanualurl":          {Type: picsTypeLink, Link: "$val$"},
	"homepage":               {Type: picsTypeLink, Link: "$val$"},
	"isfreeapp":              {Type: picsTypeBool},
	"loadallbeforelaunch":    {Type: picsTypeBool},
	"noservers":              {Type: picsTypeBool},
	"requiressse":            {Type: picsTypeBool},
	"sourcegame":             {Type: picsTypeBool},
	"vacmacmodulecache":      {Type: picsTypeLink, Link: "/apps/$val$"},
	"vacmodulecache":         {Type: picsTypeLink, Link: "/apps/$val$"},
}

var ConfigKeys = map[string]PicsKey{
	"checkforupdatesbeforelaunch":  {Type: picsTypeBool},
	"signedfiles":                  {Type: picsTypeJson},
	"steamcontrollerconfigdetails": {Type: picsTypeJson},
	"steamcontrollertemplateindex": {Type: picsTypeBool},
	"systemprofile":                {Type: picsTypeBool},
	"verifyupdates":                {Type: picsTypeBool},
	"vrcompositorsupport":          {Type: picsTypeBool},
}

var UFSKeys = map[string]PicsKey{
	"hidecloudui": {Type: picsTypeBool},
	"maxnumfiles": {Type: picsTypeNumber},
	"quota":       {Type: picsTypeBytes},
}

type PicsKey struct {
	Type PicsItemType
	Link string
}

type KeyValue struct {
	Key   string
	Value interface{}
	Type  PicsItemType
}

func (kv KeyValue) TDClass() string {

	switch kv.Type {
	case picsTypeImage:
		return "img"
	case picsTypeJson:
		return "json"
	default:
		return ""
	}
}

func GetType(key string, keys map[string]PicsKey) PicsItemType {

	if val, ok := keys[key]; ok {
		return val.Type
	}
	return ""
}

func FormatVal(key string, val string, appID int, keys map[string]PicsKey) interface{} {

	if item, ok := keys[key]; ok {
		switch item.Type {
		case picsTypeBool:
			if val == "1" {
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

		case picsTypeJson:

			j, err := helpers.FormatJSON(val)
			if err != nil {
				return val
			}

			return j

		default:
			return val
		}
	}

	return val
}
