package main

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
	"github.com/gamedb/gamedb/pkg/queue"
)

const (
	sentryFilename        = ".sentry.txt"
	currentChangeFilename = ".change.txt"
)

var (
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

	client := steam.NewClient()
	client.RegisterPacketHandler(packetHandler{})
	client.Connect()

	for event := range client.Events() {
		switch e := event.(type) {
		case *steam.ConnectedEvent:
			log.Info("Connected")
			client.Auth.LogOn(&logonDetails)
		case *steam.MachineAuthUpdateEvent:
			log.Info("Updating auth hash, it should no longer ask for auth")
			err = ioutil.WriteFile(sentryFilename, e.Hash, 0666)
			log.Err(err)
		case *steam.LoggedOnEvent:
			log.Info("Logged in")
			go checkForChanges(client)
		case *steam.LogOnFailedEvent:
			log.Info("Login failed")
			if e.Result == EResult_AccountLogonDenied {
				log.Info("Steam guard isn't letting me in! Enter auth code:")
				logonDetails.AuthCode = "xx"
			} else {
				log.Err(e.Result.String())
			}
		case *steam.DisconnectedEvent:
			log.Info("Disconnected")
		case steam.FatalErrorEvent:
			log.Critical(e.Error())
		}
	}
}

func checkForChanges(client *steam.Client) {
	for {
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

		log.Info("Trying from", changeNumber)

		changeLock.Lock()

		var x = true
		client.Write(protocol.NewClientMsgProtobuf(EMsg_ClientPICSChangesSinceRequest, &protobuf.CMsgClientPICSChangesSinceRequest{
			SendAppInfoChanges:     &x,
			SendPackageInfoChanges: &x,
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
		ph.changesSinceResponse(packet)
	}
}

func (ph packetHandler) changesSinceResponse(packet *protocol.Packet) {

	msg := protobuf.CMsgClientPICSChangesSinceResponse{}
	packet.ReadProtoMsg(&msg)

	appChanges := msg.GetAppChanges()
	if appChanges != nil && len(appChanges) > 0 {
		log.Info(len(appChanges), "apps")
		for _, v := range appChanges {
			appID := v.GetAppid()
			if appID > 0 {
				err := queue.ProduceApp(int(appID))
				log.Err(err)
			}
		}
	}

	packageChanges := msg.GetPackageChanges()
	if packageChanges != nil && len(packageChanges) > 0 {
		log.Info(len(packageChanges), "packages")
		for _, v := range packageChanges {
			packageID := v.GetPackageid()
			if packageID > 0 {
				err := queue.ProducePackage(int(packageID))
				log.Err(err)
			}
		}
	}

	// todo, changes

	// Update cached change number
	changeNumber = msg.GetCurrentChangeNumber()
	err := ioutil.WriteFile(currentChangeFilename, []byte(strconv.FormatUint(uint64(msg.GetCurrentChangeNumber()), 10)), 0644)
	log.Err(err)

	changeLock.Unlock()
}
