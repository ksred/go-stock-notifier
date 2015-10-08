// Package analysis contains functions for analysis of stock data
package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

type Stock struct {
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

type TrendingStock struct {
	*Stock
	TrendingDirection string
	TrendingStrength  int
}

//@TODO We can use sorting to show the top movers etc
// http://nerdyworm.com/blog/2013/05/15/sorting-a-slice-of-structs-in-go/

func CalculateTrends(configuration Configuration, stockList []Stock, db *sql.DB) (trendingStocks []TrendingStock) {
	db, err := sql.Open("mysql", configuration.MySQLUser+":"+configuration.MySQLPass+"@tcp("+configuration.MySQLHost+":"+configuration.MySQLPort+")/"+configuration.MySQLDB)
	if err != nil {
		fmt.Println("Could not connect to database")
		return
	}

	fmt.Println("\t\t\tChecking for trends")
	trendingStocks = make([]TrendingStock, 0)
	for i := range stockList {
		stock := stockList[i]

		// Prepare statement for inserting data
		rows, err := db.Query("SELECT `close`, `avgVolume` FROM `st_data` WHERE `symbol` = ? GROUP BY `day` LIMIT 3", stock.Symbol)
		//rows, err := db.Query("SELECT `close`, `volume` FROM `st_data` WHERE `symbol` = ? LIMIT 3", stock.Symbol)
		if err != nil {
			fmt.Println("Error with select query: " + err.Error())
		}
		defer rows.Close()

		allCloses := make([]float64, 0)
		allVolumes := make([]float64, 0)
		count := 0
		for rows.Next() {
			var stockClose float64
			var stockVolume float64
			if err := rows.Scan(&stockClose, &stockVolume); err != nil {
				log.Fatal(err)
			}
			allCloses = append(allCloses, stockClose)
			allVolumes = append(allVolumes, stockVolume)
			count++
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		if count == 3 {
			if doTrendCalculation(allCloses, allVolumes, "up") {
				fmt.Printf("\t\t\tTrend UP for %s\n", stock.Symbol)

				trendingStock := TrendingStock{&stock, "up", 0}
				trendingStocks = append(trendingStocks, trendingStock)
			} else if doTrendCalculation(allCloses, allVolumes, "down") {
				fmt.Printf("\t\t\tTrend DOWN for %s\n", stock.Symbol)

				trendingStock := TrendingStock{&stock, "down", 0}
				trendingStocks = append(trendingStocks, trendingStock)
			}
		}

	}
	defer db.Close()

	return
}

func doTrendCalculation(closes []float64, volumes []float64, trendType string) (trending bool) {
	/*@TODO
	  Currently a simple analysis is done on daily stock data. This analysis is to identify trending stocks, with a trend being identified by:
	  - A price increase (or decrease) each day for three days
	  - A volume increase (or decrease) over two of the three days
	*/
	fmt.Printf("\t\t\t\tChecking trends with data: price: %f, %f, %f and volume: %f, %f, %f\n", closes[0], closes[1], closes[2], volumes[0], volumes[1], volumes[2])
	switch trendType {
	case "up":
		if closes[2] > closes[1] && closes[1] > closes[0] && (volumes[2] > volumes[0] || volumes[1] > volumes[1]) {
			return true
		}
		break
	case "down":
		if closes[2] < closes[1] && closes[1] < closes[0] && (volumes[2] > volumes[0] || volumes[1] > volumes[0]) {
			return true
		}
		break
	}

	return false
}
