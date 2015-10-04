package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
)

type MailTemplate struct {
	Title  string
	Stocks []StockSingle
}

func composeMailTemplate(stockList []Stocks, mailType string) (notifyMail string) {
	// https://jan.newmarch.name/go/template/chapter-template.html
	var templateString bytes.Buffer
	// Massage data
	allStocks := make([]StockSingle, 0)
	for i := range stockList {
		stock := stockList[i].Stock
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

func composeMailString(stockList []Stocks, mailType string) (notifyMail string) {
	switch mailType {
	case "update":
		notifyMail = "Stock Update\n\n"
		break
	case "trend":
		notifyMail = "TRENDING STOCKS\n\n"
		break
	}

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
