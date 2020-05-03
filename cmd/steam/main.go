package main

import (
	"io/ioutil"
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steamvdf"
	"github.com/Philipp15b/go-steam"
	"github.com/Philipp15b/go-steam/protocol"
	"github.com/Philipp15b/go-steam/protocol/protobuf"
	"github.com/Philipp15b/go-steam/protocol/steamlang"
	"github.com/gamedb/gamedb/pkg/config"
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

	config.Init(version, helpers.GetIP())
	log.Initialise([]log.LogName{log.LogNameSteam})

	var err error

	loginDetails := steam.LogOnDetails{}
	loginDetails.Username = config.Config.SteamUsername.Get()
	loginDetails.Password = config.Config.SteamPassword.Get()
	loginDetails.SentryFileHash, _ = ioutil.ReadFile(steamSentryFilename)
	loginDetails.ShouldRememberPassword = true
	loginDetails.AuthCode = ""

	err = steam.InitializeSteamDirectory()
	if err != nil {
		log.Err(err)
	}

	steamClient = steam.NewClient()
	steamClient.RegisterPacketHandler(packetHandler{})
	steamClient.Connect()

	queue.SetSteamClient(steamClient)

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
				queue.Init(queue.QueueSteamDefinitions)

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
				if err != nil {
					log.Err(err)
				}

			case steam.FatalErrorEvent:

				// Disconnects
				log.Info("Steam: Disconnected:", e.Error())
				steamLoggedOn = false
				go steamClient.Connect()

			case error:
				if e != nil {
					log.Err(e)
				}
			}
		}
	}()

	helpers.KeepAlive()
}

func checkForChanges() {
	for {
		if !config.IsLocal() || checkForChangesOnLocal {
			if steamClient.Connected() && steamLoggedOn {

				steamChangeLock.Lock()

				// Get last change number from file
				if steamChangeNumber == 0 {
					b, _ := ioutil.ReadFile(steamCurrentChangeFilename)
					if len(b) > 0 {
						ui, err := strconv.ParseUint(string(b), 10, 32)
						if err != nil {
							log.Err(err)
						} else {
							steamChangeNumber = uint32(ui)
						}
					}
				}

				var true = true
				steamClient.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientPICSChangesSinceRequest, &protobuf.CMsgClientPICSChangesSinceRequest{
					SendAppInfoChanges:     &true,
					SendPackageInfoChanges: &true,
					SinceChangeNumber:      &steamChangeNumber,
				}))
			}
		}

		time.Sleep(time.Second * 5)
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
		// log.Debug(packet.String())
	case steamlang.EMsg_ClientRequestFreeLicenseResponse:
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

			kv, err := steamvdf.ReadBytes(app.GetBuffer())
			if err != nil {
				log.Err(err, id)
			} else {
				m = kv.ToMapOuter()
			}

			err = queue.ProduceApp(queue.AppMessage{ID: id, ChangeNumber: int(app.GetChangeNumber()), VDF: m})
			if err != nil {
				log.Err(err, id)
			}
		}
	}

	unknownApps := body.GetUnknownAppids()
	if len(unknownApps) > 0 {
		for _, app := range unknownApps {

			var id = int(app)
			err := queue.ProduceApp(queue.AppMessage{ID: id})
			if err != nil {
				log.Err(err, id)
			}
		}
	}

	packages := body.GetPackages()
	if len(packages) > 0 {
		for _, pack := range packages {

			var m = map[string]interface{}{}
			var id = int(pack.GetPackageid())

			kv, err := steamvdf.ReadBytes(pack.GetBuffer())
			if err != nil {
				log.Err(err, id)
			} else {
				m = kv.ToMapOuter()
			}

			err = queue.ProducePackage(queue.PackageMessage{ID: int(pack.GetPackageid()), ChangeNumber: int(pack.GetChangeNumber()), VDF: m})
			if err != nil {
				log.Err(err, id)
			}
		}
	}

	unknownPackages := body.GetUnknownPackageids()
	if len(unknownPackages) > 0 {
		for _, pack := range unknownPackages {

			var id = int(pack)
			err := queue.ProducePackage(queue.PackageMessage{ID: id})
			if err != nil {
				log.Err(err, id)
			}
		}
	}
}

func (ph packetHandler) handleChangesSince(packet *protocol.Packet) {

	defer steamChangeLock.Unlock()

	body := protobuf.CMsgClientPICSChangesSinceResponse{}
	packet.ReadProtoMsg(&body)

	if body.GetCurrentChangeNumber() <= steamChangeNumber {
		return
	}

	var false = false

	var appMap = map[uint32]uint32{}
	var packageMap = map[uint32]uint32{}

	var apps []*protobuf.CMsgClientPICSProductInfoRequest_AppInfo
	var packages []*protobuf.CMsgClientPICSProductInfoRequest_PackageInfo

	var changes = strconv.FormatUint(uint64(body.GetSinceChangeNumber()), 10) + " (latest: " + strconv.FormatUint(uint64(body.GetCurrentChangeNumber()), 10) + ")"

	// Apps
	appChanges := body.GetAppChanges()
	if len(appChanges) > 0 {

		if config.IsLocal() {
			log.Info(strconv.Itoa(len(appChanges)) + " apps since change " + changes)
		}

		for _, appChange := range appChanges {

			appMap[appChange.GetChangeNumber()] = appChange.GetAppid()

			apps = append(apps, &protobuf.CMsgClientPICSProductInfoRequest_AppInfo{
				Appid:      appChange.Appid,
				OnlyPublic: &false,
			})
		}
	}

	// Packages
	packageChanges := body.GetPackageChanges()
	if len(packageChanges) > 0 {

		if config.IsLocal() {
			log.Info(strconv.Itoa(len(packageChanges)) + " pack since change " + changes)
		}

		for _, packageChange := range packageChanges {

			packageMap[packageChange.GetChangeNumber()] = packageChange.GetPackageid()

			packages = append(packages, &protobuf.CMsgClientPICSProductInfoRequest_PackageInfo{
				Packageid: packageChange.Packageid,
			})
		}
	}

	// Send off for app/package info
	steamClient.Write(protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientPICSProductInfoRequest, &protobuf.CMsgClientPICSProductInfoRequest{
		Apps:         apps,
		Packages:     packages,
		MetaDataOnly: &false,
	}))

	// Save change
	err := queue.ProduceChanges(queue.ChangesMessage{
		AppIDs:     appMap,
		PackageIDs: packageMap,
	})
	if err != nil {
		log.Err(err)
		return
	}

	// Update cached change number
	steamChangeNumber = body.GetCurrentChangeNumber()
	err = ioutil.WriteFile(steamCurrentChangeFilename, []byte(strconv.FormatUint(uint64(steamChangeNumber), 10)), 0644)
	if err != nil {
		log.Err(err)
	}
}

func (ph packetHandler) handleProfileInfo(packet *protocol.Packet) {

	body := protobuf.CMsgClientFriendProfileInfoResponse{}
	packet.ReadProtoMsg(&body)

	var id = int64(body.GetSteamidFriend())
	err := queue.ProducePlayer(queue.PlayerMessage{ID: id})
	if err != nil {
		log.Err(err, id)
	}
}
