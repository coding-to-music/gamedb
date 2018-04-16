package steam

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/steam-authority/steam-authority/logger"
)

var (
	ErrPICSNoResponse = errors.New("empty response from pics")
)

func GetPICSInfo(apps []int, packages []int) (jsInfo JsInfo, err error) {

	var stringApps []string
	var stringPackages []string

	for _, vv := range apps {
		stringApps = append(stringApps, strconv.Itoa(vv))
	}
	for _, vv := range packages {
		stringPackages = append(stringPackages, strconv.Itoa(vv))
	}

	// Grab the JSON from node
	url := "http://localhost:8086/info?apps=" + strings.Join(stringApps, ",") + "&packages=" + strings.Join(stringPackages, ",") + "&prettyprint=0"
	//logger.Info("PICS: " + url)
	response, err := http.Get(url)
	if err != nil {
		return jsInfo, err
	}
	defer response.Body.Close()

	// Convert to bytes
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return jsInfo, err
	}

	if len(bytes) == 0 {
		return jsInfo, ErrPICSNoResponse
	}

	// Fix arrays that should be objects
	var b = string(bytes)
	b = strings.Replace(b, "\"appitems\":[]", "\"appitems\":null", 1)
	b = strings.Replace(b, "\"extended\":[]", "\"extended\":null", 1)
	bytes = []byte(b)

	// Unmarshal JSON
	info := JsInfo{}
	if err := json.Unmarshal(bytes, &info); err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(bytes))
		}
		return jsInfo, err
	}

	return info, nil
}

type JsInfo struct {
	Success         int8                 `json:"success"`
	Apps            map[string]JsApp     `json:"apps"`
	Packages        map[string]JsPackage `json:"packages"`
	UnknownApps     []int                `json:"unknown_apps"`
	UnknownPackages []int                `json:"unknown_packages"`
}

type JsApp struct {
	AppID      string      `json:"appid"`
	PublicOnly string      `json:"public_only"`
	Common     JsAppCommon `json:"common"`
	Extended struct {
		Developer           string `json:"developer"`
		DeveloperURL        string `json:"developer_url"`
		GameDir             string `json:"gamedir"`
		GameManualURL       string `json:"gamemanualurl"`
		Homepage            string `json:"homepage"`
		Icon                string `json:"icon"`
		Icon2               string `json:"icon2"`
		IsFreeApp           string `json:"isfreeapp"`
		Languages           string `json:"languages"`
		LoadAllBeforeLaunch string `json:"loadallbeforelaunch"`
		MinClientVersion    string `json:"minclientversion"`
		NoServers           string `json:"noservers"`
		PrimaryCache        string `json:"primarycache"`
		PrimaryCacheLinux   string `json:"primarycache_linux"`
		RequireSSSE         string `json:"requiressse"`
		ServerBrowserName   string `json:"serverbrowsername"`
		SourceGame          string `json:"sourcegame"`
		State               string `json:"state"`
		VacMacModuleCache   string `json:"vacmacmodulecache"`
		VacModuleCache      string `json:"vacmodulecache"`
		VacModuleFilename   string `json:"vacmodulefilename"`
		ValidosList         string `json:"validoslist"`
		Publisher           string `json:"publisher"`
		ListofDLC           string `json:"listofdlc"`
	} `json:"extended"`
	Config struct {
	} `json:"config"`
	Depots struct {
	} `json:"depots"`
	UFS struct {
		Quota       string `json:"quota"`
		MaxNumFiles string `json:"maxnumfiles"`
		HideCloudUI string `json:"hidecloudui"`
	} `json:"ufs"`
	SystemRequirements struct {
	} `json:"sysreqs"`
	ChangeNumber int `json:"change_number"`
}

type JsAppCommon struct {
	ClientICNS            string `json:"clienticns"`
	ClientIcon            string `json:"clienticon"`
	ClientTGA             string `json:"clienttga"`
	CommunityHubVisible   string `json:"community_hub_visible"`
	CommunityVisibleStats string `json:"community_visible_stats"`
	ControllerSupport     string `json:"controller_support"`
	EULAs map[string]struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"eulas"`
	Exfgls            string            `json:"exfgls"`
	GameID            string            `json:"gameid"`
	HeaderImage       map[string]string `json:"header_image"`
	Icon              string            `json:"icon"`
	Languages         map[string]string `json:"languages"`
	LinuxClientIcon   string            `json:"linuxclienticon"`
	Logo              string            `json:"logo"`
	LogoSmall         string            `json:"logo_small"`
	MetacriticFullurl string            `json:"metacritic_fullurl"`
	MetacriticName    string            `json:"metacritic_name"`
	MetacriticScore   string            `json:"metacritic_score"`
	MetacriticURL     string            `json:"metacritic_url"`
	Name              string            `json:"name"`
	OSList            string            `json:"oslist"`
	OSArch            string            `json:"osarch"`
	ReleaseState      string            `json:"releasestate"`
	SmallCapsule      map[string]string `json:"small_capsule"`
	StoreTags         map[string]string `json:"store_tags"`
	Type              string            `json:"type"`
	WorkshopVisible   string            `json:"workshop_visible"`
}

// JsPackage ...
type JsPackage struct {
	PackageID   int                    `json:"packageid"`
	BillingType int8                   `json:"billingtype"`
	LicenseType int8                   `json:"licensetype"`
	Status      int8                   `json:"status"`
	Extended    map[string]interface{} `json:"extended"`
	AppIDs      []int                  `json:"appids"`
	DepotIDs    []int                  `json:"depotids"`
	AppItems    map[string][]int       `json:"appitems"`
}
