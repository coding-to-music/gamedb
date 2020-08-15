package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"go.uber.org/zap"
)

var (
	countries = []string{"US", "CA", "AF", "AX", "AL", "DZ", "AS", "AD", "AO", "AI", "AQ", "AG", "AR", "AM", "AW", "AU", "AT", "AZ", "BS", "BH", "BD", "BB", "BY", "BE", "BZ", "BJ", "BM", "BT", "BO", "BQ", "BA", "BW", "BV", "BR", "IO", "VG", "BN", "BG", "BF", "BI", "KH", "CM", "CV", "KY", "CF", "TD", "CL", "CN", "CX", "CC", "CO", "KM", "CG", "CD", "CK", "CR", "CI", "HR", "CU", "CW", "CY", "CZ", "DK", "DJ", "DM", "DO", "EC", "EG", "SV", "GQ", "ER", "EE", "ET", "FK", "FO", "FJ", "FI", "FR", "GF", "PF", "TF", "GA", "GM", "GE", "DE", "GH", "GI", "GR", "GL", "GD", "GP", "GU", "GT", "GG", "GN", "GW", "GY", "HT", "HM", "HN", "HK", "HU", "IS", "IN", "ID", "IQ", "IE", "IR", "IM", "IL", "IT", "JM", "JP", "JE", "JO", "KZ", "KE", "KI", "KP", "KR", "XK", "KW", "KG", "LA", "LV", "LB", "LS", "LR", "LY", "LI", "LT", "LU", "MO", "MK", "MG", "MW", "MY", "MV", "ML", "MT", "MH", "MQ", "MR", "MU", "YT", "MX", "FM", "MD", "MC", "MN", "MS", "ME", "MA", "MZ", "MM", "NA", "NR", "NP", "NL", "NC", "NZ", "NI", "NE", "NG", "NU", "NF", "MP", "NO", "OM", "PK", "PW", "PS", "PA", "PG", "PY", "PE", "PH", "PN", "PL", "PT", "PR", "QA", "RE", "RO", "RU", "RW", "BL", "LC", "MF", "WS", "SM", "ST", "SA", "SN", "RS", "SC", "SL", "SG", "SX", "SK", "SI", "SB", "SO", "ZA", "GS", "SS", "ES", "LK", "SH", "KN", "PM", "VC", "SD", "SR", "SJ", "SZ", "SE", "CH", "SY", "TW", "TJ", "TZ", "TH", "TL", "TG", "TK", "TO", "TT", "TN", "TR", "TM", "TC", "TV", "UG", "UA", "AE", "GB", "UM", "VI", "UY", "UZ", "VU", "VA", "VE", "VN", "WF", "EH", "YE", "ZM", "ZW"}

	sessionID        = ""
	steamLoginSecure = ""
)

func syncStates() {

	f, err := os.OpenFile("states.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		zap.S().Error(err)
		return
	}
	defer f.Close()

	for _, v := range countries {

		zap.S().Info(v)

		var urlx = "https://steamcommunity.com/actions/EditProcess?sId=76561197968626192"
		var form = url.Values{}

		form.Set("json", "1")
		form.Set("type", "locationUpdate")
		form.Set("country", v)

		req, err := http.NewRequest("POST", urlx, strings.NewReader(form.Encode()))
		if err != nil {
			zap.S().Error(err)
			continue
		}
		req.Header.Set("content-type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "sessionid", Value: sessionID, Path: "/", Domain: "steamcommunity.com", Secure: true})
		req.AddCookie(&http.Cookie{Name: "steamLoginSecure", Value: steamLoginSecure, Path: "/", Domain: "steamcommunity.com", Secure: true, HttpOnly: true})

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			zap.S().Error(err)
			continue
		}

		b, err := ioutil.ReadAll(resp.Body)

		steamResponse := response{}
		err = json.Unmarshal(b, &steamResponse)
		if err != nil {
			zap.S().Error(err)
			continue
		}

		err = resp.Body.Close()
		if err != nil {
			zap.S().Error(err)
			continue
		}

		f.WriteString(`"` + v + `": {` + "\n")
		for _, v := range steamResponse.State {
			if v.Attribs.Key != "" {
				f.WriteString(`    "` + v.Attribs.Key + `": "` + v.Val + `",` + "\n")
			}
		}
		for _, v := range steamResponse.City {
			if v.Attribs.Key != "" {
				f.WriteString(`    "` + v.Attribs.Key + `": "` + v.Val + `",` + "\n")
			}
		}
		f.WriteString(`},` + "\n")
	}
}

type response struct {
	Results    string `json:"results"`
	Country    string `json:"country"`
	ChangeType string `json:"changeType"`
	State      []struct {
		Attribs struct {
			Key string `json:"key"`
		} `json:"attribs"`
		Val string `json:"val"`
	} `json:"state"`
	City []struct {
		Attribs struct {
			Key string `json:"key"`
		} `json:"attribs"`
		Val string `json:"val"`
	} `json:"city"`
}
