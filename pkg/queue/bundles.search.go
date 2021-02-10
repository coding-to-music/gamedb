package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
	"go.uber.org/zap"
)

type BundlesSearchMessage struct {
	Bundle mysql.Bundle `json:"bundle"`
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
		ID:              payload.Bundle.ID,
		UpdatedAt:       payload.Bundle.UpdatedAt,
		Name:            payload.Bundle.Name,
		Discount:        payload.Bundle.Discount,
		SaleDiscount:    payload.Bundle.SaleDiscount,
		HighestDiscount: payload.Bundle.HighestDiscount,
		Apps:            payload.Bundle.AppsCount(),
		Packages:        payload.Bundle.PackagesCount(),
		Icon:            payload.Bundle.Icon,
		Prices:          payload.Bundle.GetPrices(),
		SalePrices:      payload.Bundle.GetPrices(),
		Type:            payload.Bundle.Type,
	}

	err = elasticsearch.IndexBundle(bundle)
	if err != nil {
		log.ErrS(err)
		sendToRetryQueue(message)
		return
	}

	message.Ack()
}
