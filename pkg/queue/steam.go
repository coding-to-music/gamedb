package queue

import (
	"errors"

	"github.com/Philipp15b/go-steam"
	"github.com/Philipp15b/go-steam/protocol"
	"github.com/Philipp15b/go-steam/protocol/protobuf"
	"github.com/Philipp15b/go-steam/protocol/steamlang"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/streadway/amqp"
)

type steamMessage struct {
	AppIDs     []int   `json:"app_ids,omitempty"`
	PackageIDs []int   `json:"package_ids,omitempty"`
	PlayerIDs  []int64 `json:"player_ids,omitempty"`
}

type steamQueue struct {
	SteamClient *steam.Client
}

func (q steamQueue) processMessages(msgs []amqp.Delivery) {

	msg := msgs[0]

	var err error

	var payload = baseMessage{
		Message:       steamMessage{},
		OriginalQueue: QueueSteam,
	}

	if q.SteamClient == nil || !q.SteamClient.Connected() {
		logError(errors.New("steamClient not connected"))
		payload.ackRetry(msg)
		return
	}

	err = helpers.UnmarshalNumber(msg.Body, &payload)
	if err != nil {
		logError(err)
		payload.ack(msg)
		return
	}

	var message steamMessage
	err = helpers.MarshalUnmarshal(payload.Message, &message)
	if err != nil {
		logError(err)
		payload.ack(msg)
		return
	}

	// todo, chunk into 100s
	// Apps
	if len(message.AppIDs) > 0 {

		var apps []*protobuf.CMsgClientPICSProductInfoRequest_AppInfo
		for _, id := range message.AppIDs {

			uid := uint32(id)

			apps = append(apps, &protobuf.CMsgClientPICSProductInfoRequest_AppInfo{
				Appid: &uid,
			})
		}

		false := false

		q.SteamClient.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientPICSProductInfoRequest, &protobuf.CMsgClientPICSProductInfoRequest{
			Apps:         apps,
			MetaDataOnly: &false,
		}))
	}

	// todo, chunk into 100s
	// Packages
	if len(message.PackageIDs) > 0 {

		var packages []*protobuf.CMsgClientPICSProductInfoRequest_PackageInfo
		for _, id := range message.PackageIDs {

			uid := uint32(id)

			packages = append(packages, &protobuf.CMsgClientPICSProductInfoRequest_PackageInfo{
				Packageid: &uid,
			})
		}

		false := false

		q.SteamClient.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientPICSProductInfoRequest, &protobuf.CMsgClientPICSProductInfoRequest{
			Packages:     packages,
			MetaDataOnly: &false,
		}))
	}

	// Profiles
	for _, id := range message.PlayerIDs {

		ui := uint64(id)

		q.SteamClient.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientFriendProfileInfo, &protobuf.CMsgClientFriendProfileInfo{
			SteamidFriend: &ui,
		}))
	}

	//
	payload.ack(msg)
}
