package hosts

import (
	"context"
	"io/ioutil"
	"time"

	"github.com/digitalocean/godo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"golang.org/x/oauth2"
)

type DigitalOcean struct {
	ctx    context.Context
	client *godo.Client
}

func (do DigitalOcean) getClient() *godo.Client {

	if do.client == nil {

		oauthClient := oauth2.NewClient(context.Background(), &DOTokenSource{
			AccessToken: config.Config.DigitalOceanAccessToken.Get(),
		})

		do.client = godo.NewClient(oauthClient)
	}

	return do.client
}

func (do DigitalOcean) getContext() context.Context {

	if do.ctx == nil {
		do.ctx = context.TODO()
	}

	return do.ctx
}

func (do DigitalOcean) ListConsumers() (consumers []Consumer, err error) {

	droplets, _, err := do.getClient().Droplets.ListByTag(do.getContext(), ConsumerTag, &godo.ListOptions{PerPage: 100, Page: 1})
	if err != nil {
		return
	}

	for _, v := range droplets {

		var ip string
		ip, err = v.PublicIPv4()
		if err != nil {
			return
		}
		if ip == "" {
			ip = "-"
		}

		var t time.Time
		t, err = time.Parse(time.RFC3339, v.Created)
		if err != nil {
			return
		}

		consumers = append(consumers, Consumer{
			ID:        v.ID,
			Name:      v.Name,
			IP:        ip,
			Tags:      v.Tags,
			CreatedAt: t.Unix(),
		})
	}

	return
}

func (do DigitalOcean) CreateConsumer() (consumer Consumer, err error) {

	// todo, download file off github
	cc, err := ioutil.ReadFile("cloud-config.yaml")
	if err != nil {
		return consumer, err
	}

	//
	key2 := godo.Key{
		ID:          config.Config.DigitalOceanKeyID.GetInt(),
		Fingerprint: config.Config.DigitalOceanKeyFingerprint.Get(),
	}

	createRequest := &godo.DropletCreateRequest{
		Name:              "gamedb-consumer-" + helpers.RandString(5, helpers.Letters),
		Region:            "ams3",
		Size:              "s-1vcpu-1gb",
		Image:             godo.DropletCreateImage{Slug: "debian-9-x64"},
		SSHKeys:           []godo.DropletCreateSSHKey{{ID: key2.ID, Fingerprint: key2.Fingerprint}},
		Monitoring:        true,
		UserData:          string(cc),
		Tags:              []string{ConsumerTag},
		PrivateNetworking: true,
	}

	droplet, _, err := do.getClient().Droplets.Create(do.getContext(), createRequest)
	if err != nil {
		return consumer, err
	}

	_, _, err = do.getClient().Projects.AssignResources(do.getContext(), config.Config.DigitalOceanProjectID.Get(), droplet.URN())
	return consumer, err
}

func (do DigitalOcean) DeleteConsumer(id int) (err error) {

	_, err = do.getClient().Droplets.Delete(do.getContext(), id)
	if err != nil {
		return
	}

	time.Sleep(time.Second * 3)

	return err
}

type DOTokenSource struct {
	AccessToken string
}

func (t *DOTokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}
