package datastore

import (
	"time"

	"cloud.google.com/go/datastore"
	"github.com/steam-authority/steam-authority/helpers"
)

type Donation struct {
	CreatedAt time.Time `datastore:"created_at"`
	PlayerID  int       `datastore:"player_id"`
	Amount    int       `datastore:"amount"`
	AmountUSD int       `datastore:"amount_usd"`
	Currency  string    `datastore:"currency"`
}

func (d Donation) GetKey() (key *datastore.Key) {
	return datastore.IncompleteKey(KindDonation, nil)
}

func (d Donation) GetCreatedNice() (ret string) {
	return d.CreatedAt.Format(helpers.DateYear)
}

func (d Donation) GetCreatedUnix() int64 {
	return d.CreatedAt.Unix()
}

func GetDonations(playerID int, limit int) (donations []Donation, err error) {

	client, ctx, err := getDSClient()
	if err != nil {
		return donations, err
	}

	q := datastore.NewQuery(KindDonation).Order("-created_at")

	if limit != 0 {
		q = q.Limit(limit)
	}

	if playerID != 0 {
		q = q.Filter("player_id =", playerID)
	}

	client.GetAll(ctx, q, &donations)

	return donations, err
}
