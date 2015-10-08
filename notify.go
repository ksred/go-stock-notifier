package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"sort"
	"strconv"
)

type MailTemplate struct {
	Title  string
	Stocks []Stock
}

type MailTemplateTrending struct {
	Title  string
	Stocks []TrendingStock
}

type Stocks []Stock
type TrendingStocks []TrendingStock

func (slice Stocks) Len() int {
	return len(slice)
}

func (slice Stocks) Less(i, j int) bool {
	val1, _ := strconv.ParseFloat(slice[i].PercentageChange, 64)
	val2, _ := strconv.ParseFloat(slice[j].PercentageChange, 64)
	return val1 < val2
}

func (slice Stocks) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// Trending stocks sort
func (slice TrendingStocks) Len() int {
	return len(slice)
}

func (slice TrendingStocks) Less(i, j int) bool {
	val1, _ := strconv.ParseFloat(slice[i].PercentageChange, 64)
	val2, _ := strconv.ParseFloat(slice[j].PercentageChange, 64)
	return val1 < val2
}

func (slice TrendingStocks) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func composeMailTemplateTrending(stockList []TrendingStock, mailType string) (notifyMail string) {
	// Order by most gained to most lost
	// @TODO Change template to show "top gainers" and "top losers"
	displayStocks := TrendingStocks{}

	displayStocks = stockList
	sort.Sort(sort.Reverse(displayStocks))

	// https://jan.newmarch.name/go/template/chapter-template.html
	var templateString bytes.Buffer
	// Massage data
	allStocks := make([]TrendingStock, 0)
	for i := range displayStocks {
		stock := displayStocks[i]
		allStocks = append(allStocks, stock)
	}

	mailTpl := MailTemplateTrending{
		Stocks: allStocks,
	}

	switch mailType {
	case "update":
		mailTpl.Title = "Stock update"
		t, err := template.ParseFiles("tpl/notification.html")
		if err != nil {
			fmt.Println("template parse error: ", err)
			return
		}
		err = t.Execute(&templateString, mailTpl)
		if err != nil {
			fmt.Println("template executing error: ", err)
			return
		}
		break
	case "trend":
		mailTpl.Title = "Trends update"
		t, err := template.ParseFiles("tpl/trending.html")
		if err != nil {
			fmt.Println("template parse error: ", err)
			return
		}
		err = t.Execute(&templateString, mailTpl)
		if err != nil {
			fmt.Println("template executing error: ", err)
			return
		}
		break
	}

	notifyMail = templateString.String()

	return
}

func composeMailTemplate(stockList []Stock, mailType string) (notifyMail string) {
	// Order by most gained to most lost
	// @TODO Change template to show "top gainers" and "top losers"
	displayStocks := Stocks{}

	displayStocks = stockList
	sort.Sort(sort.Reverse(displayStocks))

	// https://jan.newmarch.name/go/template/chapter-template.html
	var templateString bytes.Buffer
	// Massage data
	allStocks := make([]Stock, 0)
	for i := range displayStocks {
		stock := displayStocks[i]
		allStocks = append(allStocks, stock)
	}

	mailTpl := MailTemplate{
		Stocks: allStocks,
	}

	switch mailType {
	case "update":
		mailTpl.Title = "Stock update"
		t, err := template.ParseFiles("tpl/notification.html")
		if err != nil {
			fmt.Println("template parse error: ", err)
			return
		}
		err = t.Execute(&templateString, mailTpl)
		if err != nil {
			fmt.Println("template executing error: ", err)
			return
		}
		break
	case "trend":
		mailTpl.Title = "Trends update"
		t, err := template.ParseFiles("tpl/notification.html")
		if err != nil {
			fmt.Println("template parse error: ", err)
			return
		}
		err = t.Execute(&templateString, mailTpl)
		if err != nil {
			fmt.Println("template executing error: ", err)
			return
		}
		break
	}

	notifyMail = templateString.String()

	return
}

func composeMailString(stockList []Stock, mailType string) (notifyMail string) {
	switch mailType {
	case "update":
		notifyMail = "Stock Update\n\n"
		break
	case "trend":
		notifyMail = "TRENDING STOCKS\n\n"
		break
	}

	for i := range stockList {
		stock := stockList[i]
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
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	to := []string{configuration.MailRecipient}
	msg := []byte("To: " + configuration.MailRecipient + "\r\n" +
		"Subject: Quote update!\r\n" +
		mime + "\r\n" +
		"\r\n" +
		notifyMail +
		"\r\n")

	err := smtp.SendMail(configuration.MailSMTPServer+":"+configuration.MailSMTPPort, auth, configuration.MailSender, to, msg)
	//err = smtp.SendMail("mail.messagingengine.com:587", auth, "ksred@fastmail.fm", []string{"kyle@ksred.me"}, msg)
	if err != nil {
		log.Fatal(err)
	}
}
