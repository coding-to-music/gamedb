package helpers

import (
	"github.com/Jleagle/rabbit-go/rabbit"
	"github.com/gamedb/website/config"
)

var rabbitC = rabbit.Rabbit{
	Port:     config.Config.RabbitManagmentPort,
	Host:     config.Config.RabbitHost,
	Username: config.Config.RabbitUsername.Get(),
	Password: config.Config.RabbitPassword.Get(),
}

func GetRabbit() *rabbit.Rabbit {
	return &rabbitC
}
