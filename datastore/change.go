package datastore

import (
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/steam-authority/steam-authority/helpers"
)

type Change struct {
	CreatedAt time.Time `datastore:"created_at"`
	ChangeID  int       `datastore:"change_id"`
	Apps      []int     `datastore:"apps,noindex"`
	Packages  []int     `datastore:"packages,noindex"`
}

func (change Change) GetKey() (key *datastore.Key) {
	return datastore.NameKey(KindChange, strconv.Itoa(change.ChangeID), nil)
}

func (change Change) GetName() (name string) {

	return "Change " + strconv.Itoa(change.ChangeID)
}

func (change Change) GetTimestamp() (int64) {
	return change.CreatedAt.Unix()
}

func (change Change) GetNiceDate() (string) {
	return change.CreatedAt.Format(helpers.DateYearTime)
}

func (change Change) GetPath() string {
	return "/changes/" + strconv.Itoa(change.ChangeID)
}

func (change *Change) AddApp(app int) {
	change.Apps = append(change.Apps, app)
}

func (change *Change) AddPackage(pack int) {
	change.Packages = append(change.Packages, pack)
}

func GetLatestChanges(limit int, page int) (changes []Change, err error) {

	client, ctx, err := getClient()
	if err != nil {
		return changes, err
	}

	offset := (page - 1) * 100

	q := datastore.NewQuery(KindChange).Order("-change_id").Limit(limit).Offset(offset)

	client.GetAll(ctx, q, &changes)

	return changes, err
}

func GetChange(id string) (change Change, err error) {

	client, context, err := getClient()
	if err != nil {
		return change, err
	}

	key := datastore.NameKey(KindChange, id, nil)

	change = Change{}
	err = client.Get(context, key, &change)
	if err != nil {
		return change, err
	}

	return change, nil
}

// todo, handle more than 500
func BulkAddAChanges(changes []*Change) (err error) {

	changesLen := len(changes)
	if changesLen == 0 {
		return nil
	}

	client, context, err := getClient()
	if err != nil {
		return err
	}

	keys := make([]*datastore.Key, 0, changesLen)

	for _, v := range changes {
		keys = append(keys, v.GetKey())
	}

	_, err = client.PutMulti(context, keys, changes)
	if err != nil {
		return err
	}

	return nil
}
