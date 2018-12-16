package web

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"

	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/go-chi/chi"
	"github.com/spf13/viper"
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
	log.Log(err)
}

type queuesTemplate struct {
	GlobalTemplate
}

func queuesJSONHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	overview, err := getOverview()
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the queues.", Error: err})
		return
	}

	var ret [][]interface{}

	for _, v := range overview.QueueTotals.MessagesDetails.Samples {
		ret = append(ret, []interface{}{v.Timestamp, v.Sample})
	}

	// Encode
	bytes, err := json.Marshal(ret)
	if err != nil {
		log.Log(err)
		bytes = []byte("[]")
	}

	err = returnJSON(w, r, bytes)
	log.Log(err)
}

func getOverview() (resp Overview, err error) {

	URL := "http://" + os.Getenv("STEAM_RABBIT_HOST") + ":" + viper.GetString("RABBIT_MANAGEMENT_PORT")
	URL += "/api/overview?lengths_age=3600&lengths_incr=60&msg_rates_age=3600&msg_rates_incr=60"

	req, err := http.NewRequest("GET", URL, nil)
	req.SetBasicAuth(os.Getenv("STEAM_RABBIT_USER"), os.Getenv("STEAM_RABBIT_PASS"))

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return resp, err
	}

	// Close read
	if response != nil {
		defer func(response *http.Response) {
			err := response.Body.Close()
			log.Log(err)
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
			Rate    float64 `json:"rate"`
			Samples []struct {
				Sample    int   `json:"sample"`
				Timestamp int64 `json:"timestamp"`
			} `json:"samples"`
			AvgRate float64 `json:"avg_rate"`
			Avg     float64 `json:"avg"`
		} `json:"ack_details"`
		Confirm        int `json:"confirm"`
		ConfirmDetails struct {
			Rate    float64 `json:"rate"`
			Samples []struct {
				Sample    int   `json:"sample"`
				Timestamp int64 `json:"timestamp"`
			} `json:"samples"`
			AvgRate float64 `json:"avg_rate"`
			Avg     float64 `json:"avg"`
		} `json:"confirm_details"`
		Deliver        int `json:"deliver"`
		DeliverDetails struct {
			Rate    float64 `json:"rate"`
			Samples []struct {
				Sample    int   `json:"sample"`
				Timestamp int64 `json:"timestamp"`
			} `json:"samples"`
			AvgRate float64 `json:"avg_rate"`
			Avg     float64 `json:"avg"`
		} `json:"deliver_details"`
		DeliverGet        int `json:"deliver_get"`
		DeliverGetDetails struct {
			Rate    float64 `json:"rate"`
			Samples []struct {
				Sample    int   `json:"sample"`
				Timestamp int64 `json:"timestamp"`
			} `json:"samples"`
			AvgRate float64 `json:"avg_rate"`
			Avg     float64 `json:"avg"`
		} `json:"deliver_get_details"`
		DeliverNoAck        int `json:"deliver_no_ack"`
		DeliverNoAckDetails struct {
			Rate    float64 `json:"rate"`
			Samples []struct {
				Sample    int   `json:"sample"`
				Timestamp int64 `json:"timestamp"`
			} `json:"samples"`
			AvgRate float64 `json:"avg_rate"`
			Avg     float64 `json:"avg"`
		} `json:"deliver_no_ack_details"`
		DiskReads        int `json:"disk_reads"`
		DiskReadsDetails struct {
			Rate    float64 `json:"rate"`
			Samples []struct {
				Sample    int   `json:"sample"`
				Timestamp int64 `json:"timestamp"`
			} `json:"samples"`
			AvgRate float64 `json:"avg_rate"`
			Avg     float64 `json:"avg"`
		} `json:"disk_reads_details"`
		DiskWrites        int `json:"disk_writes"`
		DiskWritesDetails struct {
			Rate    float64 `json:"rate"`
			Samples []struct {
				Sample    int   `json:"sample"`
				Timestamp int64 `json:"timestamp"`
			} `json:"samples"`
			AvgRate float64 `json:"avg_rate"`
			Avg     float64 `json:"avg"`
		} `json:"disk_writes_details"`
		Get        int `json:"get"`
		GetDetails struct {
			Rate    float64 `json:"rate"`
			Samples []struct {
				Sample    int   `json:"sample"`
				Timestamp int64 `json:"timestamp"`
			} `json:"samples"`
			AvgRate float64 `json:"avg_rate"`
			Avg     float64 `json:"avg"`
		} `json:"get_details"`
		GetNoAck        int `json:"get_no_ack"`
		GetNoAckDetails struct {
			Rate    float64 `json:"rate"`
			Samples []struct {
				Sample    int   `json:"sample"`
				Timestamp int64 `json:"timestamp"`
			} `json:"samples"`
			AvgRate float64 `json:"avg_rate"`
			Avg     float64 `json:"avg"`
		} `json:"get_no_ack_details"`
		Publish        int `json:"publish"`
		PublishDetails struct {
			Rate    float64 `json:"rate"`
			Samples []struct {
				Sample    int   `json:"sample"`
				Timestamp int64 `json:"timestamp"`
			} `json:"samples"`
			AvgRate float64 `json:"avg_rate"`
			Avg     float64 `json:"avg"`
		} `json:"publish_details"`
		Redeliver        int `json:"redeliver"`
		RedeliverDetails struct {
			Rate    float64 `json:"rate"`
			Samples []struct {
				Sample    int   `json:"sample"`
				Timestamp int64 `json:"timestamp"`
			} `json:"samples"`
			AvgRate float64 `json:"avg_rate"`
			Avg     float64 `json:"avg"`
		} `json:"redeliver_details"`
		ReturnUnroutable        int `json:"return_unroutable"`
		ReturnUnroutableDetails struct {
			Rate    float64 `json:"rate"`
			Samples []struct {
				Sample    int   `json:"sample"`
				Timestamp int64 `json:"timestamp"`
			} `json:"samples"`
			AvgRate float64 `json:"avg_rate"`
			Avg     float64 `json:"avg"`
		} `json:"return_unroutable_details"`
	} `json:"message_stats"`
	QueueTotals struct {
		Messages        int `json:"messages"`
		MessagesDetails struct {
			Rate    float64 `json:"rate"`
			Samples []struct {
				Sample    int   `json:"sample"`
				Timestamp int64 `json:"timestamp"`
			} `json:"samples"`
			AvgRate float64 `json:"avg_rate"`
			Avg     float64 `json:"avg"`
		} `json:"messages_details"`
		MessagesReady        int `json:"messages_ready"`
		MessagesReadyDetails struct {
			Rate    float64 `json:"rate"`
			Samples []struct {
				Sample    int   `json:"sample"`
				Timestamp int64 `json:"timestamp"`
			} `json:"samples"`
			AvgRate float64 `json:"avg_rate"`
			Avg     float64 `json:"avg"`
		} `json:"messages_ready_details"`
		MessagesUnacknowledged        int `json:"messages_unacknowledged"`
		MessagesUnacknowledgedDetails struct {
			Rate    float64 `json:"rate"`
			Samples []struct {
				Sample    int   `json:"sample"`
				Timestamp int64 `json:"timestamp"`
			} `json:"samples"`
			AvgRate float64 `json:"avg_rate"`
			Avg     float64 `json:"avg"`
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
