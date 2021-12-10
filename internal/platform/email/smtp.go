package email

import (
	"fmt"
	"bytes"
	"gopkg.in/gomail.v2"
	"html/template"
	"os"
	"strconv"

	param "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/interfaces/requestparams"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
)

// SMTPGoMail SMTPGoMail
type SMTPGoMail struct {
	hostMail         string
	hostMailPassword string
	smtpAdrr         string
	smtpPort         int
}

// InitSmtp
func (smtp *SMTPGoMail) InitSmtp(email string, password string) {
	smtp.hostMail = email
	smtp.hostMailPassword = password
	smtp.smtpAdrr = os.Getenv("MAIL_HOST")
	smtp.smtpPort, _ = strconv.Atoi(os.Getenv("MAIL_PORT"))
}

func (smtp *SMTPGoMail) SendMail(subject string, sampleData *param.SampleData, templateFile string) error {

	dryrun, _ := strconv.Atoi(os.Getenv("DRY_RUN"))
	if dryrun == 0 {
		fmt.Println("subject : " + subject)
		fmt.Println("templateFile : " + templateFile)
		utils.PrintVars(os.Stdout, true, sampleData)
		return nil
	}

	t, err := template.ParseFiles(templateFile)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	err = t.Execute(buf, sampleData)
	if err != nil {
		return err
	}
	body := buf.String()

	m := gomail.NewMessage()
	m.SetHeader("From", smtp.hostMail)
	m.SetHeader("To", sampleData.SendTo...)
	if len(sampleData.SendCc) > 0 {
		m.SetHeader("Cc", sampleData.SendCc...)
	}
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtp.smtpAdrr, smtp.smtpPort, smtp.hostMail, smtp.hostMailPassword)

	// Send the email
	err = d.DialAndSend(m)
	return err
}
