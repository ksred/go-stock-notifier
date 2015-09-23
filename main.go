package main

import (
	//"encoding/hex"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	//"strconv"
)

type StockData struct {
	symbol   string
	exchange string
}

func main() {
	// Yahoo: http://chartapi.finance.yahoo.com/instrument/1.0/msft/chartdata;type=quote;ys=2005;yz=4;ts=1234567890/json
	var urlIMP string = "https://www.google.com/finance?q=JSE%3AIMP&ei=TrUBVomhAsKcUsP5mZAG&output=json"

	resp, err := http.Get(urlIMP)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	/*
		body, err := ioutil.ReadFile("/Users/ksred/golang/gobank/stock-notifier/data.json")
		if err != nil {
			fmt.Println(err)
		}
	*/

	jsonString := sanitizeBody("google", body)

	//fmt.Println(jsonString)

	// Do some json
	type Stock struct {
		Symbol           string `json:"symbol"`
		Exchange         string `json:"exchange"`
		Change           string `json:"c"`
		Close            string `json:"l"`
		PercentageChange string `json:"cp"`
		Open             string `json:"op"`
		High             string `json:"hi"`
		Low              string `json:"lo"`
		Volume           string `json:"vo"`
		AverageVolume    string `json:"avvo"`
		High52           string `json:"hi52"`
		Low52            string `json:"lo52"`
		MarketCap        string `json:"mc"`
		EPS              string `json:"eps"`
		Shares           string `json:"shares"`
	}

	dec := json.NewDecoder(strings.NewReader(jsonString))
	for {
		var m Stock
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("=====================================\n")
		fmt.Printf("%s: %s\n", m.Symbol, m.Exchange)
		fmt.Printf("Change: %s - %s\n", m.Change, m.PercentageChange)
		fmt.Printf("Open:   %s, Close:   %s\n", m.Open, m.Close)
		fmt.Printf("High:   %s, Low:     %s\n", m.High, m.Low)
		fmt.Printf("Volume: %s, Average Volume:     %s\n", m.Volume, m.AverageVolume)
		fmt.Printf("High 52: %s, Low 52:     %s\n", m.High52, m.Low52)
		fmt.Printf("Market Cap: %s\n", m.MarketCap)
		fmt.Printf("EPS: %s\n", m.EPS)
		fmt.Printf("Shares: %s\n", m.Shares)
		fmt.Printf("=====================================\n")
	}

}

func sanitizeBody(source string, body []byte) (bodyResponse string) {
	switch source {
	case "google":
		body = body[6 : len(body)-2]

		body = bytes.Replace(body, []byte("\\x2F"), []byte("/"), -1)
		body = bytes.Replace(body, []byte("\\x26"), []byte("&"), -1)
		body = bytes.Replace(body, []byte("\\x3B"), []byte(";"), -1)
		body = bytes.Replace(body, []byte("\\x27"), []byte("'"), -1)
	}

	bodyResponse = string(body)

	return

}
