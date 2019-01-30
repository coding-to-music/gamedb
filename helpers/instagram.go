package helpers

import (
	"fmt"

	"github.com/gamedb/website/log"
	"github.com/robfig/cron"
)

const midnight = "* * 0 * * *"

func GetInstagram() {

}

func RunInstagram() {

	c := cron.New()
	err := c.AddFunc(midnight, func() {
		fmt.Println("Every hour on the half hour")
	})
	c.Start()

	if err != nil {
		log.Critical(err)
	}
}
