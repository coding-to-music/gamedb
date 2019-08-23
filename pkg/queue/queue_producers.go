package queue

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/Philipp15b/go-steam/protocol/protobuf"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
)

type SteamPayload struct {
	AppIDs     []int
	PackageIDs []int
	ProfileIDs []int64
}

func ProduceToSteam(payload SteamPayload) (err error) {

	time.Sleep(time.Millisecond)

	var appIDs []int
	var packageIDs []int
	var profileIDs []int64

	mc := helpers.GetMemcache()

	for _, appID := range payload.AppIDs {

		if !config.IsLocal() {

			item := helpers.MemcacheAppInQueue(appID)

			_, err := mc.Get(item.Key)
			if err == nil {
				continue
			}

			err = mc.Set(&item)
			log.Err(err)
		}

		appIDs = append(appIDs, appID)
	}

	for _, packageID := range payload.PackageIDs {

		if !config.IsLocal() {

			item := helpers.MemcachePackageInQueue(packageID)

			_, err := mc.Get(item.Key)
			if err == nil {
				continue
			}

			err = mc.Set(&item)
			log.Err(err)
		}

		packageIDs = append(packageIDs, packageID)
	}

	for _, profileID := range payload.ProfileIDs {

		if !config.IsLocal() {

			item := helpers.MemcacheProfileInQueue(profileID)

			_, err := mc.Get(item.Key)
			if err == nil {
				continue
			}

			err = mc.Set(&item)
			log.Err(err)
		}

		profileIDs = append(profileIDs, profileID)
	}

	return produce(baseMessage{
		Message: steamMessage{
			AppIDs:     appIDs,
			PackageIDs: packageIDs,
			PlayerIDs:  profileIDs,
		},
	}, QueueSteam)
}

func ProduceApp(id int, changeNumber int, vdf map[string]interface{}) (err error) {

	time.Sleep(time.Millisecond)

	if !helpers.IsValidAppID(id) {
		return sql.ErrInvalidAppID
	}

	if !config.IsLocal() {

		mc := helpers.GetMemcache()

		item := helpers.MemcacheAppInQueue(id)

		_, err := mc.Get(item.Key)
		if err == nil {
			return nil
		}

		err = mc.Set(&item)
		log.Err(err)
	}

	if vdf == nil {
		vdf = map[string]interface{}{}
	}

	return produce(baseMessage{
		Message: appMessage{
			ID:           id,
			ChangeNumber: changeNumber,
			VDF:          vdf,
		},
	}, queueGoApps)
}

func ProducePackage(ID int, changeNumber int, vdf map[string]interface{}) (err error) {

	time.Sleep(time.Millisecond)

	if !sql.IsValidPackageID(ID) {
		return sql.ErrInvalidPackageID
	}

	if vdf == nil {
		vdf = map[string]interface{}{}
	}

	return produce(baseMessage{
		Message: packageMessage{
			ID:           ID,
			ChangeNumber: changeNumber,
			VDF:          vdf,
		},
	}, queueGoPackages)
}

func ProducePlayer(ID int64, pb *protobuf.CMsgClientFriendProfileInfoResponse) (err error) {

	time.Sleep(time.Millisecond)

	if pb == nil {
		pb = &protobuf.CMsgClientFriendProfileInfoResponse{}
	}

	if !helpers.IsValidPlayerID(ID) {
		return errors.New("invalid player id: " + strconv.FormatInt(ID, 10))
	}

	return produce(baseMessage{
		Message: playerMessage{
			ID:            ID,
			Eresult:       pb.GetEresult(),
			SteamidFriend: pb.GetSteamidFriend(),
			TimeCreated:   pb.GetTimeCreated(),
			RealName:      pb.GetRealName(),
			CityName:      pb.GetCityName(),
			StateName:     pb.GetStateName(),
			CountryName:   pb.GetCountryName(),
			Headline:      pb.GetHeadline(),
			Summary:       pb.GetSummary(),
		},
	}, queueGoPlayers)
}

func ProduceChange(apps map[int]int, packages map[int]int) (err error) {

	time.Sleep(time.Millisecond)

	return produce(baseMessage{
		Message: changeMessage{
			AppIDs:     apps,
			PackageIDs: packages,
		},
	}, queueGoChanges)
}

func ProduceBundle(ID int, appID int) (err error) {

	time.Sleep(time.Millisecond)

	return produce(baseMessage{
		Message: bundleMessage{
			ID:    ID,
			AppID: appID,
		},
	}, queueGoBundles)
}

func ProduceAppPlayers(IDs []int) (err error) {

	time.Sleep(time.Millisecond)

	if len(IDs) == 0 {
		return nil
	}

	return produce(baseMessage{
		Message: appPlayerMessage{
			IDs: IDs,
		},
	}, queueGoAppPlayer)
}

func ProduceGroup(IDs []string) (err error) {

	time.Sleep(time.Millisecond)

	mc := helpers.GetMemcache()

	var prodIDs []string

	for _, v := range IDs {

		v = strings.TrimSpace(v)

		if helpers.IsValidGroupID(v) {

			if config.IsProd() {

				item := helpers.MemcacheGroupInQueue(v)

				_, err := mc.Get(item.Key)
				if err == nil {
					continue
				}

				err = mc.Set(&item)
				log.Err(err)
			}

			prodIDs = append(prodIDs, v)
		}
	}

	if len(prodIDs) == 0 {
		return nil
	}

	chunks := helpers.ChunkStrings(prodIDs, 10)

	for _, chunk := range chunks {
		err = produce(baseMessage{
			Message: groupMessage{
				IDs: chunk,
			},
		}, queueGoGroups)
		log.Err(err)
	}

	return nil
}

func produceGroupNew(ID string) (err error) {

	time.Sleep(time.Millisecond)

	ID = strings.TrimSpace(ID)

	if !helpers.IsValidGroupID(ID) {
		return nil
	}

	err = produce(baseMessage{
		Message: groupMessage{
			ID: ID,
		},
	}, queueGoGroupsNew)
	if err != nil {
		log.Err(err, ID)
	}

	return nil
}
