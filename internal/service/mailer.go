package service

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/sangtandoan/social/internal/config"
	"github.com/sangtandoan/social/internal/utils"
	"gopkg.in/gomail.v2"
)

type Mailer interface {
	Send(req *SendRequest) error
	SendWithRetry(req *SendRequest, retries int) error
}

type TemplateOpt int

const (
	ConfirmTemplate TemplateOpt = iota
	DeleteTemplate
)

type SendRequest struct {
	Data any
	To   []string
	Temp TemplateOpt
}

type ConfirmData struct {
	Username string
	Token    string
}

type EmailTemplate struct {
	Subject string
	Body    string
	Path    string
}

type SMTPMailer struct {
	config *config.MailerConfig
	dialer *gomail.Dialer
}

func NewSMTPMailer(config *config.MailerConfig) *SMTPMailer {
	dialer := gomail.NewDialer(config.Host, config.Port, config.Username, config.Password)

	return &SMTPMailer{
		config,
		dialer,
	}
}

// Retry mechanisim with expoential backoff
func (m *SMTPMailer) SendWithRetry(req *SendRequest, retries int) error {
	for i := range retries {
		if err := m.Send(req); err != nil {
			utils.Log.Warnf(" %d attempt to send email of %d", i+1, retries)

			time.Sleep(time.Second * time.Duration((i + 1))) // exponential backoff
			continue
		}
		utils.Log.Info("send email successfully")
		return nil
	}

	return utils.NewApiError(http.StatusInternalServerError, "can not send email")
}

func (m *SMTPMailer) Send(req *SendRequest) error {
	tmp := getTemplateEmail(req.Temp)

	data := m.getData(req.Temp, req.Data)
	if data == nil {
		return utils.NewApiError(
			http.StatusInternalServerError,
			"can not extract data for template",
		)
	}

	if err := parseTemplate(tmp, data); err != nil {
		return err
	}

	message := gomail.NewMessage()
	message.SetHeader("From", m.config.From)
	message.SetHeader("To", req.To...)
	message.SetHeader("Subject", tmp.Subject)
	message.SetBody("text/html", tmp.Body)

	if err := m.dialer.DialAndSend(message); err != nil {
		utils.Log.Error(err.Error())
		return err
	}

	return nil
}

func getTemplateEmail(opt TemplateOpt) *EmailTemplate {
	var template EmailTemplate
	switch opt {
	case ConfirmTemplate:
		template.Path = "confirm-email.tmpl"
	}

	return &template
}

func (m *SMTPMailer) getData(opt TemplateOpt, data any) any {
	switch opt {
	case ConfirmTemplate:
		if newData, ok := data.(*ConfirmData); ok {
			return struct {
				Username      string
				ActivationURL string
			}{
				Username:      newData.Username,
				ActivationURL: fmt.Sprintf("%s/activate/%s", m.config.ServerAddr, newData.Token),
			}
		}
	}

	return nil
}

//go:embed "templates/*"
var FS embed.FS

func parseTemplate(tmp *EmailTemplate, data any) error {
	t, err := template.ParseFS(FS, "templates/"+tmp.Path)
	if err != nil {
		return err
	}

	var subject bytes.Buffer
	if err := t.ExecuteTemplate(&subject, "subject", data); err != nil {
		return err
	}

	var body bytes.Buffer
	if err := t.ExecuteTemplate(&body, "body", data); err != nil {
		return err
	}

	tmp.Body = body.String()
	tmp.Subject = subject.String()
	return nil
}
