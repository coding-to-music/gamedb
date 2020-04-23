package hosts

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/Jleagle/go-durationfmt"
	"github.com/gamedb/gamedb/pkg/helpers"
)

const (
	selectedService = hostHetzner

	hostDO      = "do"
	hostVU      = "vultr"
	hostHetzner = "hetzner"

	ConsumerTag = "scaler"
)

type Host interface {
	ListConsumers() ([]Consumer, error)
	CreateConsumer() (Consumer, error)
	DeleteConsumer(int) error
}

func GetHost() Host {
	if selectedService == hostDO {
		return DigitalOcean{}
	}
	if selectedService == hostHetzner {
		return Hetzner{}
	}
	return nil
}

type Consumer struct {
	ID        int
	Name      string
	IP        string
	Tags      []string
	CreatedAt int64
	Locked    bool
}

func (c Consumer) GetTags() string {
	sort.Strings(c.Tags)
	return strings.Join(c.Tags, ", ")
}

func (c Consumer) CanDelete() bool {
	return helpers.SliceHasString(ConsumerTag, c.Tags) && !c.Locked
}

func (c Consumer) LeftOfHour() (string, error) {

	diff := time.Now().Unix() - c.CreatedAt

	f := float64(diff) / float64(3600)

	var seconds = (int64(math.Ceil(f)) * 3600) - diff

	return durationfmt.Format(time.Second*time.Duration(seconds), "%mm %ss")
}
