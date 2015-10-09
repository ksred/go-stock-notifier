package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"math"
	"net/http"
	"os"
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

func main() {

	configuration := Configuration{}
	loadConfig(&configuration)

	db := loadDatabase(&configuration)

	checkFlags(configuration, db)

	symbolString := convertStocksString(configuration.Symbols)

	// Yahoo: http://chartapi.finance.yahoo.com/instrument/1.0/msft/chartdata;type=quote;ys=2005;yz=4;ts=1234567890/json
	// URL to get detailed company information for a single stock
	// var urlDetailed string = "https://www.google.com/finance?q=JSE%3AIMP&q=JSE%3ANPN&ei=TrUBVomhAsKcUsP5mZAG&output=json"
	// URL to get broad financials for multiple stocks
	var urlStocks string = "https://www.google.com/finance/info?infotype=infoquoteall&q=" + symbolString

	/*
		body := getDataFromURL(urlStocks)

		jsonString := sanitizeBody("google", body)

		stockList := make([]Stock, 0)
		stockList = parseJSONData(jsonString)

		fmt.Println("\t\tOn chosen hours")
		trendingStocks := CalculateTrends(configuration, stockList, db)
		notifyMail := composeMailTemplateTrending(trendingStocks, "trend")
		fmt.Println(len(notifyMail))
		sendMail(configuration, notifyMail)

		return
	*/
	// We check for updates every minute
	//duration, _ := time.ParseDuration(configuration.UpdateInterval)
	fmt.Println("Go finance started")
	go updateAtInterval(60, urlStocks, configuration, db)

	select {} // this will cause the program to run forever
}

func checkFlags(configuration Configuration, db *sql.DB) {
	// Check for any flags
	testFlag := flag.String("test", "", "Test to run")
	symbolFlag := flag.String("symbol", "", "Symbol to run test against")

	flag.Parse()

	// Dereference
	flagParsed := *testFlag
	symbolParsed := *symbolFlag

	switch flagParsed {
	case "trends":
		symbolString := convertStocksString(configuration.Symbols)
		var urlStocks string = "https://www.google.com/finance/info?infotype=infoquoteall&q=" + symbolString
		body := getDataFromURL(urlStocks)

		jsonString := sanitizeBody("google", body)

		stockList := make([]Stock, 0)
		stockList = parseJSONData(jsonString)

		CalculateTrends(configuration, stockList, db)
		os.Exit(0)

		break
	case "trendMail":
		symbolString := convertStocksString(configuration.Symbols)
		var urlStocks string = "https://www.google.com/finance/info?infotype=infoquoteall&q=" + symbolString
		body := getDataFromURL(urlStocks)

		jsonString := sanitizeBody("google", body)

		stockList := make([]Stock, 0)
		stockList = parseJSONData(jsonString)

		trendingStocks := CalculateTrends(configuration, stockList, db)
		notifyMail := composeMailTemplateTrending(trendingStocks, "trend")
		sendMail(configuration, notifyMail)

		os.Exit(0)

		break
	case "update":
		break
	case "stdDev":
		calculateStdDev(configuration, db, symbolParsed, 2)

		os.Exit(0)
		break
	default:
		fmt.Println("Incorrect option chosen: trends, trendMail, update, stdDev")
		os.Exit(0)
		break
	}
}

func updateAtInterval(n time.Duration, urlStocks string, configuration Configuration, db *sql.DB) {

	for _ = range time.Tick(n * time.Second) {
		t := time.Now()
		fmt.Println("BEGIN. Location:", t.Location(), ":Time:", t)
		utc, err := time.LoadLocation(configuration.TimeZone)
		if err != nil {
			fmt.Println("err: ", err.Error())
		}
		hour := t.In(utc).Hour()
		minute := t.In(utc).Minute()
		weekday := t.In(utc).Weekday()

		// Send trending mail
		if weekday != 6 && weekday != 0 {
			if hour == 7 && minute == 0 {
				body := getDataFromURL(urlStocks)

				jsonString := sanitizeBody("google", body)

				stockList := make([]Stock, 0)
				stockList = parseJSONData(jsonString)

				fmt.Println("\t\tTrending")
				// Calculate any trends at end of day
				trendingStocks := CalculateTrends(configuration, stockList, db)
				notifyMail := composeMailTemplateTrending(trendingStocks, "trend")
				sendMail(configuration, notifyMail)
				break
			}
		}

		// This must only be run when the markets are open
		if weekday != 6 && weekday != 0 && hour >= 9 && hour < 17 {
			fmt.Println("\tFalls within operating hours")
			fmt.Println(hour)
			fmt.Println(minute)
			// Save results every 15 minutes
			if math.Mod(float64(minute), 15.) == 0 {
				fmt.Println("\tFalls within 15 minute interval ")
				body := getDataFromURL(urlStocks)

				jsonString := sanitizeBody("google", body)

				stockList := make([]Stock, 0)
				stockList = parseJSONData(jsonString)

				saveToDB(db, stockList, configuration)
				// Mail every X, here is 2 hours
				if minute == 15 {
					switch hour {
					//@TODO Make this dynamic from config
					case 9, 11, 13, 15, 17:
						fmt.Println("\t\tOn chosen hours")
						notifyMail := composeMailTemplate(stockList, "update")
						sendMail(configuration, notifyMail)
						break
					}
				}
			}
		}
		fmt.Println("END. Location:", t.Location(), ":Time:", t)
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

func getDataFromURL(urlStocks string) (body []byte) {
	resp, err := http.Get(urlStocks)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)

	return
}
