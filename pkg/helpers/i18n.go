package helpers

import (
	"sort"
	"strings"

	"github.com/Jleagle/steam-go/steam"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/pariz/gountries"
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var gountriesInstance = gountries.New()

type ProductCountryCode struct {
	ProductCode  steam.ProductCC
	CurrencyCode steam.CurrencyCode
	CountryCodes []string // Used to get a currency from an country from an IP
	Name         string
	Symbol       string
	Enabled      bool
}

func (pcc ProductCountryCode) GetFlag() string {
	switch pcc.ProductCode {
	case steam.ProductCCEU:
		return "eu"
	case steam.ProductCCUK:
		return "gb"
	case steam.ProductCCAZ:
		return "cis"
	case steam.ProductCCPK:
		return "sasia"
	default:
		return strings.ToLower(pcc.CountryCodes[0])
	}
}

var ProductCountryCodes = map[steam.ProductCC]ProductCountryCode{
	steam.ProductCCAR: {
		ProductCode:  steam.ProductCCAR,
		CountryCodes: []string{"AR"},
		CurrencyCode: steam.CurrencyARS,
		Name:         "Argentine Peso",
		Symbol:       "ARS$",
	},
	steam.ProductCCAU: {
		ProductCode:  steam.ProductCCAU,
		CountryCodes: []string{"AU"},
		CurrencyCode: steam.CurrencyAUD,
		Name:         "Australian Dollar",
		Symbol:       "A$",
	},
	steam.ProductCCBR: {
		ProductCode:  steam.ProductCCBR,
		CountryCodes: []string{"BR"},
		CurrencyCode: steam.CurrencyBRL,
		Name:         "Brazilian Real",
		Symbol:       "R$",
	},
	steam.ProductCCCA: {
		ProductCode:  steam.ProductCCCA,
		CountryCodes: []string{"CA"},
		CurrencyCode: steam.CurrencyCAD,
		Name:         "Canadian Dollar",
		Symbol:       "CDN$",
	},
	steam.ProductCCCL: {
		ProductCode:  steam.ProductCCCL,
		CountryCodes: []string{"CL"},
		CurrencyCode: steam.CurrencyCLP,
		Name:         "Chilean Peso",
		Symbol:       "CLP$",
	},
	steam.ProductCCCN: {
		ProductCode:  steam.ProductCCCN,
		CountryCodes: []string{"CN"},
		CurrencyCode: steam.CurrencyCNY,
		Name:         "Chinese Renminbi",
		Symbol:       "¥",
		Enabled:      true,
	},
	steam.ProductCCCO: {
		ProductCode:  steam.ProductCCCO,
		CountryCodes: []string{"CO"},
		CurrencyCode: steam.CurrencyCOP,
		Name:         "Colombian Peso",
		Symbol:       "COL$",
	},
	steam.ProductCCCR: {
		ProductCode:  steam.ProductCCCR,
		CountryCodes: []string{"CR"},
		CurrencyCode: steam.CurrencyCRC,
		Name:         "Costa Rican Colon",
		Symbol:       "₡",
	},
	steam.ProductCCEU: { // European Union
		ProductCode:  steam.ProductCCEU,
		CountryCodes: []string{"AT", "BE", "BG", "HR", "CY", "CZ", "DK", "EE", "FI", "FR", "DE", "EL", "HU", "IE", "IT", "LV", "LT", "LU", "MT", "NL", "PL", "PT", "RO", "SK", "SI", "ES", "SE"},
		CurrencyCode: steam.CurrencyEUR,
		Name:         "Euro",
		Symbol:       "€",
		Enabled:      true,
	},
	steam.ProductCCHK: {
		ProductCode:  steam.ProductCCHK,
		CountryCodes: []string{"HK"},
		CurrencyCode: steam.CurrencyHKD,
		Name:         "Hong Kong Dollar",
		Symbol:       "HK$",
	},
	steam.ProductCCIN: {
		ProductCode:  steam.ProductCCIN,
		CountryCodes: []string{"IN"},
		CurrencyCode: steam.CurrencyINR,
		Name:         "Indian Rupee",
		Symbol:       "₹",
	},
	steam.ProductCCID: {
		ProductCode:  steam.ProductCCID,
		CountryCodes: []string{"ID"},
		CurrencyCode: steam.CurrencyIDR,
		Name:         "Indonesian Rupiah",
		Symbol:       "Rp",
	},
	steam.ProductCCIL: {
		ProductCode:  steam.ProductCCIL,
		CountryCodes: []string{"IL"},
		CurrencyCode: steam.CurrencyILS,
		Name:         "Israeli New Shekel",
		Symbol:       "₪",
	},
	steam.ProductCCJP: {
		ProductCode:  steam.ProductCCJP,
		CountryCodes: []string{"JP"},
		CurrencyCode: steam.CurrencyJPY,
		Name:         "Japanese Yen",
		Symbol:       "¥",
	},
	steam.ProductCCKZ: {
		ProductCode:  steam.ProductCCKZ,
		CountryCodes: []string{"KZ"},
		CurrencyCode: steam.CurrencyKZT,
		Name:         "Kazakhstani Tenge",
		Symbol:       "₸",
	},
	steam.ProductCCKW: {
		ProductCode:  steam.ProductCCKW,
		CountryCodes: []string{"KW"},
		CurrencyCode: steam.CurrencyKWD,
		Name:         "Kuwaiti Dinar",
		Symbol:       "KD",
	},
	steam.ProductCCMY: {
		ProductCode:  steam.ProductCCMY,
		CountryCodes: []string{"MY"},
		CurrencyCode: steam.CurrencyMYR,
		Name:         "Malaysian Ringgit",
		Symbol:       "RM",
	},
	steam.ProductCCMX: {
		ProductCode:  steam.ProductCCMX,
		CountryCodes: []string{"MX"},
		CurrencyCode: steam.CurrencyMXN,
		Name:         "Mexican Peso",
		Symbol:       "Mex$",
	},
	steam.ProductCCTW: {
		ProductCode:  steam.ProductCCTW,
		CountryCodes: []string{"TW"},
		CurrencyCode: steam.CurrencyTWD,
		Name:         "New Taiwan Dollar",
		Symbol:       "NT$",
	},
	steam.ProductCCNZ: {
		ProductCode:  steam.ProductCCNZ,
		CountryCodes: []string{"NZ"},
		CurrencyCode: steam.CurrencyNZD,
		Name:         "New Zealand Dollar",
		Symbol:       "NZ$",
	},
	steam.ProductCCNO: {
		ProductCode:  steam.ProductCCNO,
		CountryCodes: []string{"NO"},
		CurrencyCode: steam.CurrencyNOK,
		Name:         "Norwegian Krone",
		Symbol:       "kr",
	},
	steam.ProductCCPE: {
		ProductCode:  steam.ProductCCPE,
		CountryCodes: []string{"PE"},
		CurrencyCode: steam.CurrencyPEN,
		Name:         "Peruvian Sol",
		Symbol:       "S/",
	},
	steam.ProductCCPH: {
		ProductCode:  steam.ProductCCPH,
		CountryCodes: []string{"PH"},
		CurrencyCode: steam.CurrencyPHP,
		Name:         "Philippine Peso",
		Symbol:       "₱",
	},
	steam.ProductCCPL: {
		ProductCode:  steam.ProductCCPL,
		CountryCodes: []string{"PL"},
		CurrencyCode: steam.CurrencyPLN,
		Name:         "Polish Zloty",
		Symbol:       "zł",
	},
	steam.ProductCCUK: {
		ProductCode:  steam.ProductCCUK,
		CountryCodes: []string{"GB"},
		CurrencyCode: steam.CurrencyGBP,
		Name:         "Pound Sterling",
		Symbol:       "£",
		Enabled:      true,
	},
	steam.ProductCCQA: {
		ProductCode:  steam.ProductCCQA,
		CountryCodes: []string{"QA"},
		CurrencyCode: steam.CurrencyQAR,
		Name:         "Qatari Riyal",
		Symbol:       "QR",
	},
	steam.ProductCCRU: {
		ProductCode:  steam.ProductCCRU,
		CountryCodes: []string{"RU"},
		CurrencyCode: steam.CurrencyRUB,
		Name:         "Russian Ruble",
		Symbol:       "₽",
		Enabled:      true,
	},
	steam.ProductCCSA: {
		ProductCode:  steam.ProductCCSA,
		CountryCodes: []string{"SA"},
		CurrencyCode: steam.CurrencySAR,
		Name:         "Saudi Riyal",
		Symbol:       "SR",
	},
	steam.ProductCCSG: {
		ProductCode:  steam.ProductCCSG,
		CountryCodes: []string{"SG"},
		CurrencyCode: steam.CurrencySGD,
		Name:         "Singapore Dollar",
		Symbol:       "S$",
	},
	steam.ProductCCZA: {
		ProductCode:  steam.ProductCCZA,
		CountryCodes: []string{"ZA"},
		CurrencyCode: steam.CurrencyZAR,
		Name:         "South African Rand",
		Symbol:       "R",
	},
	steam.ProductCCKR: {
		ProductCode:  steam.ProductCCKR,
		CountryCodes: []string{"KR"},
		CurrencyCode: steam.CurrencyKRW,
		Name:         "South Korean Won",
		Symbol:       "₩",
	},
	steam.ProductCCCH: {
		ProductCode:  steam.ProductCCCH,
		CountryCodes: []string{"CH"},
		CurrencyCode: steam.CurrencyCHF,
		Name:         "Swiss Franc",
		Symbol:       "CHF",
	},
	steam.ProductCCTH: {
		ProductCode:  steam.ProductCCTH,
		CountryCodes: []string{"TH"},
		CurrencyCode: steam.CurrencyTHB,
		Name:         "Thai Baht",
		Symbol:       "฿",
	},
	steam.ProductCCTR: {
		ProductCode:  steam.ProductCCTR,
		CountryCodes: []string{"TR"},
		CurrencyCode: steam.CurrencyTRY,
		Name:         "Turkish Lira",
		Symbol:       "₺",
	},
	steam.ProductCCUA: {
		ProductCode:  steam.ProductCCUA,
		CountryCodes: []string{"UA"},
		CurrencyCode: steam.CurrencyUAH,
		Name:         "Ukrainian Hryvnia",
		Symbol:       "₴",
	},
	steam.ProductCCAE: {
		ProductCode:  steam.ProductCCAE,
		CountryCodes: []string{"AE"},
		CurrencyCode: steam.CurrencyAED,
		Name:         "United Arab Emirates Dirham",
		Symbol:       "AED"},
	steam.ProductCCUS: {
		ProductCode:  steam.ProductCCUS,
		CountryCodes: []string{"US"},
		CurrencyCode: steam.CurrencyUSD,
		Name:         "United States Dollar",
		Symbol:       "$",
		Enabled:      true,
	},
	steam.ProductCCAZ: { // CIS
		ProductCode:  steam.ProductCCAZ,
		CountryCodes: []string{"AM", "AZ", "BY", "GE", "KZ", "KG", "MD", "TJ", "TM", "UZ", "UA"},
		CurrencyCode: steam.CurrencyUSD,
		Name:         "United States Dollar (CIS)",
		Symbol:       "$",
	},
	steam.ProductCCPK: { // SASIA
		ProductCode:  steam.ProductCCPK,
		CountryCodes: []string{"BD", "BT", "NP", "PK", "LK"},
		CurrencyCode: steam.CurrencyUSD,
		Name:         "United States Dollar (South Asia)",
		Symbol:       "$",
	},
	steam.ProductCCUY: {
		ProductCode:  steam.ProductCCUY,
		CountryCodes: []string{"UY"},
		CurrencyCode: steam.CurrencyUYU,
		Name:         "Uruguayan Peso",
		Symbol:       "$U",
	},
	steam.ProductCCVN: {
		ProductCode:  steam.ProductCCVN,
		CountryCodes: []string{"VN"},
		CurrencyCode: steam.CurrencyVND,
		Name:         "Vietnamese Dong",
		Symbol:       "₫",
	},
}

func IsValidProdCC(cc steam.ProductCC) bool {
	_, ok := ProductCountryCodes[cc]
	return ok
}

func GetProdCC(cc steam.ProductCC) ProductCountryCode {
	val, ok := ProductCountryCodes[cc]
	if ok {
		return val
	}
	return ProductCountryCodes[steam.ProductCCUS]
}

func GetProdCCs(activeOnly bool) (ccs []ProductCountryCode) {

	for _, v := range ProductCountryCodes {
		if !activeOnly || v.Enabled {
			ccs = append(ccs, v)
		}
	}

	sort.Slice(ccs, func(i, j int) bool {
		return ccs[i].Name < ccs[j].Name
	})

	return ccs
}

func CountryCodeToName(code string) string {

	switch code {
	case "":
		return code
	case "BQ":
		return "Bonaire, Sint Eustatius and Saba"
	case "SH":
		return "Saint Helena"
	case "XK":
		return "Kosovo"
	case "FX":
		return "France, Metropolitan"
	case "YU":
		return "Yugoslavia"
	}

	country, err := gountriesInstance.FindCountryByAlpha(code)
	if err != nil {
		log.Err(err)
		return code
	}

	return country.Name.Common
}

func CountryCodeToContinent(code string) string {

	country, err := gountriesInstance.FindCountryByAlpha(code)
	if err != nil {
		log.Err(err)
		return code
	}

	for _, v := range Continents {
		if v.Value == country.Continent {
			return v.Key
		}
	}

	return ""
}

//
func CountriesInContinent(continent string) (ret []string) {

	countries := gountriesInstance.FindCountries(gountries.Country{Geo: gountries.Geo{Continent: continent}})
	for _, country := range countries {
		ret = append(ret, country.Alpha2)
	}
	return ret
}

// Value is cents
func FormatPrice(currencyCode steam.CurrencyCode, value int, returnNumber ...bool) string {

	if value == 0 && len(returnNumber) == 0 {
		return "Free"
	}

	unit, _ := currency.ParseISO(string(currencyCode))
	printer := message.NewPrinter(language.AmericanEnglish)
	symbol := printer.Sprint(currency.Symbol(unit.Amount(0.0)))
	return strings.Replace(symbol, "0.00", humanize.FormatFloat("#,###.##", float64(value)/100), 1)
}

var States = map[string]map[string]string{
	"AU": {
		"ACT": "Australian Capital Territory",
		"JBT": "Jervis Bay Territory",
		"NSW": "New South Wales",
		"NT":  "Northern Territory",
		"QLD": "Queensland",
		"SA":  "South Australia",
		"TAS": "Tasmania",
		"VIC": "Victoria",
		"WA":  "Western Australia",
	},
	"CA": {
		"AB": "Alberta",
		"BC": "British Columbia",
		"MB": "Manitoba",
		"NB": "New Brunswick",
		"NL": "Newfoundland and Labrador",
		"NS": "Nova Scotia",
		"NT": "Northwest Territories",
		"NU": "Nunavut",
		"ON": "Ontario",
		"PE": "Prince Edward Island",
		"QC": "Quebec",
		"SK": "Saskatchewan",
		"YT": "Yukon",
	},
	"FR": {},
	"GB": {},
	"NZ": {},
	"PH": {},
	"SI": {},
	"US": {
		"AK": "Alaska",
		"AL": "Alabama",
		"AR": "Arkansas",
		"AS": "American Samoa",
		"AZ": "Arizona",
		"CA": "California",
		"CO": "Colorado",
		"CT": "Connecticut",
		"DC": "District of Columbia",
		"DE": "Delaware",
		"FL": "Florida",
		"FM": "Micronesia",
		"GA": "Georgia",
		"GU": "Guam",
		"HI": "Hawaii",
		"IA": "Iowa",
		"ID": "Idaho",
		"IL": "Illinois",
		"IN": "Indiana",
		"KS": "Kansas",
		"KY": "Kentucky",
		"LA": "Louisiana",
		"MA": "Massachusetts",
		"MD": "Maryland",
		"ME": "Maine",
		"MH": "Marshall Islands",
		"MI": "Michigan",
		"MN": "Minnesota",
		"MO": "Missouri",
		"MP": "Northern Mariana Islands",
		"MS": "Mississippi",
		"MT": "Montana",
		"NC": "North Carolina",
		"ND": "North Dakota",
		"NE": "Nebraska",
		"NH": "New Hampshire",
		"NJ": "New Jersey",
		"NM": "New Mexico",
		"NV": "Nevada",
		"NY": "New York",
		"OH": "Ohio",
		"OK": "Oklahoma",
		"OR": "Oregon",
		"PA": "Pennsylvania",
		"PR": "Puerto Rico",
		"PW": "Palau",
		"RI": "Rhode Island",
		"SC": "South Carolina",
		"SD": "South Dakota",
		"TN": "Tennessee",
		"TX": "Texas",
		"UT": "Utah",
		"VA": "Virginia",
		"VI": "Virgin Islands",
		"VT": "Vermont",
		"WA": "Washington",
		"WI": "Wisconsin",
		"WV": "West Virginia",
		"WY": "Wyoming",
	},
}

// Continents
const (
	CAF = "C-AF"
	CAN = "C-AN"
	CAS = "C-AS"
	CEU = "C-EU"
	CNA = "C-NA"
	CSA = "C-SA"
	COC = "C-OC"
)

// These strings must match the continents in the gountries library
var Continents = []Tuple{
	{Key: CAF, Value: "Africa"},
	{Key: CAN, Value: "Antarctica"},
	{Key: CAS, Value: "Asia"},
	{Key: CEU, Value: "Australia"},
	{Key: CNA, Value: "Europe"},
	{Key: CSA, Value: "North America"},
	{Key: COC, Value: "South America"},
}
