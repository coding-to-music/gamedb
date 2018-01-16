package main

import (
	"context"
	"os"
	"strings"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/go-helpers/logger"
)

func createDsAppFromJsApp(js JsApp) *dsApp {

	jsTags := js.Common.StoreTags
	tags := make([]string, 0, len(jsTags))
	for _, value := range jsTags {
		tags = append(tags, value)
	}

	dsApp := dsApp{}
	dsApp.AppID = js.AppID
	dsApp.Name = js.Common.Name
	dsApp.Type = js.Common.Type
	dsApp.ReleaseState = js.Common.ReleaseState
	dsApp.OSList = strings.Split(js.Common.OSList, ",")
	dsApp.MetacriticScore = js.Common.MetacriticScore
	dsApp.MetacriticFullURL = js.Common.MetacriticURL
	dsApp.StoreTags = tags
	dsApp.Developer = js.Extended.Developer
	dsApp.Publisher = js.Extended.Publisher
	dsApp.Homepage = js.Extended.Homepage
	dsApp.ChangeNumber = js.ChangeNumber

	return &dsApp
}

func createDsPackageFromJsPackage(js JsPackage) *dsPackage {

	dsPackage := dsPackage{}

	return &dsPackage

}

func savePackage(data dsPackage) {

	key := datastore.NameKey(
		"Package",
		data.PackageID,
		nil,
	)

	saveKind(key, &data)
}

func saveKind(key *datastore.Key, data interface{}) (newKey *datastore.Key) {

	client, context := getDSClient()
	newKey, err := client.Put(context, key, data)
	if err != nil {
		logger.Error(err)
	}

	return newKey
}

func getDSClient() (*datastore.Client, context.Context) {

	context := context.Background()
	client, err := datastore.NewClient(
		context,
		os.Getenv("STEAM_GOOGLE_PROJECT"),
	)
	if err != nil {
		logger.Error(err)
	}

	return client, context
}

type dsChange struct {
	ChangeID int      `datastore:"change_id"`
	Apps     []string `datastore:"apps"`
	Packages []string `datastore:"packages"`
}

type dsApp struct {
	AppID             string   `datastore:"app_id"`
	Name              string   `datastore:"name"`
	Type              string   `datastore:"type"`
	ReleaseState      string   `datastore:"releasestate"`
	OSList            []string `datastore:"oslist"`
	MetacriticScore   string   `datastore:"metacritic_score"`
	MetacriticFullURL string   `datastore:"metacritic_fullurl"`
	StoreTags         []string `datastore:"store_tags"`
	Developer         string   `datastore:"developer"`
	Publisher         string   `datastore:"publisher"`
	Homepage          string   `datastore:"homepage"`
	ChangeNumber      int      `datastore:"change_number"`
}

type dsPackage struct {
	PackageID   string `datastore:"package_id"`
	BillingType int8   `datastore:"billingtype"`
	LicenseType int8   `datastore:"licensetype"`
	Status      int8   `datastore:"status"`
	Apps        []int  `datastore:"apps"`
}
