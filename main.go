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
	TelegramBotApi string
	TelegramBotID  string
}

const (
	SYMBOL_INTERVAL = 50
)

func main() {

	configuration := Configuration{}
	loadConfig(&configuration)

	db := loadDatabase(&configuration)

	checkFlags(configuration, db)

	// Start Telegram bot
	go startTelegramBot(configuration)

	// Do a loop over symbols
	interval := true
	count := 0
	for interval {

		start := count * SYMBOL_INTERVAL
		end := (count + 1) * SYMBOL_INTERVAL

		if end > len(configuration.Symbols) {
			end = len(configuration.Symbols)
			interval = false
		}

		symbolSlice := configuration.Symbols[start:end]
		symbolString := convertStocksString(symbolSlice)

		// Yahoo: http://chartapi.finance.yahoo.com/instrument/1.0/msft/chartdata;type=quote;ys=2005;yz=4;ts=1234567890/json
		// URL to get detailed company information for a single stock
		// var urlDetailed string = "https://www.google.com/finance?q=JSE%3AIMP&q=JSE%3ANPN&ei=TrUBVomhAsKcUsP5mZAG&output=json"
		// URL to get broad financials for multiple stocks
		var urlStocks string = "https://www.google.com/finance/info?infotype=infoquoteall&q=" + symbolString

		// We check for updates every minute
		//duration, _ := time.ParseDuration(configuration.UpdateInterval)
		fmt.Printf("Go finance started, slice %d\n", count)
		go updateAtInterval(60, urlStocks, configuration, db)

		count++
	}

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
		interval := true
		count := 0
		for interval {

			start := count * SYMBOL_INTERVAL
			end := (count + 1) * SYMBOL_INTERVAL

			if end > len(configuration.Symbols) {
				end = len(configuration.Symbols)
				interval = false
			}

			symbolSlice := configuration.Symbols[start:end]
			symbolString := convertStocksString(symbolSlice)
			var urlStocks string = "https://www.google.com/finance/info?infotype=infoquoteall&q=" + symbolString
			body := getDataFromURL(urlStocks)

			jsonString := sanitizeBody("google", body)

			stockList := make([]Stock, 0)
			stockList = parseJSONData(jsonString)

			CalculateTrends(configuration, stockList, db, "day", 3)

			count++
		}
		os.Exit(0)

		break
	case "trendMail":
		interval := true
		count := 0
		for interval {

			start := count * SYMBOL_INTERVAL
			end := (count + 1) * SYMBOL_INTERVAL

			if end > len(configuration.Symbols) {
				end = len(configuration.Symbols)
				interval = false
			}

			symbolSlice := configuration.Symbols[start:end]
			symbolString := convertStocksString(symbolSlice)

			var urlStocks string = "https://www.google.com/finance/info?infotype=infoquoteall&q=" + symbolString
			body := getDataFromURL(urlStocks)

			jsonString := sanitizeBody("google", body)
			fmt.Println(jsonString)

			stockList := make([]Stock, 0)
			stockList = parseJSONData(jsonString)

			trendingStocks := CalculateTrends(configuration, stockList, db, "day", 3)
			if len(trendingStocks) != 0 {
				notifyMail := composeMailTemplateTrending(trendingStocks, "trend")
				sendMail(configuration, notifyMail)
			}

			count++
		}

		os.Exit(0)

		break
	case "trendMailHourly":
		interval := true
		count := 0
		for interval {

			start := count * SYMBOL_INTERVAL
			end := (count + 1) * SYMBOL_INTERVAL

			if end > len(configuration.Symbols) {
				end = len(configuration.Symbols)
				interval = false
			}

			symbolSlice := configuration.Symbols[start:end]
			symbolString := convertStocksString(symbolSlice)
			var urlStocks string = "https://www.google.com/finance/info?infotype=infoquoteall&q=" + symbolString
			body := getDataFromURL(urlStocks)

			jsonString := sanitizeBody("google", body)

			stockList := make([]Stock, 0)
			stockList = parseJSONData(jsonString)

			trendingStocks := CalculateTrends(configuration, stockList, db, "hour", 3)
			if len(trendingStocks) != 0 {
				notifyMail := composeMailTemplateTrending(trendingStocks, "trend")
				sendMail(configuration, notifyMail)
			}

			count++
		}

		os.Exit(0)

		break
	case "update":
		interval := true
		count := 0
		for interval {

			start := count * SYMBOL_INTERVAL
			end := (count + 1) * SYMBOL_INTERVAL

			if end > len(configuration.Symbols) {
				end = len(configuration.Symbols)
				interval = false
			}

			symbolSlice := configuration.Symbols[start:end]
			symbolString := convertStocksString(symbolSlice)
			var urlStocks string = "https://www.google.com/finance/info?infotype=infoquoteall&q=" + symbolString
			body := getDataFromURL(urlStocks)

			jsonString := sanitizeBody("google", body)

			stockList := make([]Stock, 0)
			stockList = parseJSONData(jsonString)

			fmt.Println("\t\tOn chosen hours")
			notifyMail := composeMailTemplate(stockList, "update")
			sendMail(configuration, notifyMail)

			count++
		}

		os.Exit(0)

		break
	case "stdDev":
		calculateStdDev(configuration, db, symbolParsed, 2)

		os.Exit(0)
		break
	case "trendBot":
		symbolString := convertStocksString(configuration.Symbols)
		var urlStocks string = "https://www.google.com/finance/info?infotype=infoquoteall&q=" + symbolString
		body := getDataFromURL(urlStocks)

		jsonString := sanitizeBody("google", body)

		stockList := make([]Stock, 0)
		stockList = parseJSONData(jsonString)

		trendingStocks := CalculateTrends(configuration, stockList, db, "day", 3)
		notifyTelegramTrends(trendingStocks, configuration)

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
			return
		}
		hour := t.In(utc).Hour()
		minute := t.In(utc).Minute()
		weekday := t.In(utc).Weekday()

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
						//sendMail(configuration, notifyMail)
						fmt.Println(len(notifyMail))
						break
					}
				}
			}
		}

		// Get close
		if weekday != 6 && weekday != 0 && hour == 17 && minute == 15 {
			body := getDataFromURL(urlStocks)

			jsonString := sanitizeBody("google", body)

			stockList := make([]Stock, 0)
			stockList = parseJSONData(jsonString)

			saveToDB(db, stockList, configuration)

			// Send trending update
			trendingStocks := CalculateTrends(configuration, stockList, db, "day", 3)
			// Send to telegram
			notifyTelegramTrends(trendingStocks, configuration)
			if len(trendingStocks) != 0 {
				notifyMail := composeMailTemplateTrending(trendingStocks, "trend")
				sendMail(configuration, notifyMail)
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
		return
	}
}

func getDataFromURL(urlStocks string) (body []byte) {
	resp, err := http.Get(urlStocks)
	if err != nil {
		// handle error
		return
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)

	return
}
