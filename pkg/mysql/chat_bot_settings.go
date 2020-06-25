package mysql

import (
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/memcache"
)

type ChatBotSetting struct {
	CreatedAt   time.Time          `gorm:"not null"`
	UpdatedAt   time.Time          `gorm:"not null"`
	DeletedAt   *time.Time         `gorm:""`
	DiscordID   string             `gorm:"not null;column:discord_id;primary_key"`
	ProductCode steamapi.ProductCC `gorm:"not null;column:product_cc;index:name"`
}

func GetChatBotSettings(discordID string) (settings ChatBotSetting, err error) {

	var item = memcache.MemcacheChatBotSettings(discordID)
	err = memcache.GetSetInterface(item.Key, item.Expiration, &settings, func() (interface{}, error) {

		db, err := GetMySQLClient()
		if err != nil {
			return settings, err
		}

		db = db.Where("discord_id = ?", discordID).First(&settings)
		if db.Error != nil {
			return settings, db.Error
		}

		if settings.ProductCode == "" {
			settings.ProductCode = steamapi.ProductCCUS
		}

		return settings, nil
	})

	return settings, err
}

func SetChatBotSettings(discordID string, callback func(s *ChatBotSetting)) (err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	var settings = ChatBotSetting{
		DiscordID: discordID,
	}

	db = db.Where(settings).FirstOrInit(&settings)
	if db.Error != nil {
		return db.Error
	}

	callback(&settings)

	db = db.Save(&settings)
	if db.Error != nil {
		return db.Error
	}

	return memcache.Delete(memcache.MemcacheChatBotSettings(discordID).Key)
}
