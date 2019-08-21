package queue

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/Philipp15b/go-steam/protocol"
	"github.com/Philipp15b/go-steam/protocol/protobuf"
	"github.com/Philipp15b/go-steam/protocol/steamlang"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
)

func ProduceApps(ids []int) {

	// todo, chunk into 100s

	var apps []*protobuf.CMsgClientPICSProductInfoRequest_AppInfo
	for _, id := range ids {

		uid := uint32(id)

		apps = append(apps, &protobuf.CMsgClientPICSProductInfoRequest_AppInfo{
			Appid: &uid,
		})
	}

	false := false

	steamClient.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientPICSProductInfoRequest, &protobuf.CMsgClientPICSProductInfoRequest{
		Apps:         apps,
		MetaDataOnly: &false,
	}))
}

func produceApp(id int, changeNumber int, vdf map[string]interface{}) (err error) {

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

func ProducePackages(ids []int) {

	// todo, chunk into 100s

	var packages []*protobuf.CMsgClientPICSProductInfoRequest_PackageInfo
	for _, id := range ids {

		uid := uint32(id)

		packages = append(packages, &protobuf.CMsgClientPICSProductInfoRequest_PackageInfo{
			Packageid: &uid,
		})
	}

	false := false

	steamClient.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientPICSProductInfoRequest, &protobuf.CMsgClientPICSProductInfoRequest{
		Packages:     packages,
		MetaDataOnly: &false,
	}))
}

func producePackage(ID int, changeNumber int, vdf map[string]interface{}) (err error) {

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

func ProducePlayer(id int64) {

	ui := uint64(id)

	steamClient.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientFriendProfileInfo, &protobuf.CMsgClientFriendProfileInfo{
		SteamidFriend: &ui,
	}))
}

func producePlayer(ID int64, pb *protobuf.CMsgClientFriendProfileInfoResponse) (err error) {

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

func produceChange(apps map[int]int, packages map[int]int) (err error) {

	time.Sleep(time.Millisecond)

	return produce(baseMessage{
		Message: changeMessage{
			AppIDs:     apps,
			PackageIDs: packages,
		},
	}, queueGoChanges)
}
