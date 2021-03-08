package consumers

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.uber.org/zap"
)

type BundlesSearchMessage struct {
	Bundle mongo.Bundle `json:"bundle"`
}

func (m BundlesSearchMessage) Queue() rabbit.QueueName {
	return QueueBundlesSearch
}

func bundleSearchHandler(message *rabbit.Message) {

	payload := BundlesSearchMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	bundle := elasticsearch.Bundle{
		Apps:            len(payload.Bundle.Apps),
		CreatedAt:       payload.Bundle.CreatedAt.Unix(),
		Discount:        payload.Bundle.Discount,
		DiscountHighest: payload.Bundle.DiscountHighest,
		DiscountLowest:  payload.Bundle.DiscountLowest,
		DiscountSale:    payload.Bundle.DiscountSale,
		Giftable:        payload.Bundle.Giftable,
		Icon:            payload.Bundle.Icon,
		ID:              payload.Bundle.ID,
		Image:           payload.Bundle.Image,
		OnSale:          payload.Bundle.OnSale,
		Name:            payload.Bundle.Name,
		Packages:        len(payload.Bundle.Packages),
		Prices:          payload.Bundle.Prices,
		PricesSale:      payload.Bundle.PricesSale,
		Type:            payload.Bundle.Type,
		UpdatedAt:       payload.Bundle.UpdatedAt.Unix(),
		NameMarked:      "",
		Score:           0,
	}

	err = elasticsearch.IndexBundle(bundle)
	if err != nil {
		log.ErrS(err)
		sendToRetryQueue(message)
		return
	}

	message.Ack()
}
