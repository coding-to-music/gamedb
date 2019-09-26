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

var ErrInQueue = errors.New("already in queue")

type SteamPayload struct {
	AppIDs     []int
	PackageIDs []int
	ProfileIDs []int64
}

func ProduceToSteam(payload SteamPayload, force bool) (err error) {

	time.Sleep(time.Millisecond)

	var appIDs []int
	var packageIDs []int
	var profileIDs []int64

	mc := helpers.GetMemcache()

	for _, appID := range payload.AppIDs {

		if !config.IsLocal() && !force {

			item := helpers.MemcacheAppInQueue(appID)

			_, err := mc.Get(item.Key)
			if err == nil {
				if len(payload.AppIDs) == 1 {
					return ErrInQueue
				}
				continue
			}

			err = mc.Set(&item)
			log.Err(err)
		}

		appIDs = append(appIDs, appID)
	}

	for _, packageID := range payload.PackageIDs {

		if !config.IsLocal() && !force {

			item := helpers.MemcachePackageInQueue(packageID)

			_, err := mc.Get(item.Key)
			if err == nil {
				if len(payload.PackageIDs) == 1 {
					return ErrInQueue
				}
				continue
			}

			err = mc.Set(&item)
			log.Err(err)
		}

		packageIDs = append(packageIDs, packageID)
	}

	for _, profileID := range payload.ProfileIDs {

		if !config.IsLocal() && !force {

			item := helpers.MemcachePlayerInQueue(profileID)

			_, err := mc.Get(item.Key)
			if err == nil {
				if len(payload.ProfileIDs) == 1 {
					return ErrInQueue
				}
				continue
			}

			err = mc.Set(&item)
			log.Err(err)
		}

		profileIDs = append(profileIDs, profileID)
	}

	return produce(&steamMessage{
		Message: steamMessageInner{
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

	if vdf == nil {
		vdf = map[string]interface{}{}
	}

	return produce(&appMessage{
		Message: appMessageInner{
			ID:           id,
			ChangeNumber: changeNumber,
			VDF:          vdf,
		},
	}, queueApps)
}

func ProducePackage(ID int, changeNumber int, vdf map[string]interface{}) (err error) {

	time.Sleep(time.Millisecond)

	if !sql.IsValidPackageID(ID) {
		return sql.ErrInvalidPackageID
	}

	if vdf == nil {
		vdf = map[string]interface{}{}
	}

	return produce(&packageMessage{
		Message: packageMessageInner{
			ID:           ID,
			ChangeNumber: changeNumber,
			VDF:          vdf,
		},
	}, queuePackages)
}

func ProducePlayer(ID int64, pb *protobuf.CMsgClientFriendProfileInfoResponse) (err error) {

	time.Sleep(time.Millisecond)

	if pb == nil {
		pb = &protobuf.CMsgClientFriendProfileInfoResponse{}
	}

	if !helpers.IsValidPlayerID(ID) {
		return errors.New("invalid player id: " + strconv.FormatInt(ID, 10))
	}

	return produce(&playerMessage{
		Message: playerMessageInner{
			ID:            ID,
			Eresult:       pb.GetEresult(),
			SteamidFriend: int64(pb.GetSteamidFriend()),
			TimeCreated:   pb.GetTimeCreated(),
			RealName:      pb.GetRealName(),
			CityName:      pb.GetCityName(),
			StateName:     pb.GetStateName(),
			CountryName:   pb.GetCountryName(),
			Headline:      pb.GetHeadline(),
			Summary:       pb.GetSummary(),
		},
	}, queuePlayers)
}

func ProduceChange(apps map[int]int, packages map[int]int) (err error) {

	time.Sleep(time.Millisecond)

	return produce(&changeMessage{
		Message: changeMessageInner{
			AppIDs:     apps,
			PackageIDs: packages,
		},
	}, queueChanges)
}

func ProduceBundle(ID int, appID int) (err error) {

	time.Sleep(time.Millisecond)

	return produce(&bundleMessage{
		Message: bundleMessageInner{
			ID:    ID,
			AppID: appID,
		},
	}, queueBundles)
}

func ProduceAppPlayers(IDs []int) (err error) {

	time.Sleep(time.Millisecond)

	if len(IDs) == 0 {
		return nil
	}

	return produce(&appPlayerMessage{
		Message: appPlayerMessageInner{
			IDs: IDs,
		},
	}, queueAppPlayer)
}

func ProduceGroup(ids []string, force bool) (err error) {

	time.Sleep(time.Millisecond)

	mc := helpers.GetMemcache()

	var filteredIDs []string

	for _, id := range ids {

		id = strings.TrimSpace(id)

		if helpers.IsValidGroupID(id) {

			if !config.IsLocal() && !force {

				item := helpers.MemcacheGroupInQueue(id)

				_, err := mc.Get(item.Key)
				if err == nil {
					continue
				}

				err = mc.Set(&item)
				log.Err(err)
			}

			filteredIDs = append(filteredIDs, id)
		}
	}

	if len(filteredIDs) == 0 {
		return nil
	}

	chunks := helpers.ChunkStrings(filteredIDs, 10)

	for _, chunk := range chunks {
		err = produce(&groupMessage{
			Message: groupMessageInner{
				IDs: chunk,
			},
		}, queueGroups)
		log.Err(err)
	}

	return nil
}

func produceGroupNew(id string) (err error) {

	time.Sleep(time.Millisecond)

	id = strings.TrimSpace(id)

	if !helpers.IsValidGroupID(id) {
		return nil
	}

	err = produce(&groupMessage{
		Message: groupMessageInner{
			IDs: []string{id},
		},
	}, queueGroupsNew)
	if err != nil {
		log.Err(err, id)
	}

	return nil
}
