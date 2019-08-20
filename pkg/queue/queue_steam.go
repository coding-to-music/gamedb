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
	steamSentryFilename        = ".sentry.txt"
	steamCurrentChangeFilename = ".change.txt"
)

var (
	steamClient *steam.Client

	steamChangeNumber uint32
	steamChangeLock   sync.Mutex
	steamLoggedOn     bool
)

func init() {

	var err error

	logonDetails := steam.LogOnDetails{}
	logonDetails.Username = config.Config.SteamUsername.Get()
	logonDetails.Password = config.Config.SteamPassword.Get()
	logonDetails.SentryFileHash, _ = ioutil.ReadFile(steamSentryFilename)

	err = steam.InitializeSteamDirectory()
	steamLogError(err)

	steamClient = steam.NewClient()
	steamClient.RegisterPacketHandler(packetHandler{})
	steamClient.Connect()

	go func() {
		for event := range steamClient.Events() {
			switch e := event.(type) {
			case *steam.ConnectedEvent:
				steamLogInfo("Connected")
				go steamClient.Auth.LogOn(&logonDetails)
			case *steam.LoggedOnEvent:
				steamLogInfo("Logged in")
				steamLoggedOn = true
				go checkForChanges()
			case *steam.LoggedOffEvent:
				steamLogInfo("Logged off")
				steamLoggedOn = false
				go steamClient.Disconnect()
			case *steam.DisconnectedEvent:
				steamLogInfo("Disconnected")
				steamLoggedOn = false
				go steamClient.Connect()
			case *steam.LogOnFailedEvent:
				steamLogInfo("Login failed")
			case *steam.MachineAuthUpdateEvent:
				steamLogInfo("Updating auth hash, it should no longer ask for auth")
				err = ioutil.WriteFile(steamSentryFilename, e.Hash, 0666)
				steamLogError(err)
			case steam.FatalErrorEvent:
				// Disconnects
				steamLogError(e.Error())
			case error:
				steamLogError(e)
			}
		}
	}()
}

func checkForChanges() {
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
				steamLogError(err)
				if err == nil {
					steamChangeNumber = uint32(ui)
					steamLogError(err)
				}
			}
		}

		steamLogInfo("Trying from: " + strconv.FormatUint(uint64(steamChangeNumber), 10))

		var b = true
		steamClient.Write(protocol.NewClientMsgProtobuf(EMsg_ClientPICSChangesSinceRequest, &protobuf.CMsgClientPICSChangesSinceRequest{
			SendAppInfoChanges:     &b,
			SendPackageInfoChanges: &b,
			SinceChangeNumber:      &steamChangeNumber,
		}))

		time.Sleep(time.Second * 5)
	}
}

func getApp() {

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
		// steamLogInfo(packet.String())
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
			steamLogError(err)
		}
	}

	if packages != nil {
		for _, pack := range packages {
			err := ProducePackage(int(pack.GetPackageid()), pack.GetBuffer())
			steamLogError(err)
		}
	}

	if unknownApps != nil {
		for _, app := range unknownApps {
			err := ProduceApp(int(app), nil)
			steamLogError(err)
		}
	}

	if unknownPackages != nil {
		for _, pack := range unknownPackages {
			err := ProducePackage(int(pack), nil)
			steamLogError(err)
		}
	}
}

func (ph packetHandler) handleChanges(packet *protocol.Packet) {

	defer steamChangeLock.Unlock()

	var false = false

	var apps []*protobuf.CMsgClientPICSProductInfoRequest_AppInfo
	var packages []*protobuf.CMsgClientPICSProductInfoRequest_PackageInfo

	body := protobuf.CMsgClientPICSChangesSinceResponse{}
	packet.ReadProtoMsg(&body)

	appChanges := body.GetAppChanges()
	if appChanges != nil && len(appChanges) > 0 {
		steamLogInfo(len(appChanges), "apps")
		for _, appChange := range appChanges {
			apps = append(apps, &protobuf.CMsgClientPICSProductInfoRequest_AppInfo{
				Appid:      appChange.Appid,
				OnlyPublic: &false,
			})
		}
	}

	packageChanges := body.GetPackageChanges()
	if packageChanges != nil && len(packageChanges) > 0 {
		steamLogInfo(len(packageChanges), "packages")
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
	steamChangeNumber = body.GetCurrentChangeNumber()
	err := ioutil.WriteFile(steamCurrentChangeFilename, []byte(strconv.FormatUint(uint64(body.GetCurrentChangeNumber()), 10)), 0644)
	steamLogError(err)
}

func steamLogInfo(interfaces ...interface{}) {
	log.Info(append(interfaces, log.LogNamePICS)...)
}

func steamLogError(interfaces ...interface{}) {
	log.Err(append(interfaces, log.LogNamePICS)...)
}
