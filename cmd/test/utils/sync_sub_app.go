package utils

import (
	"sort"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type syncSubApp struct{}

func (syncSubApp) name() string {
	return "test"
}

func (syncSubApp) run() {

	var i int
	err := mongo.BatchPackages(nil, nil, func(packages []mongo.Package) {
		for _, v := range packages {
			log.InfoS("Batch ", i)
			syncSubAppInner(&v)
		}
	})
	log.ErrS(err)
}

func syncSubAppInner(pack *mongo.Package) {

	if pack.HasEmptyName() || pack.HasEmptyIcon() || pack.ImageLogo == "" || pack.ImagePage == "" {

		projection := bson.M{
			"_id":                 1,
			"player_peak_alltime": 1,
			"name":                1,
			"icon":                1,
		}

		apps, err := mongo.GetAppsByID(pack.Apps, projection)
		if err != nil {
			log.ErrS(err)
			return
		}

		if len(apps) == 0 {
			return
		}

		sort.Slice(apps, func(i, j int) bool {
			return apps[i].PlayerPeakAllTime > apps[j].PlayerPeakAllTime
		})

		update := bson.D{}

		if pack.HasEmptyName() && apps[0].Name != "" {
			update = append(update, bson.E{Key: "name", Value: apps[0].GetName()})
		}

		if pack.HasEmptyIcon() {
			update = append(update, bson.E{Key: "icon", Value: apps[0].GetIcon()})
		}

		if pack.ImageLogo == "" {
			update = append(update, bson.E{Key: "image_logo", Value: apps[0].GetHeaderImage()})
		}

		if pack.ImagePage == "" {
			update = append(update, bson.E{Key: "image_page", Value: apps[0].GetHeaderImage()})
		}

		if len(update) > 0 {
			_, err = mongo.UpdateOne(mongo.CollectionPackages, bson.D{{"_id", pack.ID}}, update)
			if err != nil {
				log.ErrS(err)
			}
			log.InfoS(pack.ID)
		}
	}
}
