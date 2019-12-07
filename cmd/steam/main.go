package main

import (
	"io/ioutil"
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/valve-data-format-go/vdf"
	"github.com/Philipp15b/go-steam"
	"github.com/Philipp15b/go-steam/protocol"
	"github.com/Philipp15b/go-steam/protocol/protobuf"
	"github.com/Philipp15b/go-steam/protocol/steamlang"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
)

const (
	steamSentryFilename        = "sentry.txt"
	steamCurrentChangeFilename = "change.txt"
	checkForChangesOnLocal     = false
)

var (
	version string

	steamClient *steam.Client

	steamChangeNumber uint32
	steamChangeLock   sync.Mutex
	steamLoggedOn     bool
)

func main() {

	config.SetVersion(version)
	log.Initialise([]log.LogName{log.LogNameSteam})

	var err error

	loginDetails := steam.LogOnDetails{}
	loginDetails.Username = config.Config.SteamUsername.Get()
	loginDetails.Password = config.Config.SteamPassword.Get()
	loginDetails.SentryFileHash, _ = ioutil.ReadFile(steamSentryFilename)
	loginDetails.ShouldRememberPassword = true
	loginDetails.AuthCode = ""

	err = steam.InitializeSteamDirectory()
	log.Err(err)

	steamClient = steam.NewClient()
	steamClient.RegisterPacketHandler(packetHandler{})
	steamClient.Connect()

	consumers.SetSteamClient(steamClient)

	go func() {
		for event := range steamClient.Events() {

			switch e := event.(type) {
			case *steam.ConnectedEvent:

				log.Info("Steam: Connected")
				go steamClient.Auth.LogOn(&loginDetails)

			case *steam.LoggedOnEvent:

				// Load change checker
				log.Info("Steam: Logged in")
				steamLoggedOn = true
				go checkForChanges()

				// Load consumer
				log.Info("Starting Steam consumers")
				consumers.Init(consumers.QueueSteamDefinitions, true)

			case *steam.LoggedOffEvent:

				log.Info("Steam: Logged out")
				steamLoggedOn = false
				go steamClient.Disconnect()

			case *steam.DisconnectedEvent:

				log.Info("Steam: Disconnected")
				steamLoggedOn = false

				time.Sleep(time.Second * 5)

				go steamClient.Connect()

			case *steam.LogOnFailedEvent:

				// Disconnects
				log.Info("Steam: Login failed")

			case *steam.MachineAuthUpdateEvent:

				log.Info("Steam: Updating auth hash, it should no longer ask for auth")
				loginDetails.SentryFileHash = e.Hash
				err = ioutil.WriteFile(steamSentryFilename, e.Hash, 0666)
				log.Err(err)

			case steam.FatalErrorEvent:

				// Disconnects
				log.Info("Steam: Disconnected because of error")
				steamLoggedOn = false
				go steamClient.Connect()

			case error:
				log.Err(e)
			}
		}
	}()

	helpers.KeepAlive()
}

func checkForChanges() {
	if !config.IsLocal() || checkForChangesOnLocal {
		for {
			if !steamClient.Connected() || !steamLoggedOn {
				continue
			}

			steamChangeLock.Lock()

			// Get last change number from file
			if steamChangeNumber == 0 {
				b, _ := ioutil.ReadFile(steamCurrentChangeFilename)
				if len(b) > 0 {
					ui, err := strconv.ParseUint(string(b), 10, 32)
					log.Err(err)
					if err == nil {
						steamChangeNumber = uint32(ui)
						log.Err(err)
					}
				}
			}

			var t = true
			steamClient.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientPICSChangesSinceRequest, &protobuf.CMsgClientPICSChangesSinceRequest{
				SendAppInfoChanges:     &t,
				SendPackageInfoChanges: &t,
				SinceChangeNumber:      &steamChangeNumber,
			}))

			time.Sleep(time.Second * 5)
		}
	}
}

type packetHandler struct {
}

func (ph packetHandler) HandlePacket(packet *protocol.Packet) {

	switch packet.EMsg {
	case steamlang.EMsg_ClientPICSProductInfoResponse:
		ph.handleProductInfo(packet)
	case steamlang.EMsg_ClientPICSChangesSinceResponse:
		ph.handleChangesSince(packet)
	case steamlang.EMsg_ClientFriendProfileInfoResponse:
		ph.handleProfileInfo(packet)
	case steamlang.EMsg_ClientMarketingMessageUpdate2:
		log.Debug(packet.String())
	default:
		// log.Info(packet.String())
	}
}

func (ph packetHandler) handleProductInfo(packet *protocol.Packet) {

	body := protobuf.CMsgClientPICSProductInfoResponse{}
	packet.ReadProtoMsg(&body)

	apps := body.GetApps()
	if len(apps) > 0 {
		for _, app := range apps {

			var m = map[string]interface{}{}
			var id = int(app.GetAppid())

			kv, err := vdf.ReadBytes(app.GetBuffer())
			if err != nil {
				log.Err(err, id)
			} else {
				m = kv.ToMap()
			}

			err = queue.ProduceApp(queue.AppPayload{ID: id, ChangeNumber: int(app.GetChangeNumber()), VDF: m})
			if err != nil {
				log.Err(err, id)
			}
		}
	}

	unknownApps := body.GetUnknownAppids()
	if len(unknownApps) > 0 {
		for _, app := range unknownApps {

			var id = int(app)
			err := queue.ProduceApp(queue.AppPayload{ID: id})
			log.Err(err, id)
		}
	}

	packages := body.GetPackages()
	if len(packages) > 0 {
		for _, pack := range packages {

			var m = map[string]interface{}{}
			var id = int(pack.GetPackageid())

			kv, err := vdf.ReadBytes(pack.GetBuffer())
			if err != nil {
				log.Err(err, id)
			} else {
				m = kv.ToMap()
			}

			err = queue.ProducePackage(queue.PackagePayload{
				ID:           int(pack.GetPackageid()),
				ChangeNumber: int(pack.GetChangeNumber()),
				VDF:          m,
			})
			if err != nil {
				log.Err(err, id)
			}
		}
	}

	unknownPackages := body.GetUnknownPackageids()
	if len(unknownPackages) > 0 {
		for _, pack := range unknownPackages {

			var id = int(pack)
			err := queue.ProducePackage(queue.PackagePayload{ID: int(pack)})
			log.Err(err, id)
		}
	}
}

func (ph packetHandler) handleChangesSince(packet *protocol.Packet) {

	defer steamChangeLock.Unlock()

	var false = false

	var appMap = map[int]int{}
	var packageMap = map[int]int{}

	var apps []*protobuf.CMsgClientPICSProductInfoRequest_AppInfo
	var packages []*protobuf.CMsgClientPICSProductInfoRequest_PackageInfo

	body := protobuf.CMsgClientPICSChangesSinceResponse{}
	packet.ReadProtoMsg(&body)

	appChanges := body.GetAppChanges()
	if len(appChanges) > 0 {
		log.Info(strconv.Itoa(len(appChanges)) + " apps in change " + strconv.FormatUint(uint64(steamChangeNumber), 10))
		for _, appChange := range appChanges {

			appMap[int(appChange.GetChangeNumber())] = int(appChange.GetAppid())

			apps = append(apps, &protobuf.CMsgClientPICSProductInfoRequest_AppInfo{
				Appid:      appChange.Appid,
				OnlyPublic: &false,
			})
		}
	}

	packageChanges := body.GetPackageChanges()
	if len(packageChanges) > 0 {
		log.Info(strconv.Itoa(len(packageChanges)) + " packages in change " + strconv.FormatUint(uint64(steamChangeNumber), 10))
		for _, packageChange := range packageChanges {

			packageMap[int(packageChange.GetChangeNumber())] = int(packageChange.GetPackageid())

			packages = append(packages, &protobuf.CMsgClientPICSProductInfoRequest_PackageInfo{
				Packageid: packageChange.Packageid,
			})
		}
	}

	steamClient.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientPICSProductInfoRequest, &protobuf.CMsgClientPICSProductInfoRequest{
		Apps:         apps,
		Packages:     packages,
		MetaDataOnly: &false,
	}))

	err := queue.ProduceChange(appMap, packageMap)
	if err != nil {
		log.Err(err)
		return
	}

	// Update cached change number
	steamChangeNumber = body.GetCurrentChangeNumber()
	err = ioutil.WriteFile(steamCurrentChangeFilename, []byte(strconv.FormatUint(uint64(body.GetCurrentChangeNumber()), 10)), 0644)
	log.Err(err)
}

func (ph packetHandler) handleProfileInfo(packet *protocol.Packet) {

	body := protobuf.CMsgClientFriendProfileInfoResponse{}
	packet.ReadProtoMsg(&body)

	var id = int64(body.GetSteamidFriend())
	err := queue.ProducePlayer(queue.PlayerPayload{ID: id, PBResponse: &body})
	log.Err(err, id)
}
