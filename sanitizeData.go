package main

import (
	//"encoding/hex"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
)

func parseJSONData(jsonString []byte) (stockList []Stock) {
	raw := make([]json.RawMessage, 10)
	if err := json.Unmarshal(jsonString, &raw); err != nil {
		log.Fatalf("error %v", err)
	}

	for i := 0; i < len(raw); i += 1 {
		stock := Stock{}
		if err := json.Unmarshal(raw[i], &stock); err != nil {
			fmt.Println("error %v", err)
		}

		stockList = append(stockList, stock)
	}

	return
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
