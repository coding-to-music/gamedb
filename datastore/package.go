package datastore

import (
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
)

type Package struct {
	CreatedAt    time.Time         `datastore:"created_at"`
	UpdatedAt    time.Time         `datastore:"updated_at"`
	Name         string            `datastore:"name"`
	ChangeID     int               `datastore:"change_id"`
	PackageID    int               `datastore:"packageid"`
	BillingType  int               `datastore:"billingtype"`
	LicensetType int               `datastore:"licensetype"`
	Status       int               `datastore:"status"`
	AppIDs       []int             `datastore:"appids"`
	DepotIDs     []int             `datastore:"depotids"`
	Extended     map[string]string `datastore:"extended"`
	Raw          string            `datastore:"raw"`
}

func (p Package) GetKey() (key *datastore.Key) {
	return datastore.NameKey(KindPlayer, strconv.FormatInt(int64(p.ChangeID), 10), nil)
}
