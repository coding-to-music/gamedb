package interactions

import (
	"time"
)

type Event struct {
	ChannelID string `json:"channel_id"`
	Data      Data   `json:"data"`
	GuildID   string `json:"guild_id"`
	ID        string `json:"id"`
	Member    Member `json:"member"`
	Token     string `json:"token"`
	Type      int    `json:"type"`
	Version   int    `json:"version"`
}

type Data struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type User struct {
	Avatar        string `json:"avatar"`
	Discriminator string `json:"discriminator"`
	ID            string `json:"id"`
	PublicFlags   int    `json:"public_flags"`
	Username      string `json:"username"`
}

type Member struct {
	Deaf         bool        `json:"deaf"`
	IsPending    bool        `json:"is_pending"`
	JoinedAt     time.Time   `json:"joined_at"`
	Mute         bool        `json:"mute"`
	Nick         interface{} `json:"nick"`
	Pending      bool        `json:"pending"`
	Permissions  string      `json:"permissions"`
	PremiumSince interface{} `json:"premium_since"`
	Roles        []string    `json:"roles"`
	User         User        `json:"user"`
}
