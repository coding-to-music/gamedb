package queue

import (
	"errors"

	"github.com/Jleagle/rabbit-go"
	"github.com/Philipp15b/go-steam"
	"github.com/Philipp15b/go-steam/protocol"
	"github.com/Philipp15b/go-steam/protocol/protobuf"
	"github.com/Philipp15b/go-steam/protocol/steamlang"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.uber.org/zap"
)

var steamClient *steam.Client

func SetSteamClient(c *steam.Client) {
	steamClient = c
}

type SteamMessage struct {
	AppIDs     []int `json:"app_ids"`
	PackageIDs []int `json:"package_ids"`
}

func steamHandler(message *rabbit.Message) {

	false := false

	if steamClient == nil || !steamClient.Connected() {
		log.ErrS(errors.New("steamClient not connected"))
		sendToRetryQueue(message)
		return
	}

	payload := SteamMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	// Apps
	if len(payload.AppIDs) > 0 {

		chunks := helpers.ChunkInts(payload.AppIDs, 100)
		for _, chunk := range chunks {

			var apps []*protobuf.CMsgClientPICSProductInfoRequest_AppInfo
			for _, id := range chunk {

				uid := uint32(id)
				apps = append(apps, &protobuf.CMsgClientPICSProductInfoRequest_AppInfo{
					Appid: &uid,
				})
			}

			steamClient.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientPICSProductInfoRequest, &protobuf.CMsgClientPICSProductInfoRequest{
				Apps:         apps,
				MetaDataOnly: &false,
			}))
		}
	}

	// Packages
	if len(payload.PackageIDs) > 0 {

		chunks := helpers.ChunkInts(payload.PackageIDs, 100)
		for _, chunk := range chunks {

			var packages []*protobuf.CMsgClientPICSProductInfoRequest_PackageInfo
			for _, id := range chunk {

				uid := uint32(id)
				packages = append(packages, &protobuf.CMsgClientPICSProductInfoRequest_PackageInfo{
					Packageid: &uid,
				})
			}

			steamClient.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientPICSProductInfoRequest, &protobuf.CMsgClientPICSProductInfoRequest{
				Packages:     packages,
				MetaDataOnly: &false,
			}))
		}
	}

	// Profiles
	// for _, id := range payload.PlayerIDs {
	//
	// 	ui := uint64(id)
	// 	steamClient.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientFriendProfileInfo, &protobuf.CMsgClientFriendProfileInfo{
	// 		SteamidFriend: &ui,
	// 	}))
	// }

	//
	message.Ack()
}
