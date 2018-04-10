package queue

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/steam-authority/steam-authority/logger"
)

func GetQeueus() (resp []Queue, err error) {

	req, err := http.NewRequest("GET", os.Getenv("STEAM_DOMAIN_RABBIT")+"/api/queues", nil)
	req.SetBasicAuth("guest", "guest")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return resp, err
	}
	defer response.Body.Close()

	// Convert to bytes
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return resp, err
	}

	regex := regexp.MustCompile(`"target_ram_count":\s?(\d+)`)
	s := regex.ReplaceAllString(string(bytes), `"target_ram_count":"$1"`)

	bytes = []byte(s)

	// Unmarshal JSON
	if err := json.Unmarshal(bytes, &resp); err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(bytes))
		}
		return resp, err
	}

	for k, v := range resp {
		resp[k].Name = strings.Replace(v.Name, namespace, "", 1)
	}

	return resp, nil
}

type Queue struct {
	MessagesDetails struct {
		Rate float64 `json:"rate"`
	} `json:"messages_details"`
	Messages int `json:"messages"`
	MessagesUnacknowledgedDetails struct {
		Rate float64 `json:"rate"`
	} `json:"messages_unacknowledged_details"`
	MessagesUnacknowledged int `json:"messages_unacknowledged"`
	MessagesReadyDetails struct {
		Rate float64 `json:"rate"`
	} `json:"messages_ready_details"`
	MessagesReady int `json:"messages_ready"`
	ReductionsDetails struct {
		Rate float64 `json:"rate"`
	} `json:"reductions_details"`
	Reductions int `json:"reductions"`
	MessageStats struct {
		DeliverGetDetails struct {
			Rate float64 `json:"rate"`
		} `json:"deliver_get_details"`
		DeliverGet int `json:"deliver_get"`
		AckDetails struct {
			Rate float64 `json:"rate"`
		} `json:"ack_details"`
		Ack int `json:"ack"`
		RedeliverDetails struct {
			Rate float64 `json:"rate"`
		} `json:"redeliver_details"`
		Redeliver int `json:"redeliver"`
		DeliverNoAckDetails struct {
			Rate float64 `json:"rate"`
		} `json:"deliver_no_ack_details"`
		DeliverNoAck int `json:"deliver_no_ack"`
		DeliverDetails struct {
			Rate float64 `json:"rate"`
		} `json:"deliver_details"`
		Deliver int `json:"deliver"`
		GetNoAckDetails struct {
			Rate float64 `json:"rate"`
		} `json:"get_no_ack_details"`
		GetNoAck int `json:"get_no_ack"`
		GetDetails struct {
			Rate float64 `json:"rate"`
		} `json:"get_details"`
		Get int `json:"get"`
		PublishDetails struct {
			Rate float64 `json:"rate"`
		} `json:"publish_details"`
		Publish int `json:"publish"`
	} `json:"message_stats"`
	Node string `json:"node"`
	Arguments struct {
	} `json:"arguments"`
	Exclusive            bool   `json:"exclusive"`
	AutoDelete           bool   `json:"auto_delete"`
	Durable              bool   `json:"durable"`
	Vhost                string `json:"vhost"`
	Name                 string `json:"name"`
	MessageBytesPagedOut int    `json:"message_bytes_paged_out"`
	MessagesPagedOut     int    `json:"messages_paged_out"`
	BackingQueueStatus struct {
		AvgAckEgressRate  float64       `json:"avg_ack_egress_rate"`
		AvgAckIngressRate float64       `json:"avg_ack_ingress_rate"`
		AvgEgressRate     float64       `json:"avg_egress_rate"`
		AvgIngressRate    float64       `json:"avg_ingress_rate"`
		Delta             []interface{} `json:"delta"`
		Len               int           `json:"len"`
		Mode              string        `json:"mode"`
		NextSeqID         int           `json:"next_seq_id"`
		Q1                int           `json:"q1"`
		Q2                int           `json:"q2"`
		Q3                int           `json:"q3"`
		Q4                int           `json:"q4"`
		TargetRAMCount    string        `json:"target_ram_count"`
	} `json:"backing_queue_status"`
	HeadMessageTimestamp       interface{} `json:"head_message_timestamp"`
	MessageBytesPersistent     int         `json:"message_bytes_persistent"`
	MessageBytesRAM            int         `json:"message_bytes_ram"`
	MessageBytesUnacknowledged int         `json:"message_bytes_unacknowledged"`
	MessageBytesReady          int         `json:"message_bytes_ready"`
	MessageBytes               int         `json:"message_bytes"`
	MessagesPersistent         int         `json:"messages_persistent"`
	MessagesUnacknowledgedRAM  int         `json:"messages_unacknowledged_ram"`
	MessagesReadyRAM           int         `json:"messages_ready_ram"`
	MessagesRAM                int         `json:"messages_ram"`
	GarbageCollection struct {
		MinorGcs        int `json:"minor_gcs"`
		FullsweepAfter  int `json:"fullsweep_after"`
		MinHeapSize     int `json:"min_heap_size"`
		MinBinVheapSize int `json:"min_bin_vheap_size"`
		MaxHeapSize     int `json:"max_heap_size"`
	} `json:"garbage_collection"`
	State                     string        `json:"state"`
	RecoverableSlaves         interface{}   `json:"recoverable_slaves"`
	Consumers                 int           `json:"consumers"`
	ExclusiveConsumerTag      interface{}   `json:"exclusive_consumer_tag"`
	EffectivePolicyDefinition []interface{} `json:"effective_policy_definition"`
	OperatorPolicy            interface{}   `json:"operator_policy"`
	Policy                    interface{}   `json:"policy"`
	ConsumerUtilisation       interface{}   `json:"consumer_utilisation"`
	IdleSince                 string        `json:"idle_since"`
	Memory                    int           `json:"memory"`
}
