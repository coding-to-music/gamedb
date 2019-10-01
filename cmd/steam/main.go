package main

import (
	"io/ioutil"
	"strconv"
	"sync"
	"time"

	vdf "github.com/Jleagle/valve-data-format-go"
	"github.com/Philipp15b/go-steam"
	"github.com/Philipp15b/go-steam/protocol"
	"github.com/Philipp15b/go-steam/protocol/protobuf"
	. "github.com/Philipp15b/go-steam/protocol/steamlang"
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
				q := queue.QueueRegister[queue.QueueSteam]
				q.SteamClient = steamClient
				go q.ConsumeMessages()

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

			log.Info("Trying from: " + strconv.FormatUint(uint64(steamChangeNumber), 10))

			var t = true
			steamClient.Write(protocol.NewClientMsgProtobuf(EMsg_ClientPICSChangesSinceRequest, &protobuf.CMsgClientPICSChangesSinceRequest{
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
	case EMsg_ClientPICSProductInfoResponse:
		ph.handleProductInfo(packet)
	case EMsg_ClientPICSChangesSinceResponse:
		ph.handleChangesSince(packet)
	case EMsg_ClientFriendProfileInfoResponse:
		ph.handleProfileInfo(packet)
	case EMsg_ClientMarketingMessageUpdate2:
		log.Debug(packet.String())
	default:
		// log.Info(packet.String())
	}
}

func (ph packetHandler) handleProductInfo(packet *protocol.Packet) {

	body := protobuf.CMsgClientPICSProductInfoResponse{}
	packet.ReadProtoMsg(&body)

	apps := body.GetApps()
	packages := body.GetPackages()
	unknownApps := body.GetUnknownAppids()
	unknownPackages := body.GetUnknownPackageids()

	if apps != nil {
		for _, app := range apps {

			m := map[string]interface{}{}

			kv, err := vdf.ReadBytes(app.GetBuffer())
			if err != nil {
				log.Err(err)
			} else {
				m = kv.ToMap()
			}

			err = queue.ProduceApp(int(app.GetAppid()), int(app.GetChangeNumber()), m)
			if err != nil {
				log.Err(err)
			}
		}
	}

	if packages != nil {
		for _, pack := range packages {

			m := map[string]interface{}{}

			kv, err := vdf.ReadBytes(pack.GetBuffer())
			if err != nil {
				log.Err(err)
			} else {
				m = kv.ToMap()
			}

			err = queue.ProducePackage(int(pack.GetPackageid()), int(pack.GetChangeNumber()), m)
			log.Err(err)
		}
	}

	if unknownApps != nil {
		for _, app := range unknownApps {
			err := queue.ProduceApp(int(app), 0, nil)
			log.Err(err)
		}
	}

	if unknownPackages != nil {
		for _, pack := range unknownPackages {
			err := queue.ProducePackage(int(pack), 0, nil)
			log.Err(err)
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
	if appChanges != nil && len(appChanges) > 0 {
		log.Info(len(appChanges), "apps")
		for _, appChange := range appChanges {

			appMap[int(appChange.GetChangeNumber())] = int(appChange.GetAppid())

			apps = append(apps, &protobuf.CMsgClientPICSProductInfoRequest_AppInfo{
				Appid:      appChange.Appid,
				OnlyPublic: &false,
			})
		}
	}

	packageChanges := body.GetPackageChanges()
	if packageChanges != nil && len(packageChanges) > 0 {
		log.Info(len(packageChanges), "packages")
		for _, packageChange := range packageChanges {

			packageMap[int(packageChange.GetChangeNumber())] = int(packageChange.GetPackageid())

			packages = append(packages, &protobuf.CMsgClientPICSProductInfoRequest_PackageInfo{
				Packageid: packageChange.Packageid,
			})
		}
	}

	steamClient.Write(protocol.NewClientMsgProtobuf(EMsg_ClientPICSProductInfoRequest, &protobuf.CMsgClientPICSProductInfoRequest{
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

	err := queue.ProducePlayer(int64(body.GetSteamidFriend()), &body)
	log.Err(err)
}
