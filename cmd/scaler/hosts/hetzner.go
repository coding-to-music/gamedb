package hosts

import (
	"context"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/github"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

var (
	hetznerClient  *hcloud.Client
	hetznerContext context.Context
)

func getHetzner() (*hcloud.Client, context.Context) {

	if hetznerClient == nil {
		hetznerClient = hcloud.NewClient(hcloud.WithToken(config.C.HetznerAPIToken))
		hetznerContext = context.TODO()
	}

	return hetznerClient, hetznerContext
}

type Hetzner struct {
}

func (h Hetzner) ListConsumers() (consumers []Consumer, err error) {

	client, ctx := getHetzner()

	servers, err := client.Server.All(ctx)
	if err != nil {
		return consumers, err
	}

	for _, server := range servers {

		var labels []string
		for k := range server.Labels {
			labels = append(labels, k)
		}

		consumers = append(consumers, Consumer{
			ID:        server.ID,
			Name:      server.Name,
			IP:        server.PublicNet.IPv4.IP.String(),
			Tags:      labels,
			CreatedAt: server.Created.Unix(),
			Locked:    server.Locked,
		})
	}

	return consumers, nil
}

func (h Hetzner) CreateConsumer() (c Consumer, err error) {

	gh, ctx := github.GetGithub()
	ghResponse, _, _, err := gh.Repositories.GetContents(ctx, "gamedb", "infrastructure", "scaler/cloud-config.yaml", nil)
	if err != nil {
		return c, err
	}

	body, _, err := helpers.GetWithTimeout(ghResponse.GetDownloadURL(), 0)
	if err != nil {
		return c, err
	}

	client, ctx := getHetzner()

	_, _, err = client.Server.Create(ctx, hcloud.ServerCreateOpts{
		Name:       "gamedb-consumer-" + helpers.RandString(5, helpers.Letters),
		ServerType: &hcloud.ServerType{Name: "cx11"},
		Image:      &hcloud.Image{Name: "debian-10"},
		SSHKeys:    []*hcloud.SSHKey{{ID: config.C.HetznerSSHKeyID}},
		Datacenter: &hcloud.Datacenter{Name: "nbg1-dc3"},
		UserData:   string(body),
		Labels:     map[string]string{"consumers": "", ConsumerTag: ""},
		Networks:   []*hcloud.Network{{ID: config.C.HetznerNetworkID}},
	})

	return c, err
}

func (h Hetzner) DeleteConsumer(id int) (err error) {

	client, ctx := getHetzner()

	_, err = client.Server.Delete(ctx, &hcloud.Server{ID: id})
	return err
}
