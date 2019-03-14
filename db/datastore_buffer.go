package db

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gamedb/website/log"
)

type DatastoreBuffer struct {
	KeyName   string    `gorm:"not null;column:key_name;primary_key"` //
	Kind      string    `gorm:"not null;column:kind;primary_key"`     //
	CreatedAt time.Time `gorm:"not null;column:created_at"`           // Just used for sorting
	RowData   string    `gorm:"not null;column:row_data"`             //
}

func (b DatastoreBuffer) TableName() string {
	return "datastore"
}

func (b DatastoreBuffer) ToChange() (change Change, err error) {

	err = json.Unmarshal([]byte(b.RowData), &change)
	return change, err
}

func (b DatastoreBuffer) ToPlayer() (player Player, err error) {

	err = json.Unmarshal([]byte(b.RowData), &player)
	return player, err
}

func GetBufferRows(kind string, limit int, offset int) (kinds []Kind, err error) {

	gorm, err := GetMySQLClient()
	if err != nil {
		return kinds, err
	}

	gorm = gorm.Where("kind = ?", kind)
	gorm = gorm.Order("created_at DESC")
	gorm = gorm.Limit(limit)
	gorm = gorm.Offset(offset)

	var rows []DatastoreBuffer
	gorm = gorm.Find(&rows)

	for _, v := range rows {

		bufferDataToKind(&kinds, v)
	}

	return kinds, gorm.Error
}

func SaveKindsToBuffer(kinds []Kind, kindType string) (err error) {

	gorm, err := GetMySQLClient()
	if err != nil {
		return err
	}

	for _, kind := range kinds {

		buffer := DatastoreBuffer{}
		buffer.CreatedAt = time.Now()
		buffer.Kind = kindType
		buffer.KeyName = kind.GetKey().Name

		b, err := json.Marshal(kind)

		if err == nil {
			buffer.RowData = string(b)
		}

		gorm = gorm.Save(&buffer)

		return gorm.Error
	}

	return nil
}

var copyMutex sync.Mutex

func CopyBufferToDS() {

	copyMutex.Lock()
	defer copyMutex.Unlock()

	gorm, err := GetMySQLClient()
	if err != nil {
		log.Err(err)
		return
	}

	var counts []struct {
		Kind  string `gorm:"column:kind"`
		Count int    `gorm:"column:count"`
	}

	gorm = gorm.Table("datastore")
	gorm = gorm.Select([]string{"kind", "count(kind) as count"})
	gorm = gorm.Group("kind")
	gorm = gorm.Find(&counts)
	if gorm.Error != nil {
		log.Err(gorm.Error)
		return
	}

	for _, v := range counts {

		if v.Count < 600 {
			log.Info("Less than 600 " + v.Kind + " buffer rows")
			continue
		}

		rows, err := getBufferRows(v.Kind)
		if err != nil {
			log.Err(err)
			continue
		}

		log.Info("Copying " + v.Kind + " buffer rows")

		var kinds []Kind

		for _, vv := range rows {

			bufferDataToKind(&kinds, vv)
		}

		if len(kinds) > 0 {

			err = BulkSaveKinds(kinds, v.Kind, true)
			if err != nil {
				log.Err(err)
				continue
			}

			for _, v := range rows {
				err = deleteBufferRow(v)
				log.Err(err)
			}
		}
	}
}

func bufferDataToKind(kinds *[]Kind, buffer DatastoreBuffer) {

	var err error

	switch buffer.Kind {
	case KindChange:

		var change Change
		change, err = buffer.ToChange()
		*kinds = append(*kinds, change)

	case KindPlayer:

		var player Player
		player, err = buffer.ToPlayer()
		*kinds = append(*kinds, player)

		// case KindEvent:
		//
		// 	var event Event
		// 	*kinds = append(*kinds, event)
		//
		// case KindProductPrice:
		//
		// 	var priceChange ProductPrice
		// 	*kinds = append(*kinds, priceChange)

	}

	log.Err(err)
}

func getBufferRows(kind string) (rows []DatastoreBuffer, err error) {

	gorm, err := GetMySQLClient()
	if err != nil {
		log.Err(err)
		return
	}

	gorm = gorm.Table("datastore")
	gorm = gorm.Where("kind = ?", kind)
	gorm = gorm.Limit(500)
	gorm = gorm.Order("created_at ASC")
	gorm = gorm.Find(&rows)

	return rows, gorm.Error
}

func deleteBufferRow(v DatastoreBuffer) (err error) {

	gorm, err := GetMySQLClient()
	if err != nil {
		log.Err(err)
		return
	}

	gorm = gorm.Where("kind = ?", v.Kind)
	gorm = gorm.Where("key_name = ?", v.KeyName)
	gorm = gorm.Delete(DatastoreBuffer{})

	return gorm.Error
}
