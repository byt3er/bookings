package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/byt3er/bookings/internals/models"
	mail "github.com/xhit/go-simple-mail/v2"
)

func listenForMail() {
	// listen all the time for incoming data
	go func() {
		for {
			msg := <-app.MailChan
			sendMsg(msg)
		}
	}()
}

func sendMsg(m models.MailData) {
	server := mail.NewSMTPClient()
	server.Host = "localhost"
	server.Port = 1025 // real mail server use 587 or 465
	//don't keep the connection to mail server all the time
	// only make connection when I tell you to send email
	server.KeepAlive = false
	// If you can't connect in 10 second just give up
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	// In Production (you need to specify those things)
	// .Username
	// .Password
	// .Encryption

	// Client
	client, err := server.Connect()
	if err != nil {
		errorLog.Println(err)
	}

	// construct email message in a format that our client understands

	// create a new empty message
	email := mail.NewMSG()
	// set tht :from, :to, :subject
	email.SetFrom(m.From).AddTo(m.To).SetSubject(m.Subject)
	// set the email :body
	if m.Template == "" {
		email.SetBody(mail.TextHTML, m.Content)
	} else {
		// read from the disk
		data, err := ioutil.ReadFile(fmt.Sprintf("./email-template/%s", m.Template))
		if err != nil {
			app.ErrorLog.Println(err)
		}
		// template in memory
		mailTemplate := string(data)
		//  subsitude the placeholder
		msgToSend := strings.Replace(mailTemplate, "[%body%]", m.Content, 1)
		email.SetBody(mail.TextHTML, msgToSend)
	}

	// send the email message
	err = email.Send(client)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Email sent!")
	}

}
