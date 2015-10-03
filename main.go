package main

import (
	//"encoding/hex"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	//@TODO Move into separate packge
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

type Configuration struct {
	MailUser       string
	MailPass       string
	MailSMTPServer string
	MailSMTPPort   string
	MailRecipient  string
	MailSender     string
	Symbols        []string
	UpdateInterval string
	TimeZone       string
	MySQLUser      string
	MySQLPass      string
	MySQLHost      string
	MySQLPort      string
	MySQLDB        string
}

type StockSingle struct {
	Symbol           string `json:"t"`
	Exchange         string `json:"e"`
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

func main() {

	configuration := Configuration{}
	loadConfig(&configuration)

	//@TODO Move this into a function/package
	db := loadDatabase(&configuration)

	symbolString := convertStocksString(configuration.Symbols)

	// Yahoo: http://chartapi.finance.yahoo.com/instrument/1.0/msft/chartdata;type=quote;ys=2005;yz=4;ts=1234567890/json
	// URL to get detailed company information for a single stock
	// var urlDetailed string = "https://www.google.com/finance?q=JSE%3AIMP&q=JSE%3ANPN&ei=TrUBVomhAsKcUsP5mZAG&output=json"
	// URL to get broad financials for multiple stocks
	var urlStocks string = "https://www.google.com/finance/info?infotype=infoquoteall&q=" + symbolString

	// We check for updates every minute
	//duration, _ := time.ParseDuration(configuration.UpdateInterval)
	go updateAtInterval(60, urlStocks, configuration, db) // very useful for interval polling

	select {} // this will cause the program to run forever
}

func updateAtInterval(n time.Duration, urlStocks string, configuration Configuration, db *sql.DB) {

	for _ = range time.Tick(n * time.Second) {
		t := time.Now()
		fmt.Println("Location:", t.Location(), ":Time:", t)
		utc, err := time.LoadLocation(configuration.TimeZone)
		if err != nil {
			fmt.Println("err: ", err.Error())
		}
		hour := t.In(utc).Hour()
		minute := t.In(utc).Minute()

		// This must only be run when the markets are open
		switch {
		//@TODO Make this dynamic from config
		case hour == 9 && minute == 00:
		case hour == 11 && minute == 00:
		case hour == 13 && minute == 00:
		case hour == 15 && minute == 00:
		case hour == 17 && minute == 00:
			body := getDataFromURL(urlStocks)

			jsonString := sanitizeBody("google", body)

			stockList := make([]Stocks, 0)
			stockList = parseJSONData(jsonString)

			notifyMail := composeMailString(stockList)

			sendMail(configuration, notifyMail)

			saveToDB(db, stockList, configuration)
		}
	}
}

func loadConfig(configuration *Configuration) {
	// Get config
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
}

func convertStocksString(symbols []string) (symbolString string) {
	for i := range symbols {
		symbol := []byte(symbols[i])
		symbol = bytes.Replace(symbol, []byte(":"), []byte("%3A"), -1)
		symbolString += string(symbol)
		if i < len(symbols)-1 {
			symbolString += ","
		}
	}

	return
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

func getDataFromURL(urlStocks string) (body []byte) {
	resp, err := http.Get(urlStocks)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)

	return
}

func parseJSONData(jsonString []byte) (stockList []Stocks) {
	raw := make([]json.RawMessage, 10)
	if err := json.Unmarshal(jsonString, &raw); err != nil {
		log.Fatalf("error %v", err)
	}

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

	return
}

func composeMailString(stockList []Stocks) (notifyMail string) {
	for i := range stockList {
		stock := stockList[i].Stock
		notifyMail += fmt.Sprintf("=====================================\n")
		notifyMail += fmt.Sprintf("%s\n", stock.Name)
		notifyMail += fmt.Sprintf("%s: %s\n", stock.Symbol, stock.Exchange)
		notifyMail += fmt.Sprintf("Change: %s : %s%%\n", stock.Change, stock.PercentageChange)
		notifyMail += fmt.Sprintf("Open:   %s, Close:   %s\n", stock.Open, stock.Close)
		notifyMail += fmt.Sprintf("High:   %s, Low:     %s\n", stock.High, stock.Low)
		notifyMail += fmt.Sprintf("Volume: %s, Average Volume:     %s\n", stock.Volume, stock.AverageVolume)
		notifyMail += fmt.Sprintf("High 52: %s, Low 52:     %s\n", stock.High52, stock.Low52)
		notifyMail += fmt.Sprintf("Market Cap: %s\n", stock.MarketCap)
		notifyMail += fmt.Sprintf("EPS: %s\n", stock.EPS)
		notifyMail += fmt.Sprintf("Shares: %s\n", stock.Shares)
		notifyMail += fmt.Sprintf("=====================================\n")
	}

	return
}

func sendMail(configuration Configuration, notifyMail string) {
	// Send email
	// Set up authentication information.
	auth := smtp.PlainAuth("", configuration.MailUser, configuration.MailPass, configuration.MailSMTPServer)

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	to := []string{configuration.MailRecipient}
	msg := []byte("To: " + configuration.MailRecipient + "\r\n" +
		"Subject: Quote update!\r\n" +
		"\r\n" +
		notifyMail +
		"\r\n")

	err := smtp.SendMail(configuration.MailSMTPServer+":"+configuration.MailSMTPPort, auth, configuration.MailSender, to, msg)
	//err = smtp.SendMail("mail.messagingengine.com:587", auth, "ksred@fastmail.fm", []string{"kyle@ksred.me"}, msg)
	if err != nil {
		log.Fatal(err)
	}
}

func loadDatabase(configuration *Configuration) (db *sql.DB) {
	db, err := sql.Open("mysql", configuration.MySQLUser+":"+configuration.MySQLPass+"@tcp("+configuration.MySQLHost+":"+configuration.MySQLPort+")/"+configuration.MySQLDB)
	if err != nil {
		fmt.Println("Could not connect to database")
		return
	}
	defer db.Close()

	// Test connection with ping
	err = db.Ping()
	if err != nil {
		fmt.Println("Ping error: " + err.Error()) // proper error handling instead of panic in your app
		return
	}

	return
}

func saveToDB(db *sql.DB, stockList []Stocks, configuration Configuration) {
	db, err := sql.Open("mysql", configuration.MySQLUser+":"+configuration.MySQLPass+"@tcp("+configuration.MySQLHost+":"+configuration.MySQLPort+")/"+configuration.MySQLDB)
	if err != nil {
		fmt.Println("Could not connect to database")
		return
	}

	for i := range stockList {
		//@TODO Save results to database
		stock := stockList[i].Stock

		// Prepare statement for inserting data
		insertStatement := "INSERT INTO st_data (`symbol`, `exchange`, `name`, `change`, `close`, `percentageChange`, `open`, `high`, `low`, `volume` , `avgVolume`, `high52` , `low52`, `marketCap`, `eps`, `shares`, `time`, `day`, `month`, `year`) "
		insertStatement += "VALUES( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )"
		stmtIns, err := db.Prepare(insertStatement)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		defer stmtIns.Close() // Close the statement when we leave main() / the program terminates

		// Convert variables
		sqlChange, _ := strconv.ParseFloat(strings.Replace(stock.Change, ",", "", -1), 64)
		sqlClose, _ := strconv.ParseFloat(strings.Replace(stock.Close, ",", "", -1), 64)
		sqlPercChange, _ := strconv.ParseFloat(stock.PercentageChange, 64)
		sqlOpen, _ := strconv.ParseFloat(strings.Replace(stock.Open, ",", "", -1), 64)
		sqlHigh, _ := strconv.ParseFloat(strings.Replace(stock.High, ",", "", -1), 64)
		sqlLow, _ := strconv.ParseFloat(strings.Replace(stock.Low, ",", "", -1), 64)
		sqlHigh52, _ := strconv.ParseFloat(strings.Replace(stock.High52, ",", "", -1), 64)
		sqlLow52, _ := strconv.ParseFloat(strings.Replace(stock.Low52, ",", "", -1), 64)
		sqlEps, _ := strconv.ParseFloat(stock.EPS, 64)

		// Some contain letters that need to be converted
		sqlVolume := convertLetterToDigits(stock.Volume)
		sqlAvgVolume := convertLetterToDigits(stock.AverageVolume)
		sqlMarketCap := convertLetterToDigits(stock.MarketCap)
		sqlShares := convertLetterToDigits(stock.Shares)

		t := time.Now()
		utc, err := time.LoadLocation(configuration.TimeZone)
		if err != nil {
			fmt.Println("err: ", err.Error())
		}
		sqlTime := int32(t.Unix())
		sqlDay := t.In(utc).Day()
		sqlMonth := t.In(utc).Month()
		sqlYear := t.In(utc).Year()

		_, err = stmtIns.Exec(stock.Name, stock.Symbol, stock.Exchange, sqlChange, sqlClose,
			sqlPercChange, sqlOpen, sqlHigh, sqlLow, sqlVolume, sqlAvgVolume,
			sqlHigh52, sqlLow52, sqlMarketCap, sqlEps, sqlShares,
			sqlTime, sqlDay, sqlMonth, sqlYear)

		if err != nil {
			fmt.Println("Could not save results: " + err.Error())
		}
	}
	defer db.Close()
}

func convertLetterToDigits(withLetter string) (withoutLetter float64) {
	// Clear , from string
	withLetter = strings.Replace(withLetter, ",", "", -1)

	// First get multiplier
	multiplier := 1.
	switch {
	case strings.Contains(withLetter, "M"):
		multiplier = 1000000.
		break
	case strings.Contains(withLetter, "B"):
		multiplier = 1000000000.
		break
	}

	// Remove the letters
	withLetter = strings.Replace(withLetter, "M", "", -1)
	withLetter = strings.Replace(withLetter, "B", "", -1)

	// Convert to float
	withoutLetter, _ = strconv.ParseFloat(withLetter, 64)

	// Add multiplier
	withoutLetter = withoutLetter * multiplier

	return
}
