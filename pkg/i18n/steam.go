package i18n

import (
	"sort"
	"strings"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/dustin/go-humanize"
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type ProductCountryCode struct {
	ProductCode  steamapi.ProductCC
	CurrencyCode steamapi.CurrencyCode
	CountryCodes []string // Used to get a currency from an country from an IP
	Name         string
	Symbol       string
	Enabled      bool
}

func (pcc ProductCountryCode) GetFlag() string {
	switch pcc.ProductCode {
	case steamapi.ProductCCEU:
		return "eu"
	case steamapi.ProductCCUK:
		return "gb"
	case steamapi.ProductCCAZ:
		return "cis"
	case steamapi.ProductCCPK:
		return "sasia"
	default:
		return strings.ToLower(pcc.CountryCodes[0])
	}
}

var ProductCountryCodes = map[steamapi.ProductCC]ProductCountryCode{
	steamapi.ProductCCAR: {
		ProductCode:  steamapi.ProductCCAR,
		CountryCodes: []string{"AR"},
		CurrencyCode: steamapi.CurrencyARS,
		Name:         "Argentine Peso",
		Symbol:       "ARS$",
	},
	steamapi.ProductCCAU: {
		ProductCode:  steamapi.ProductCCAU,
		CountryCodes: []string{"AU"},
		CurrencyCode: steamapi.CurrencyAUD,
		Name:         "Australian Dollar",
		Symbol:       "A$",
	},
	steamapi.ProductCCBR: {
		ProductCode:  steamapi.ProductCCBR,
		CountryCodes: []string{"BR"},
		CurrencyCode: steamapi.CurrencyBRL,
		Name:         "Brazilian Real",
		Symbol:       "R$",
	},
	steamapi.ProductCCCA: {
		ProductCode:  steamapi.ProductCCCA,
		CountryCodes: []string{"CA"},
		CurrencyCode: steamapi.CurrencyCAD,
		Name:         "Canadian Dollar",
		Symbol:       "CDN$",
	},
	steamapi.ProductCCCL: {
		ProductCode:  steamapi.ProductCCCL,
		CountryCodes: []string{"CL"},
		CurrencyCode: steamapi.CurrencyCLP,
		Name:         "Chilean Peso",
		Symbol:       "CLP$",
	},
	steamapi.ProductCCCN: {
		ProductCode:  steamapi.ProductCCCN,
		CountryCodes: []string{"CN"},
		CurrencyCode: steamapi.CurrencyCNY,
		Name:         "Chinese Renminbi",
		Symbol:       "¥",
		Enabled:      true,
	},
	steamapi.ProductCCCO: {
		ProductCode:  steamapi.ProductCCCO,
		CountryCodes: []string{"CO"},
		CurrencyCode: steamapi.CurrencyCOP,
		Name:         "Colombian Peso",
		Symbol:       "COL$",
	},
	steamapi.ProductCCCR: {
		ProductCode:  steamapi.ProductCCCR,
		CountryCodes: []string{"CR"},
		CurrencyCode: steamapi.CurrencyCRC,
		Name:         "Costa Rican Colon",
		Symbol:       "₡",
	},
	steamapi.ProductCCEU: { // European Union
		ProductCode:  steamapi.ProductCCEU,
		CountryCodes: []string{"AT", "BE", "BG", "HR", "CY", "CZ", "DK", "EE", "FI", "FR", "DE", "EL", "HU", "IE", "IT", "LV", "LT", "LU", "MT", "NL", "PL", "PT", "RO", "SK", "SI", "ES", "SE"},
		CurrencyCode: steamapi.CurrencyEUR,
		Name:         "Euro",
		Symbol:       "€",
		Enabled:      true,
	},
	steamapi.ProductCCHK: {
		ProductCode:  steamapi.ProductCCHK,
		CountryCodes: []string{"HK"},
		CurrencyCode: steamapi.CurrencyHKD,
		Name:         "Hong Kong Dollar",
		Symbol:       "HK$",
	},
	steamapi.ProductCCIN: {
		ProductCode:  steamapi.ProductCCIN,
		CountryCodes: []string{"IN"},
		CurrencyCode: steamapi.CurrencyINR,
		Name:         "Indian Rupee",
		Symbol:       "₹",
	},
	steamapi.ProductCCID: {
		ProductCode:  steamapi.ProductCCID,
		CountryCodes: []string{"ID"},
		CurrencyCode: steamapi.CurrencyIDR,
		Name:         "Indonesian Rupiah",
		Symbol:       "Rp",
	},
	steamapi.ProductCCIL: {
		ProductCode:  steamapi.ProductCCIL,
		CountryCodes: []string{"IL"},
		CurrencyCode: steamapi.CurrencyILS,
		Name:         "Israeli New Shekel",
		Symbol:       "₪",
	},
	steamapi.ProductCCJP: {
		ProductCode:  steamapi.ProductCCJP,
		CountryCodes: []string{"JP"},
		CurrencyCode: steamapi.CurrencyJPY,
		Name:         "Japanese Yen",
		Symbol:       "¥",
	},
	steamapi.ProductCCKZ: {
		ProductCode:  steamapi.ProductCCKZ,
		CountryCodes: []string{"KZ"},
		CurrencyCode: steamapi.CurrencyKZT,
		Name:         "Kazakhstani Tenge",
		Symbol:       "₸",
	},
	steamapi.ProductCCKW: {
		ProductCode:  steamapi.ProductCCKW,
		CountryCodes: []string{"KW"},
		CurrencyCode: steamapi.CurrencyKWD,
		Name:         "Kuwaiti Dinar",
		Symbol:       "KD",
	},
	steamapi.ProductCCMY: {
		ProductCode:  steamapi.ProductCCMY,
		CountryCodes: []string{"MY"},
		CurrencyCode: steamapi.CurrencyMYR,
		Name:         "Malaysian Ringgit",
		Symbol:       "RM",
	},
	steamapi.ProductCCMX: {
		ProductCode:  steamapi.ProductCCMX,
		CountryCodes: []string{"MX"},
		CurrencyCode: steamapi.CurrencyMXN,
		Name:         "Mexican Peso",
		Symbol:       "Mex$",
	},
	steamapi.ProductCCTW: {
		ProductCode:  steamapi.ProductCCTW,
		CountryCodes: []string{"TW"},
		CurrencyCode: steamapi.CurrencyTWD,
		Name:         "New Taiwan Dollar",
		Symbol:       "NT$",
	},
	steamapi.ProductCCNZ: {
		ProductCode:  steamapi.ProductCCNZ,
		CountryCodes: []string{"NZ"},
		CurrencyCode: steamapi.CurrencyNZD,
		Name:         "New Zealand Dollar",
		Symbol:       "NZ$",
	},
	steamapi.ProductCCNO: {
		ProductCode:  steamapi.ProductCCNO,
		CountryCodes: []string{"NO"},
		CurrencyCode: steamapi.CurrencyNOK,
		Name:         "Norwegian Krone",
		Symbol:       "kr",
	},
	steamapi.ProductCCPE: {
		ProductCode:  steamapi.ProductCCPE,
		CountryCodes: []string{"PE"},
		CurrencyCode: steamapi.CurrencyPEN,
		Name:         "Peruvian Sol",
		Symbol:       "S/",
	},
	steamapi.ProductCCPH: {
		ProductCode:  steamapi.ProductCCPH,
		CountryCodes: []string{"PH"},
		CurrencyCode: steamapi.CurrencyPHP,
		Name:         "Philippine Peso",
		Symbol:       "₱",
	},
	steamapi.ProductCCPL: {
		ProductCode:  steamapi.ProductCCPL,
		CountryCodes: []string{"PL"},
		CurrencyCode: steamapi.CurrencyPLN,
		Name:         "Polish Zloty",
		Symbol:       "zł",
	},
	steamapi.ProductCCUK: {
		ProductCode:  steamapi.ProductCCUK,
		CountryCodes: []string{"GB"},
		CurrencyCode: steamapi.CurrencyGBP,
		Name:         "Pound Sterling",
		Symbol:       "£",
		Enabled:      true,
	},
	steamapi.ProductCCQA: {
		ProductCode:  steamapi.ProductCCQA,
		CountryCodes: []string{"QA"},
		CurrencyCode: steamapi.CurrencyQAR,
		Name:         "Qatari Riyal",
		Symbol:       "QR",
	},
	steamapi.ProductCCRU: {
		ProductCode:  steamapi.ProductCCRU,
		CountryCodes: []string{"RU"},
		CurrencyCode: steamapi.CurrencyRUB,
		Name:         "Russian Ruble",
		Symbol:       "₽",
		Enabled:      true,
	},
	steamapi.ProductCCSA: {
		ProductCode:  steamapi.ProductCCSA,
		CountryCodes: []string{"SA"},
		CurrencyCode: steamapi.CurrencySAR,
		Name:         "Saudi Riyal",
		Symbol:       "SR",
	},
	steamapi.ProductCCSG: {
		ProductCode:  steamapi.ProductCCSG,
		CountryCodes: []string{"SG"},
		CurrencyCode: steamapi.CurrencySGD,
		Name:         "Singapore Dollar",
		Symbol:       "S$",
	},
	steamapi.ProductCCZA: {
		ProductCode:  steamapi.ProductCCZA,
		CountryCodes: []string{"ZA"},
		CurrencyCode: steamapi.CurrencyZAR,
		Name:         "South African Rand",
		Symbol:       "R",
	},
	steamapi.ProductCCKR: {
		ProductCode:  steamapi.ProductCCKR,
		CountryCodes: []string{"KR"},
		CurrencyCode: steamapi.CurrencyKRW,
		Name:         "South Korean Won",
		Symbol:       "₩",
	},
	steamapi.ProductCCCH: {
		ProductCode:  steamapi.ProductCCCH,
		CountryCodes: []string{"CH"},
		CurrencyCode: steamapi.CurrencyCHF,
		Name:         "Swiss Franc",
		Symbol:       "CHF",
	},
	steamapi.ProductCCTH: {
		ProductCode:  steamapi.ProductCCTH,
		CountryCodes: []string{"TH"},
		CurrencyCode: steamapi.CurrencyTHB,
		Name:         "Thai Baht",
		Symbol:       "฿",
	},
	steamapi.ProductCCTR: {
		ProductCode:  steamapi.ProductCCTR,
		CountryCodes: []string{"TR"},
		CurrencyCode: steamapi.CurrencyTRY,
		Name:         "Turkish Lira",
		Symbol:       "₺",
	},
	steamapi.ProductCCUA: {
		ProductCode:  steamapi.ProductCCUA,
		CountryCodes: []string{"UA"},
		CurrencyCode: steamapi.CurrencyUAH,
		Name:         "Ukrainian Hryvnia",
		Symbol:       "₴",
	},
	steamapi.ProductCCAE: {
		ProductCode:  steamapi.ProductCCAE,
		CountryCodes: []string{"AE"},
		CurrencyCode: steamapi.CurrencyAED,
		Name:         "United Arab Emirates Dirham",
		Symbol:       "AED"},
	steamapi.ProductCCUS: {
		ProductCode:  steamapi.ProductCCUS,
		CountryCodes: []string{"US"},
		CurrencyCode: steamapi.CurrencyUSD,
		Name:         "United States Dollar",
		Symbol:       "$",
		Enabled:      true,
	},
	steamapi.ProductCCAZ: { // CIS
		ProductCode:  steamapi.ProductCCAZ,
		CountryCodes: []string{"AM", "AZ", "BY", "GE", "KZ", "KG", "MD", "TJ", "TM", "UZ", "UA"},
		CurrencyCode: steamapi.CurrencyUSD,
		Name:         "United States Dollar (CIS)",
		Symbol:       "$",
	},
	steamapi.ProductCCPK: { // SASIA
		ProductCode:  steamapi.ProductCCPK,
		CountryCodes: []string{"BD", "BT", "NP", "PK", "LK"},
		CurrencyCode: steamapi.CurrencyUSD,
		Name:         "United States Dollar (South Asia)",
		Symbol:       "$",
	},
	steamapi.ProductCCUY: {
		ProductCode:  steamapi.ProductCCUY,
		CountryCodes: []string{"UY"},
		CurrencyCode: steamapi.CurrencyUYU,
		Name:         "Uruguayan Peso",
		Symbol:       "$U",
	},
	steamapi.ProductCCVN: {
		ProductCode:  steamapi.ProductCCVN,
		CountryCodes: []string{"VN"},
		CurrencyCode: steamapi.CurrencyVND,
		Name:         "Vietnamese Dong",
		Symbol:       "₫",
	},
}

func IsValidProdCC(cc steamapi.ProductCC) bool {
	_, ok := ProductCountryCodes[cc]
	return ok
}

func GetProdCC(cc steamapi.ProductCC) ProductCountryCode {

	if cc == "de" {
		cc = steamapi.ProductCCEU
	}

	if val, ok := ProductCountryCodes[cc]; ok && val.Enabled {
		return val
	}

	return ProductCountryCodes[steamapi.ProductCCUS]
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

// Value is cents
func FormatPrice(currencyCode steamapi.CurrencyCode, value int, returnNumber ...bool) string {

	if value == 0 && len(returnNumber) == 0 {
		return "Free"
	}

	unit, _ := currency.ParseISO(string(currencyCode))
	printer := message.NewPrinter(language.AmericanEnglish)
	symbol := printer.Sprint(currency.Symbol(unit.Amount(0.0)))
	return strings.Replace(symbol, "0.00", humanize.FormatFloat("#,###.##", float64(value)/100), 1)
}
