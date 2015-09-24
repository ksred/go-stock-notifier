package main

import (
	//"encoding/hex"
	"bytes"
	"encoding/json"
	"fmt"
	//"io"
	"io/ioutil"
	"log"
	"net/http"
	//"strings"
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
	type StockSingle struct {
		Symbol           string `json:"symbol"`
		Exchange         string `json:"exchange"`
		Name             string `json:"name"`
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

	type Stocks struct {
		Stock StockSingle
	}

	raw := make([]json.RawMessage, 10)
	if err := json.Unmarshal(jsonString, &raw); err != nil {
		log.Fatalf("error %v", err)
	}

	stockList := make([]Stocks, 0)

	for i := 0; i < len(raw); i += 1 {
		stocks := Stocks{}

		stock := StockSingle{}
		if err := json.Unmarshal(raw[i], &stock); err != nil {
			fmt.Println("error %v", err)
		} else {
			stocks.Stock = stock
		}

		stockList = append(stockList, stocks)
	}
	fmt.Printf("%v\n", stockList)

	for i := range stockList {
		stock := stockList[i].Stock
		fmt.Printf("=====================================\n")
		fmt.Printf("%s\n", stock.Name)
		fmt.Printf("%s: %s\n", stock.Symbol, stock.Exchange)
		fmt.Printf("Change: %s : %s%%\n", stock.Change, stock.PercentageChange)
		fmt.Printf("Open:   %s, Close:   %s\n", stock.Open, stock.Close)
		fmt.Printf("High:   %s, Low:     %s\n", stock.High, stock.Low)
		fmt.Printf("Volume: %s, Average Volume:     %s\n", stock.Volume, stock.AverageVolume)
		fmt.Printf("High 52: %s, Low 52:     %s\n", stock.High52, stock.Low52)
		fmt.Printf("Market Cap: %s\n", stock.MarketCap)
		fmt.Printf("EPS: %s\n", stock.EPS)
		fmt.Printf("Shares: %s\n", stock.Shares)
		fmt.Printf("=====================================\n")
	}

}

func sanitizeBody(source string, body []byte) (bodyResponse []byte) {
	switch source {
	case "google":
		body = body[4 : len(body)-1]
		//body = body[3 : len(body)-1]

		body = bytes.Replace(body, []byte("\\x2F"), []byte("/"), -1)
		body = bytes.Replace(body, []byte("\\x26"), []byte("&"), -1)
		body = bytes.Replace(body, []byte("\\x3B"), []byte(";"), -1)
		body = bytes.Replace(body, []byte("\\x27"), []byte("'"), -1)
	}

	bodyResponse = body

	return

}
