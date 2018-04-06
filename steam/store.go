package steam

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/steam-authority/steam-authority/logger"
)

// todo
// http://store.steampowered.com/api/dlcforapp?appid=440

func GetAppDetailsFromStore(id int) (app AppDetailsBody, err error) {

	idx := strconv.Itoa(id)

	query := url.Values{}
	query.Set("appids", idx)
	query.Set("cc", "us") // Currency
	query.Set("l", "en")  // Language

	response, err := http.Get("http://store.steampowered.com/api/appdetails?" + query.Encode())
	if err != nil {
		return app, err
	}
	defer response.Body.Close()

	// Convert to bytes
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return app, err
	}

	// Check for no app
	if string(contents) == "null" {
		return app, errors.New("invalid app id: " + strconv.Itoa(id))
	}

	// Fix values that can change type, causing unmarshal errors
	var regex *regexp.Regexp
	var b = string(contents)

	// Convert strings to ints
	regex = regexp.MustCompile(`:\s?"(\d+)"`) // After colon
	b = regex.ReplaceAllString(b, `:$1`)

	regex = regexp.MustCompile(`,\s?"(\d+)"`) // After comma
	b = regex.ReplaceAllString(b, `,$1`)

	regex = regexp.MustCompile(`"(\d+)",`) // Before comma
	b = regex.ReplaceAllString(b, `$1,`)

	regex = regexp.MustCompile(`"packages":\s?\["(\d+)"\]`) // Package array with single int
	b = regex.ReplaceAllString(b, `"packages":[$1]`)

	// Make some its strings again
	regex = regexp.MustCompile(`"date":\s?(\d+)`)
	b = regex.ReplaceAllString(b, `"date":"$1"`)

	regex = regexp.MustCompile(`"name":\s?(\d+)`)
	b = regex.ReplaceAllString(b, `"name":"$1"`)

	regex = regexp.MustCompile(`"description":\s?(\d+)`)
	b = regex.ReplaceAllString(b, `"description":"$1"`)

	regex = regexp.MustCompile(`"display_type":\s?(\d+)`)
	b = regex.ReplaceAllString(b, `"display_type":"$1"`)

	regex = regexp.MustCompile(`"legal_notice":\s?(\d+)`)
	b = regex.ReplaceAllString(b, `"legal_notice":"$1"`)

	// Fix arrays that should be objects
	// todo, update to regex to use \s?
	b = strings.Replace(b, "\"pc_requirements\":[]", "\"pc_requirements\":null", 1)
	b = strings.Replace(b, "\"mac_requirements\":[]", "\"mac_requirements\":null", 1)
	b = strings.Replace(b, "\"linux_requirements\":[]", "\"linux_requirements\":null", 1)
	contents = []byte(b)

	// Unmarshal JSON
	resp := make(map[string]AppDetailsBody)
	if err := json.Unmarshal(contents, &resp); err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(contents))
		}
		return app, err
	}

	if resp[idx].Success == false {
		return app, errors.New("no app with id in steam")
	}

	return resp[idx], nil
}

type AppDetailsBody struct {
	Success bool `json:"success"`
	Data struct {
		Type                string `json:"type"`
		Name                string `json:"name"`
		SteamAppID          int    `json:"steam_appid"`
		RequiredAge         int    `json:"required_age"`
		IsFree              bool   `json:"is_free"`
		DLC                 []int  `json:"dlc"`
		ControllerSupport   string `json:"controller_support"`
		DetailedDescription string `json:"detailed_description"`
		AboutTheGame        string `json:"about_the_game"`
		ShortDescription    string `json:"short_description"`
		Fullgame struct {
			AppID int    `json:"appid"`
			Name  string `json:"name"`
		} `json:"fullgame"`
		SupportedLanguages string `json:"supported_languages"`
		Reviews            string `json:"reviews"`
		HeaderImage        string `json:"header_image"`
		Website            string `json:"website"`
		PcRequirements struct {
			Minimum     string `json:"minimum"`
			Recommended string `json:"recommended"`
		} `json:"pc_requirements"`
		MacRequirements struct {
			Minimum     string `json:"minimum"`
			Recommended string `json:"recommended"`
		} `json:"mac_requirements"`
		LinuxRequirements struct {
			Minimum     string `json:"minimum"`
			Recommended string `json:"recommended"`
		} `json:"linux_requirements"`
		LegalNotice string   `json:"legal_notice"`
		Developers  []string `json:"developers"`
		Publishers  []string `json:"publishers"`
		Demos []struct {
			Appid       int    `json:"appid"`
			Description string `json:"description"`
		} `json:"demos"`
		PriceOverview struct {
			Currency        string `json:"currency"`
			Initial         int    `json:"initial"`
			Final           int    `json:"final"`
			DiscountPercent int    `json:"discount_percent"`
		} `json:"price_overview"`
		Packages []int `json:"packages"`
		PackageGroups []struct {
			Name                    string `json:"name"`
			Title                   string `json:"title"`
			Description             string `json:"description"`
			SelectionText           string `json:"selection_text"`
			SaveText                string `json:"save_text"`
			DisplayType             string `json:"display_type"` // Can be string or int
			IsRecurringSubscription string `json:"is_recurring_subscription"`
			Subs []struct {
				Packageid                int    `json:"packageid"`
				PercentSavingsText       string `json:"percent_savings_text"`
				PercentSavings           int    `json:"percent_savings"`
				OptionText               string `json:"option_text"`
				OptionDescription        string `json:"option_description"`
				CanGetFreeLicense        int    `json:"can_get_free_license"`
				IsFreeLicense            bool   `json:"is_free_license"`
				PriceInCentsWithDiscount int    `json:"price_in_cents_with_discount"`
			} `json:"subs"`
		} `json:"package_groups"`
		Platforms struct {
			Windows bool `json:"windows"`
			Mac     bool `json:"mac"`
			Linux   bool `json:"linux"`
		} `json:"platforms"`
		Metacritic struct {
			Score int8   `json:"score"`
			URL   string `json:"url"`
		} `json:"metacritic"`
		Categories  []AppDetailsCategory   `json:"categories"`
		Genres      []AppDetailsGenre      `json:"genres"`
		Screenshots []AppDetailsScreenshot `json:"screenshots"`
		Movies []struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			Thumbnail string `json:"thumbnail"`
			Webm struct {
				Num480 string `json:"480"`
				Max    string `json:"max"`
			} `json:"webm"`
			Highlight bool `json:"highlight"`
		} `json:"movies"`
		Recommendations struct {
			Total int `json:"total"`
		} `json:"recommendations"`
		Achievements AppDetailsAchievements `json:"achievements"`
		ReleaseDate struct {
			ComingSoon bool   `json:"coming_soon"`
			Date       string `json:"date"`
		} `json:"release_date"`
		SupportInfo struct {
			URL   string `json:"url"`
			Email string `json:"email"`
		} `json:"support_info"`
		Background string `json:"background"`
	} `json:"data"`
}

type AppDetailsScreenshot struct {
	ID            int    `json:"id"`
	PathThumbnail string `json:"path_thumbnail"`
	PathFull      string `json:"path_full"`
}

type AppDetailsAchievements struct {
	Total int `json:"total"`
	Highlighted []struct {
		Name string `json:"name"`
		Path string `json:"path"`
	} `json:"highlighted"`
}

type AppDetailsGenre struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}

type AppDetailsCategory struct {
	ID          int8   `json:"id"`
	Description string `json:"description"`
}

func GetPackageDetailsFromStore(id int) (pack PackageDetailsBody, err error) {

	idx := strconv.Itoa(id)

	query := url.Values{}
	query.Set("packageids", idx)
	query.Set("cc", "us") // Currency
	query.Set("l", "en")  // Language

	response, err := http.Get("http://store.steampowered.com/api/packagedetails?" + query.Encode())
	if err != nil {
		return pack, err
	}
	defer response.Body.Close()

	// Convert to bytes
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return pack, err
	}

	// Check for no pack
	if string(contents) == "null" {
		return pack, errors.New("invalid package id: " + strconv.Itoa(id))
	}

	// Unmarshal JSON
	resp := make(map[string]PackageDetailsBody)
	if err := json.Unmarshal(contents, &resp); err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(contents))
		}
		return pack, err
	}

	if resp[idx].Success == false {
		return pack, errors.New("no package with id in steam")
	}

	return resp[idx], nil
}

type PackageDetailsBody struct {
	Success bool `json:"success"`
	Data struct {
		Name        string `json:"name"`
		PageImage   string `json:"page_image"`
		HeaderImage string `json:"header_image"`
		SmallLogo   string `json:"small_logo"`
		Apps []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"apps"`
		Price struct {
			Currency        string `json:"currency"`
			Initial         int    `json:"initial"`
			Final           int    `json:"final"`
			DiscountPercent int    `json:"discount_percent"`
			Individual      int    `json:"individual"`
		} `json:"price"`
		Platforms struct {
			Windows bool `json:"windows"`
			Mac     bool `json:"mac"`
			Linux   bool `json:"linux"`
		} `json:"platforms"`
		Controller struct {
			FullGamepad bool `json:"full_gamepad"`
		} `json:"controller"`
		ReleaseDate struct {
			ComingSoon bool   `json:"coming_soon"`
			Date       string `json:"date"`
		} `json:"release_date"`
	} `json:"data"`
}

func GetTags() (tags []steamTag, err error) {

	// Get tags names
	response, err := http.Get("http://store.steampowered.com/tagdata/populartags/english")
	if err != nil {
		logger.Error(err)
		return tags, err
	}
	defer response.Body.Close()

	// Convert to bytes
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logger.Error(err)
		return tags, err
	}

	// Unmarshal JSON
	if err := json.Unmarshal(contents, &tags); err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(contents))
		} else {
			logger.Error(err)
		}
		return tags, err
	}

	return tags, nil
}

type steamTag struct {
	TagID int    `json:"tagid"`
	Name  string `json:"name"`
}
