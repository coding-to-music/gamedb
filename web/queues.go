package web

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"sort"

	"github.com/gamedb/website/config"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/go-chi/chi"
)

func queuesRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", queuesHandler)
	r.Get("/ajax.json", queuesJSONHandler)
	return r
}

func queuesHandler(w http.ResponseWriter, r *http.Request) {

	t := queuesTemplate{}
	t.Fill(w, r, "Queues", "When new items get added to the site, they go through a queue to not overload the servers.")

	err := returnTemplate(w, r, "queues", t)
	log.Err(err, r)
}

type queuesTemplate struct {
	GlobalTemplate
}

func queuesJSONHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	s, err := helpers.GetMemcache().GetSetString(helpers.MemcacheQueues, func() (s string, err error) {

		overview, err := getOverview()
		if err != nil {
			return "", err
		}

		var response = queueJSONResponse{}

		// Items
		items := overview.QueueTotals.MessagesDetails.Samples
		sort.Slice(items, func(i, j int) bool {
			return items[i].Timestamp < items[j].Timestamp
		})
		for _, v := range items {
			response.Items = append(response.Items, []int64{v.Timestamp, int64(v.Sample)})
		}

		// Rates
		rates := overview.MessageStats.AckDetails.Samples
		sort.Slice(rates, func(i, j int) bool {
			return rates[i].Timestamp < rates[j].Timestamp
		})

		var last = rates[0].Sample
		for _, v := range rates {
			response.Rates = append(response.Rates, []int64{v.Timestamp, int64(v.Sample - last)})
			last = v.Sample
		}

		bytes, err := json.Marshal(response)
		return string(bytes), err
	})

	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the queues.", Error: err})
		return
	}

	err = returnJSON(w, r, []byte(s))
	log.Err(err, r)
}

type queueJSONResponse struct {
	Items [][]int64 `json:"items"`
	Rates [][]int64 `json:"rate"`
}

func getOverview() (resp Overview, err error) {

	values := url.Values{}
	values.Set("lengths_age", "3600")
	values.Set("lengths_incr", "60")
	values.Set("msg_rates_age", "3600")
	values.Set("msg_rates_incr", "60")
	//values.Set("data_rates_age", "3600")
	//values.Set("data_rates_incr", "60")

	req, err := http.NewRequest("GET", config.Config.RabbitAPI(values), nil)
	req.SetBasicAuth(config.Config.RabbitUsername.Get(), config.Config.RabbitPassword.Get())

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return resp, err
	}

	// Close read
	if response != nil {
		defer func(response *http.Response) {
			err := response.Body.Close()
			log.Err(err)
		}(response)
	}

	// Convert to bytes
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return resp, err
	}

	// Fix JSON
	regex := regexp.MustCompile(`"socket_opts":\[]`)
	s := regex.ReplaceAllString(string(bytes), `"socket_opts":{}`)

	bytes = []byte(s)

	// Unmarshal JSON
	err = helpers.Unmarshal(bytes, &resp)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

type Overview struct {
	ManagementVersion string `json:"management_version"`
	RatesMode         string `json:"rates_mode"`
	ExchangeTypes     []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Enabled     bool   `json:"enabled"`
	} `json:"exchange_types"`
	RabbitmqVersion   string `json:"rabbitmq_version"`
	ClusterName       string `json:"cluster_name"`
	ErlangVersion     string `json:"erlang_version"`
	ErlangFullVersion string `json:"erlang_full_version"`
	MessageStats      struct {
		Ack        int `json:"ack"`
		AckDetails struct {
			Rate    float64  `json:"rate"`
			Samples []Sample `json:"samples"`
			AvgRate float64  `json:"avg_rate"`
			Avg     float64  `json:"avg"`
		} `json:"ack_details"`
		Confirm        int `json:"confirm"`
		ConfirmDetails struct {
			Rate    float64  `json:"rate"`
			Samples []Sample `json:"samples"`
			AvgRate float64  `json:"avg_rate"`
			Avg     float64  `json:"avg"`
		} `json:"confirm_details"`
		Deliver        int `json:"deliver"`
		DeliverDetails struct {
			Rate    float64  `json:"rate"`
			Samples []Sample `json:"samples"`
			AvgRate float64  `json:"avg_rate"`
			Avg     float64  `json:"avg"`
		} `json:"deliver_details"`
		DeliverGet        int `json:"deliver_get"`
		DeliverGetDetails struct {
			Rate    float64  `json:"rate"`
			Samples []Sample `json:"samples"`
			AvgRate float64  `json:"avg_rate"`
			Avg     float64  `json:"avg"`
		} `json:"deliver_get_details"`
		DeliverNoAck        int `json:"deliver_no_ack"`
		DeliverNoAckDetails struct {
			Rate    float64  `json:"rate"`
			Samples []Sample `json:"samples"`
			AvgRate float64  `json:"avg_rate"`
			Avg     float64  `json:"avg"`
		} `json:"deliver_no_ack_details"`
		DiskReads        int `json:"disk_reads"`
		DiskReadsDetails struct {
			Rate    float64  `json:"rate"`
			Samples []Sample `json:"samples"`
			AvgRate float64  `json:"avg_rate"`
			Avg     float64  `json:"avg"`
		} `json:"disk_reads_details"`
		DiskWrites        int `json:"disk_writes"`
		DiskWritesDetails struct {
			Rate    float64  `json:"rate"`
			Samples []Sample `json:"samples"`
			AvgRate float64  `json:"avg_rate"`
			Avg     float64  `json:"avg"`
		} `json:"disk_writes_details"`
		Get        int `json:"get"`
		GetDetails struct {
			Rate    float64  `json:"rate"`
			Samples []Sample `json:"samples"`
			AvgRate float64  `json:"avg_rate"`
			Avg     float64  `json:"avg"`
		} `json:"get_details"`
		GetNoAck        int `json:"get_no_ack"`
		GetNoAckDetails struct {
			Rate    float64  `json:"rate"`
			Samples []Sample `json:"samples"`
			AvgRate float64  `json:"avg_rate"`
			Avg     float64  `json:"avg"`
		} `json:"get_no_ack_details"`
		Publish        int `json:"publish"`
		PublishDetails struct {
			Rate    float64  `json:"rate"`
			Samples []Sample `json:"samples"`
			AvgRate float64  `json:"avg_rate"`
			Avg     float64  `json:"avg"`
		} `json:"publish_details"`
		Redeliver        int `json:"redeliver"`
		RedeliverDetails struct {
			Rate    float64  `json:"rate"`
			Samples []Sample `json:"samples"`
			AvgRate float64  `json:"avg_rate"`
			Avg     float64  `json:"avg"`
		} `json:"redeliver_details"`
		ReturnUnroutable        int `json:"return_unroutable"`
		ReturnUnroutableDetails struct {
			Rate    float64  `json:"rate"`
			Samples []Sample `json:"samples"`
			AvgRate float64  `json:"avg_rate"`
			Avg     float64  `json:"avg"`
		} `json:"return_unroutable_details"`
	} `json:"message_stats"`
	QueueTotals struct {
		Messages        int `json:"messages"`
		MessagesDetails struct {
			Rate    float64  `json:"rate"`
			Samples []Sample `json:"samples"`
			AvgRate float64  `json:"avg_rate"`
			Avg     float64  `json:"avg"`
		} `json:"messages_details"`
		MessagesReady        int `json:"messages_ready"`
		MessagesReadyDetails struct {
			Rate    float64  `json:"rate"`
			Samples []Sample `json:"samples"`
			AvgRate float64  `json:"avg_rate"`
			Avg     float64  `json:"avg"`
		} `json:"messages_ready_details"`
		MessagesUnacknowledged        int `json:"messages_unacknowledged"`
		MessagesUnacknowledgedDetails struct {
			Rate    float64  `json:"rate"`
			Samples []Sample `json:"samples"`
			AvgRate float64  `json:"avg_rate"`
			Avg     float64  `json:"avg"`
		} `json:"messages_unacknowledged_details"`
	} `json:"queue_totals"`
	ObjectTotals struct {
		Channels    int `json:"channels"`
		Connections int `json:"connections"`
		Consumers   int `json:"consumers"`
		Exchanges   int `json:"exchanges"`
		Queues      int `json:"queues"`
	} `json:"object_totals"`
	StatisticsDbEventQueue int    `json:"statistics_db_event_queue"`
	Node                   string `json:"node"`
	Listeners              []struct {
		Node       string `json:"node"`
		Protocol   string `json:"protocol"`
		IPAddress  string `json:"ip_address"`
		Port       int    `json:"port"`
		SocketOpts struct {
			Backlog     int           `json:"backlog"`
			Nodelay     bool          `json:"nodelay"`
			Linger      []interface{} `json:"linger"`
			ExitOnClose bool          `json:"exit_on_close"`
		} `json:"socket_opts"`
	} `json:"listeners"`
	Contexts []struct {
		SslOpts     []interface{} `json:"ssl_opts"`
		Node        string        `json:"node"`
		Description string        `json:"description"`
		Path        string        `json:"path"`
		Port        string        `json:"port"`
		Ssl         string        `json:"ssl"`
	} `json:"contexts"`
}

type Sample struct {
	Sample    int   `json:"sample"`
	Timestamp int64 `json:"timestamp"`
}
