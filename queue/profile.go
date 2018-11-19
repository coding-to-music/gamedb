package queue

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/logging"
	"github.com/streadway/amqp"
)

type RabbitMessageProfile struct {
	ProfileInfo RabbitMessageProfilePICS `json:"ProfileInfo"`
}

func (d *RabbitMessageProfile) ToBytes() []byte {
	bytes, err := json.Marshal(d)
	logging.Error(err)
	return bytes
}

func (d RabbitMessageProfile) getQueueName() string {
	return QueueProfilesData
}

func (d RabbitMessageProfile) getRetryData() RabbitMessageDelay {
	return RabbitMessageDelay{}
}

func (d RabbitMessageProfile) process(msg amqp.Delivery) (ack bool, requeue bool, err error) {

	// Get message
	message := new(RabbitMessageProfile)

	err = helpers.Unmarshal(msg.Body, message)
	if err != nil {
		return false, false, err
	}

	var ID = message.ProfileInfo.SteamID.AccountID

	// Update player
	player, err := db.GetPlayer(ID)
	player.PlayerID = ID
	if err != nil {
		if err != db.ErrNoSuchEntity {
			logging.Error(err)
			return false, true, err
		}
	}

	r := new(http.Request)
	r.Header.Set("User-Agent", d.UserAgent)
	r.RemoteAddr = d.RemoteAddr

	err = player.Update(r, db.PlayerUpdateManual)
	if err != nil {
		logging.Error(err)
		return false, true, err
	}

	return true, false, nil
}

type RabbitMessageProfilePICS struct {
	Result  int `json:"Result"`
	SteamID struct {
		IsBlankAnonAccount            bool  `json:"IsBlankAnonAccount"`
		IsGameServerAccount           bool  `json:"IsGameServerAccount"`
		IsPersistentGameServerAccount bool  `json:"IsPersistentGameServerAccount"`
		IsAnonGameServerAccount       bool  `json:"IsAnonGameServerAccount"`
		IsContentServerAccount        bool  `json:"IsContentServerAccount"`
		IsClanAccount                 bool  `json:"IsClanAccount"`
		IsChatAccount                 bool  `json:"IsChatAccount"`
		IsLobby                       bool  `json:"IsLobby"`
		IsIndividualAccount           bool  `json:"IsIndividualAccount"`
		IsAnonAccount                 bool  `json:"IsAnonAccount"`
		IsAnonUserAccount             bool  `json:"IsAnonUserAccount"`
		IsConsoleUserAccount          bool  `json:"IsConsoleUserAccount"`
		IsValid                       bool  `json:"IsValid"`
		AccountID                     int64 `json:"AccountID"`
		AccountInstance               int   `json:"AccountInstance"`
		AccountType                   int   `json:"AccountType"`
		AccountUniverse               int   `json:"AccountUniverse"`
	} `json:"SteamID"`
	TimeCreated time.Time   `json:"TimeCreated"`
	RealName    string      `json:"RealName"`
	CityName    string      `json:"CityName"`
	StateName   string      `json:"StateName"`
	CountryName string      `json:"CountryName"`
	Headline    string      `json:"Headline"`
	Summary     string      `json:"Summary"`
	JobID       SteamKitJob `json:"JobID"`
}
