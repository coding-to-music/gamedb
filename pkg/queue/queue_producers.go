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

	if len(payload.AppIDs) == 0 && len(payload.PackageIDs) == 0 && len(payload.ProfileIDs) == 0 {
		return nil
	}

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

	message := &steamMessage{
		Message: steamMessageInner{
			AppIDs:     appIDs,
			PackageIDs: packageIDs,
			PlayerIDs:  profileIDs,
		},
	}
	message.Force = force

	return produce(message, QueueSteam)
}

type AppPayload struct {
	ID           int
	ChangeNumber int
	VDF          map[string]interface{}
	Force        bool
}

func ProduceApp(payload AppPayload) (err error) {

	time.Sleep(time.Millisecond)

	if !helpers.IsValidAppID(payload.ID) {
		return sql.ErrInvalidAppID
	}

	message := &appMessage{
		Message: appMessageInner{
			ID:           payload.ID,
			ChangeNumber: payload.ChangeNumber,
			VDF:          payload.VDF,
		},
	}
	message.Force = payload.Force

	return produce(message, queueApps)
}

type PackagePayload struct {
	ID           int
	ChangeNumber int
	VDF          map[string]interface{}
	Force        bool
}

func ProducePackage(payload PackagePayload) (err error) {

	time.Sleep(time.Millisecond)

	if !sql.IsValidPackageID(payload.ID) {
		return sql.ErrInvalidPackageID
	}

	message := &packageMessage{
		Message: packageMessageInner{
			ID:           payload.ID,
			ChangeNumber: payload.ChangeNumber,
			VDF:          payload.VDF,
		},
	}
	message.Force = payload.Force

	return produce(message, queuePackages)
}

type PlayerPayload struct {
	ID         int64
	PBResponse *protobuf.CMsgClientFriendProfileInfoResponse
	Force      bool
}

func ProducePlayer(payload PlayerPayload) (err error) {

	time.Sleep(time.Millisecond)

	if payload.PBResponse == nil {
		payload.PBResponse = &protobuf.CMsgClientFriendProfileInfoResponse{}
	}

	if !helpers.IsValidPlayerID(payload.ID) {
		return errors.New("invalid player id: " + strconv.FormatInt(payload.ID, 10))
	}

	message := &playerMessage{
		Message: playerMessageInner{
			ID:            payload.ID,
			Eresult:       payload.PBResponse.GetEresult(),
			SteamidFriend: int64(payload.PBResponse.GetSteamidFriend()),
			TimeCreated:   payload.PBResponse.GetTimeCreated(),
			RealName:      payload.PBResponse.GetRealName(),
			CityName:      payload.PBResponse.GetCityName(),
			StateName:     payload.PBResponse.GetStateName(),
			CountryName:   payload.PBResponse.GetCountryName(),
			Headline:      payload.PBResponse.GetHeadline(),
			Summary:       payload.PBResponse.GetSummary(),
		},
	}
	message.Force = payload.Force

	return produce(message, queuePlayers2)
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

func ProduceTest(id int) (err error) {

	time.Sleep(time.Millisecond)

	return produce(&testMessage{
		Message: testMessageInner{
			ID: id,
		},
	}, queueTest)
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

	if len(ids) == 0 {
		return nil
	}

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

		message := &groupMessage{
			Message: groupMessageInner{
				IDs: chunk,
			},
		}
		message.Force = force

		err = produce(message, queueGroups)
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
