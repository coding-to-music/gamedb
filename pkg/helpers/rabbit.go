package helpers

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/config"
)

var RabbitClient = rabbit.Rabbit{
	Host:     config.Config.RabbitHost.Get(),
	Port:     config.Config.RabbitManagmentPort.Get(),
	Username: config.Config.RabbitUsername.Get(),
	Password: config.Config.RabbitPassword.Get(),
}
