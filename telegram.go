// Telegram functions
package main

import (
	"database/sql"
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strconv"
	"strings"
)

func sendTelegramBotMessage(message string, configuration Configuration, replyId int) {
	bot, err := tgbotapi.NewBotAPI(configuration.TelegramBotApi)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	fmt.Printf("Authorized on account %s", bot.Self.UserName)
	botId, err := strconv.Atoi(configuration.TelegramBotID)
	if err != nil {
		fmt.Println("Could not convert telegram bot id")
		return
	}

	msg := tgbotapi.NewMessage(botId, message)
	if replyId != 0 {
		msg.ReplyToMessageID = replyId
	}

	bot.Send(msg)

	/*
	 */
}

func startTelegramBot(configuration Configuration) {
	bot, err := tgbotapi.NewBotAPI(configuration.TelegramBotApi)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	fmt.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {

		// Parse responses
		response := strings.Split(update.Message.Text, " ")

		if response[0] == "" {
			sendTelegramBotMessage("Please enter in a command:\nstock [exchange] [symbol], trends", configuration, update.Message.MessageID)
			return
		}

		switch strings.ToLower(response[0]) {
		case "stock":
			processStockBotCommand(response, configuration, update.Message.MessageID)
			break
		case "trends":
			// @TODO Abstract this
			db := loadDatabase(&configuration)
			symbolString := convertStocksString(configuration.Symbols)
			var urlStocks string = "https://www.google.com/finance/info?infotype=infoquoteall&q=" + symbolString
			body := getDataFromURL(urlStocks)

			jsonString := sanitizeBody("google", body)

			stockList := make([]Stock, 0)
			stockList = parseJSONData(jsonString)

			trendingStocks := CalculateTrends(configuration, stockList, db, "day", 3)
			notifyTelegramTrends(trendingStocks, configuration)
			break
		}

	}
}

func processStockBotCommand(response []string, configuration Configuration, replyId int) {
	// Check if all fields are valid
	if len(response) < 3 {
		sendTelegramBotMessage("Not enough commands! \nUsage: stock [exchange] [symbol]", configuration, replyId)
	}

	db, err := sql.Open("mysql", configuration.MySQLUser+":"+configuration.MySQLPass+"@tcp("+configuration.MySQLHost+":"+configuration.MySQLPort+")/"+configuration.MySQLDB)
	if err != nil {
		fmt.Println("Could not connect to database")
		return
	}
	rows, err := db.Query("SELECT `close`, `avgVolume`, `percentageChange` FROM `st_data` WHERE `exchange` = ? AND `symbol` = ? ORDER BY `id` DESC LIMIT ?", response[1], response[2], 1)

	count := 0

	var stockClose float64
	var stockVolume float64
	var stockChange float64
	for rows.Next() {
		if err := rows.Scan(&stockClose, &stockVolume, &stockChange); err != nil {
			log.Fatal(err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	if count == 0 {
		sendTelegramBotMessage("Stock not found", configuration, replyId)
	}

	message := fmt.Sprintf("%s:%s\nClose: %v\nVolume: %v\nChange: %v%%", response[1], response[2], stockClose, stockVolume, stockChange)
	sendTelegramBotMessage(message, configuration, replyId)
}
