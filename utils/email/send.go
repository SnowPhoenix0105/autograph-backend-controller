package email

import (
	"gopkg.in/gomail.v2"
)

func SendHtml(email string, subject string, htmlContent string) error {
	msg := gomail.NewMessage()

	msg.SetHeader("From", globalConfig.SMTP.UserName)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", subject)

	msg.SetBody("text/html", htmlContent)

	dialer := gomail.NewDialer(
		globalConfig.SMTP.Host,
		globalConfig.SMTP.Port,
		globalConfig.SMTP.UserName,
		globalConfig.SMTP.Password)

	if err := dialer.DialAndSend(msg); err != nil {
		return err
	}

	return nil
}
