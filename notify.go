package main

import (
	"fmt"
	"log"
	"net/smtp"
)

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
