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
	baseMessage
	Message steamMessageInner `json:"message"`
}

type steamMessageInner struct {
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

	message := steamMessage{}
	message.OriginalQueue = QueueSteam

	if q.SteamClient == nil || !q.SteamClient.Connected() {
		logError(errors.New("steamClient not connected"))
		ackRetry(msg, &message)
		return
	}

	err = helpers.Unmarshal(msg.Body, &message)
	if err != nil {
		logError(err, msg.Body)
		message.ack(msg)
		return
	}

	// todo, chunk into 100s
	// Apps
	if len(message.Message.AppIDs) > 0 {

		var apps []*protobuf.CMsgClientPICSProductInfoRequest_AppInfo
		for _, id := range message.Message.AppIDs {

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
	if len(message.Message.PackageIDs) > 0 {

		var packages []*protobuf.CMsgClientPICSProductInfoRequest_PackageInfo
		for _, id := range message.Message.PackageIDs {

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
	for _, number := range message.Message.PlayerIDs {

		ui := uint64(number)

		q.SteamClient.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientFriendProfileInfo, &protobuf.CMsgClientFriendProfileInfo{
			SteamidFriend: &ui,
		}))
	}

	//
	message.ack(msg)
}
