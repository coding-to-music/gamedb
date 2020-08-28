package rabbitweb

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
)

var rabbitWebClient *rabbit.Rabbit

func GetRabbitWebClient() *rabbit.Rabbit {

	if rabbitWebClient == nil {

		if config.C.RabbitHost == "" {
			log.Fatal("Missing environment variables")
		}

		rabbitWebClient = &rabbit.Rabbit{
			Host:     config.C.RabbitHost,
			Port:     config.C.RabbitManagmentPort,
			Username: config.C.RabbitUsername,
			Password: config.C.RabbitPassword,
		}
	}

	return rabbitWebClient
}
