package utils

import (
	"gopkg.in/gomail.v2"
)

type EmailSender struct {
	Dialer *gomail.Dialer
	User   string
}

func (esr *EmailSender) InitGomail(Host string, Port int, User, Password string) {
	esr.Dialer = gomail.NewDialer(Host, Port, User, Password)
	esr.User = User
}

func (esr *EmailSender) SendEmail(To, Msg, Subject string) error {
	//Subject like xxx验证码
	//To : the email receiver
	m := gomail.NewMessage()
	m.SetHeader("From", esr.User)
	m.SetHeader("To", To)
	m.SetHeader("Subject", Subject)
	m.SetBody("text/plain", Msg)
	return esr.Dialer.DialAndSend(m)
}
