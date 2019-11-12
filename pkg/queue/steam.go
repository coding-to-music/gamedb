package queue

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/Philipp15b/go-steam"
	"github.com/Philipp15b/go-steam/protocol"
	"github.com/Philipp15b/go-steam/protocol/protobuf"
	"github.com/Philipp15b/go-steam/protocol/steamlang"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type steamMessage struct {
	baseMessage
	Message steamMessageInner `json:"message"`
}

type steamMessageInner struct {
	AppIDs     []int   `json:"app_ids"`
	PackageIDs []int   `json:"package_ids"`
	PlayerIDs  []int64 `json:"player_ids"`
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
		log.Err(errors.New("steamClient not connected"))
		ackRetry(msg, &message)
		return
	}

	err = helpers.Unmarshal(msg.Body, &message)
	if err != nil {
		log.Err(err, msg.Body)
		message.ack(msg)
		return
	}

	// todo, chunk into 100s
	// Apps
	if len(message.Message.AppIDs) > 0 {

		var apps []*protobuf.CMsgClientPICSProductInfoRequest_AppInfo
		for _, id := range message.Message.AppIDs {

			if message.Force {
				IDsToForce.Write("app-" + strconv.Itoa(id))
			}

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

			if message.Force {
				IDsToForce.Write("package-" + strconv.Itoa(id))
			}

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
	for _, id := range message.Message.PlayerIDs {

		if message.Force {
			IDsToForce.Write("player-" + strconv.FormatInt(id, 10))
		}

		ui := uint64(id)

		q.SteamClient.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientFriendProfileInfo, &protobuf.CMsgClientFriendProfileInfo{
			SteamidFriend: &ui,
		}))
	}

	//
	message.ack(msg)
}

var IDsToForce = IDsToForceType{
	IDs: map[string]time.Time{},
}

type IDsToForceType struct {
	IDs map[string]time.Time
	sync.Mutex
}

func (ids *IDsToForceType) Read(key string) (force bool) {

	ids.Lock()
	defer ids.Unlock()

	_, force = ids.IDs[key]
	if force {
		delete(ids.IDs, key)
	}
	return force
}

func (ids *IDsToForceType) Write(key string) {

	ids.Lock()
	defer ids.Unlock()

	ids.IDs[key] = time.Now()
}

func (ids *IDsToForceType) Cleanup() {

	ids.Lock()
	defer ids.Unlock()

	for k, v := range ids.IDs {
		if v.Unix() < time.Now().Add(-time.Minute).Unix() {
			delete(ids.IDs, k)
		}
	}
}
