package interactions

import (
	"github.com/bwmarrin/discordgo"
)

type ResponseType int

const (
	ResponseTypePong ResponseType = iota + 1
	ResponseTypeAcknowledge
	ResponseTypeMessage
	ResponseTypeMessageWithSource
	ResponseTypeAcknowledgeWithSource
)

type Response struct {
	Type ResponseType `json:"type"`
	Data ResponseData `json:"data"`
}

type ResponseData struct {
	Content string                    `json:"content"`
	Embeds  []*discordgo.MessageEmbed `json:"embeds"`
}

// type Embed struct {
// 	Description string       `json:"description"` // The main content box of the embed
// 	Color       string       `json:"color"`       // The embed's color in int hex
// 	Fields      []EmbedField `json:"fields"`
// 	Image       EmbedImage   `json:"image"`  // Image
// 	Author      EmbedAuthor  `json:"author"` // Author (top)
// 	Footer      EmbedFooter  `json:"footer"` // Footer (bottom)
// }
//
// type EmbedField struct {
// 	Name   string `json:"name"`
// 	Value  string `json:"value"`
// 	Inline bool   `json:"inline"`
// }
//
// type EmbedAuthor struct {
// 	Name    string `json:"name"`
// 	IconURL string `json:"icon_url"`
// 	URL     string `json:"url"`
// }
//
// type EmbedFooter struct {
// 	Text    string `json:"text"`
// 	IconURL string `json:"icon_url"`
// }
//
// type EmbedImage struct {
// 	URL string `json:"url"`
// }
