package db

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

	client, ctx, err := GetDSClient()
	if err != nil {
		return changes, err
	}

	offset := (page - 1) * 100

	q := datastore.NewQuery(KindChange).Order("-change_id").Limit(limit).Offset(offset)

	_, err = client.GetAll(ctx, q, &changes)

	err = checkForMissingChangeFields(err)
	if err != nil {
		return changes, err
	}

	return changes, nil
}

func GetChange(id string) (change Change, err error) {

	client, context, err := GetDSClient()
	if err != nil {
		return change, err
	}

	key := datastore.NameKey(KindChange, id, nil)

	change = Change{}
	err = client.Get(context, key, &change)
	if err != nil {
		if err, ok := err.(*datastore.ErrFieldMismatch); ok {

			old := []string{
				"updated_at",
			}

			if !helpers.SliceHasString(old, err.FieldName) {
				return change, err
			}
		} else {
			return change, err
		}
	}

	return change, nil
}

func BulkAddAChanges(changes []*Change) (keys []*datastore.Key, err error) {

	if len(changes) == 0 {
		return
	}

	client, ctx, err := GetDSClient()
	if err != nil {
		return
	}

	chunks := chunkChanges(changes, 500)

	for _, chunk := range chunks {

		multiKeys := make([]*datastore.Key, 0, len(chunk))
		for _, v := range chunk {
			multiKeys = append(multiKeys, v.GetKey())
		}

		keys, err = client.PutMulti(ctx, multiKeys, chunk)
		if err != nil {
			return
		}
	}

	return
}

func chunkChanges(changes []*Change, chunkSize int) (divided [][]*Change) {

	for i := 0; i < len(changes); i += chunkSize {
		end := i + chunkSize

		if end > len(changes) {
			end = len(changes)
		}

		divided = append(divided, changes[i:end])
	}

	return divided
}

func checkForMissingChangeFields(err error) error {

	if err == nil {
		return nil
	}

	if err2, ok := err.(*datastore.ErrFieldMismatch); ok {

		removedColumns := []string{
			"updated_at",
		}

		if helpers.SliceHasString(removedColumns, err2.FieldName) {
			return nil
		}
	}

	return err
}
