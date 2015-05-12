/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package mail

import (
	"crypto/tls"
	"fmt"

	"github.com/desertbit/bulldozer/settings"

	"gopkg.in/gomail.v1"
)

//#############//
//### Types ###//
//#############//

type Message struct {
	To      []string
	ReplyTo []string
	Subject string
	Body    string

	// Optional: filenames of attachments.
	Attachments []string
}

//##############//
//### Public ###//
//##############//

// Send an e-mail message with the system's noreply e-mail.
func Send(m *Message) error {
	var err error

	// Get the settings pointer.
	s := &settings.Settings

	// Create a new message.
	msg := gomail.NewMessage()
	msg.SetHeader("From", s.MailFrom)
	msg.SetHeader("To", m.To...)
	msg.SetHeader("Subject", m.Subject)
	msg.SetBody("text/html", m.Body)

	if len(m.ReplyTo) > 0 {
		msg.SetHeader("Reply-To", m.ReplyTo...)
	}

	// Add all attachments.
	for _, a := range m.Attachments {
		f, err := gomail.OpenFile(a)
		if err != nil {
			return fmt.Errorf("utils.SendEMail: failed to add attachment: %v", err)
		}
		msg.Attach(f)
	}

	// Create the mailer.
	var mailer *gomail.Mailer
	if s.MailSkipCertificateVerify {
		mailer = gomail.NewMailer(s.MailSMTPHost, s.MailUsername, s.MailPassword, s.MailSMTPPort, gomail.SetTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	} else {
		mailer = gomail.NewMailer(s.MailSMTPHost, s.MailUsername, s.MailPassword, s.MailSMTPPort)
	}

	// Send the email.
	if err = mailer.Send(msg); err != nil {
		return fmt.Errorf("utils.SendEmail: failed to send message: %v", err)
	}

	return nil

}
