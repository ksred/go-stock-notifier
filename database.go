package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"strings"
	"time"
)

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
		insertStatement := "INSERT INTO st_data (`symbol`, `exchange`, `name`, `change`, `close`, `percentageChange`, `open`, `high`, `low`, `volume` , `avgVolume`, `high52` , `low52`, `marketCap`, `eps`, `shares`, `time`, `minute`, `hour`, `day`, `month`, `year`) "
		insertStatement += "VALUES( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )"
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
		sqlMinute := t.In(utc).Minute()
		sqlHour := t.In(utc).Hour()
		sqlDay := t.In(utc).Day()
		sqlMonth := t.In(utc).Month()
		sqlYear := t.In(utc).Year()

		_, err = stmtIns.Exec(stock.Symbol, stock.Exchange, stock.Name, sqlChange, sqlClose,
			sqlPercChange, sqlOpen, sqlHigh, sqlLow, sqlVolume, sqlAvgVolume,
			sqlHigh52, sqlLow52, sqlMarketCap, sqlEps, sqlShares,
			sqlTime, sqlMinute, sqlHour, sqlDay, sqlMonth, sqlYear)

		if err != nil {
			fmt.Println("Could not save results: " + err.Error())
		}
	}
	defer db.Close()
}
