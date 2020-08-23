package rabbitweb

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/config"
)

var RabbitClient = rabbit.Rabbit{
	Host:     config.C.RabbitHost,
	Port:     config.C.RabbitManagmentPort,
	Username: config.C.RabbitUsername,
	Password: config.C.RabbitPassword,
}
