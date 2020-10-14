package geo

import (
	"errors"
	"net"
	"regexp"
	"strings"
	"sync"

	"github.com/oschwald/maxminddb-golang"
)

var (
	lock    sync.Mutex
	maxMind *maxminddb.Reader
)

func GetLocation(ipIn string) (record *Record, err error) {

	lock.Lock()
	defer lock.Unlock()

	if maxMind == nil {
		maxMind, err = maxminddb.Open("./assets/GeoLite2-City.mmdb")
		if err != nil {
			return nil, err
		}
	}

	ip := net.ParseIP(GetFirstIP(ipIn))
	if ip == nil {
		return nil, errors.New("invalid ip")
	}

	record = &Record{}
	err = maxMind.Lookup(ip, record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

// More fields available @ https://github.com/oschwald/geoip2-golang/blob/master/reader.go#L85
// Only using what we need is faster
type Record struct {
	City struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
	Continent struct {
		Code  string            `maxminddb:"code"`
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"continent"`
	Country struct {
		ISOCode           string            `maxminddb:"iso_code"`
		Names             map[string]string `maxminddb:"names"`
		IsInEuropeanUnion bool              `maxminddb:"is_in_european_union"`
	} `maxminddb:"country"`
}

var trimPort = regexp.MustCompile(`:[0-9]{2,}$`)

func GetFirstIP(ip string) string {

	for _, v := range strings.Split(ip, ",") {

		v = strings.TrimSpace(v)
		v = trimPort.ReplaceAllString(v, "")
		if v != "" {
			return v
		}
	}

	return ip
}
