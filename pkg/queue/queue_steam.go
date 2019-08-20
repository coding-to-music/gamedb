package queue

import (
	"io/ioutil"
	"strconv"
	"sync"
	"time"

	"github.com/Philipp15b/go-steam"
	"github.com/Philipp15b/go-steam/protocol"
	"github.com/Philipp15b/go-steam/protocol/protobuf"
	. "github.com/Philipp15b/go-steam/protocol/steamlang"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
)

const (
	sentryFilename        = ".sentry.txt"
	currentChangeFilename = ".change.txt"
)

var (
	steamClient *steam.Client

	changeNumber uint32
	changeLock   sync.Mutex
)

func main() {

	var err error

	logonDetails := steam.LogOnDetails{}
	logonDetails.Username = config.Config.SteamUsername.Get()
	logonDetails.Password = config.Config.SteamPassword.Get()
	logonDetails.SentryFileHash, _ = ioutil.ReadFile(sentryFilename)

	err = steam.InitializeSteamDirectory()
	log.Err(err)

	steamClient = steam.NewClient()
	steamClient.RegisterPacketHandler(packetHandler{})
	steamClient.Connect()

	for event := range steamClient.Events() {
		switch e := event.(type) {
		case *steam.ConnectedEvent:
			log.Info("Connected")
			steamClient.Auth.LogOn(&logonDetails)
		case *steam.LoggedOnEvent:
			log.Info("Logged in")
			go checkForChanges(steamClient)
		case *steam.LoggedOffEvent:
			steamClient.Disconnect()
		case *steam.DisconnectedEvent:
			log.Info("Disconnected")
			go steamClient.Connect()
		case *steam.LogOnFailedEvent:
			log.Info("Login failed")
			if e.Result == EResult_AccountLogonDenied {
				log.Info("Steam guard isn't letting me in! Enter auth code:")
				logonDetails.AuthCode = "xx"
			} else {
				log.Err(e.Result.String())
			}
		case *steam.MachineAuthUpdateEvent:
			log.Info("Updating auth hash, it should no longer ask for auth")
			err = ioutil.WriteFile(sentryFilename, e.Hash, 0666)
			log.Err(err)
		case steam.FatalErrorEvent:
			// Disconnects
			log.Err(e.Error())
		case error:
			log.Err(e)
		}
	}
}

func checkForChanges(client *steam.Client) {
	for {
		changeLock.Lock()

		// Get last change number from file
		if changeNumber == 0 {
			b, _ := ioutil.ReadFile(currentChangeFilename)
			if len(b) > 0 {
				ui, err := strconv.ParseUint(string(b), 10, 32)
				log.Err(err)
				if err == nil {
					changeNumber = uint32(ui)
					log.Err(err)
				}
			}
		}

		log.Info("Trying from: " + strconv.FormatUint(uint64(changeNumber), 10))

		var b = true
		client.Write(protocol.NewClientMsgProtobuf(EMsg_ClientPICSChangesSinceRequest, &protobuf.CMsgClientPICSChangesSinceRequest{
			SendAppInfoChanges:     &b,
			SendPackageInfoChanges: &b,
			SinceChangeNumber:      &changeNumber,
		}))

		time.Sleep(time.Second * 5)
	}
}

type packetHandler struct {
}

func (ph packetHandler) HandlePacket(packet *protocol.Packet) {

	switch packet.EMsg {
	case EMsg_ClientPICSChangesSinceResponse:
		ph.handleChanges(packet)
	case EMsg_ClientPICSProductInfoResponse:
		ph.handleProductInfo(packet)
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
			err := ProduceApp(int(app.GetAppid()), app.GetBuffer())
			log.Err(err)
		}
	}

	if packages != nil {
		for _, pack := range packages {
			err := ProducePackage(int(pack.GetPackageid()), pack.GetBuffer())
			log.Err(err)
		}
	}

	if unknownApps != nil {
		for _, app := range unknownApps {
			err := ProduceApp(int(app), nil)
			log.Err(err)
		}
	}

	if unknownPackages != nil {
		for _, pack := range unknownPackages {
			err := ProducePackage(int(pack), nil)
			log.Err(err)
		}
	}
}

func (ph packetHandler) handleChanges(packet *protocol.Packet) {

	defer changeLock.Unlock()

	var false = false

	var apps []*protobuf.CMsgClientPICSProductInfoRequest_AppInfo
	var packages []*protobuf.CMsgClientPICSProductInfoRequest_PackageInfo

	body := protobuf.CMsgClientPICSChangesSinceResponse{}
	packet.ReadProtoMsg(&body)

	appChanges := body.GetAppChanges()
	if appChanges != nil && len(appChanges) > 0 {
		log.Info(len(appChanges), "apps")
		for _, appChange := range appChanges {
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

	// todo, queue changes

	// Update cached change number
	changeNumber = body.GetCurrentChangeNumber()
	err := ioutil.WriteFile(currentChangeFilename, []byte(strconv.FormatUint(uint64(body.GetCurrentChangeNumber()), 10)), 0644)
	log.Err(err)
}
