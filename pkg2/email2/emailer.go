package email2

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"time"

	"github.com/jordan-wright/email"
)

type Emailer struct {
	smtpAddress string
	auth        smtp.Auth
	tlsConfig   *tls.Config
}

func (e *Emailer) Send(mail *email.Email, timeout time.Duration) error {
	if err := mail.SendWithTLS(e.smtpAddress, e.auth, e.tlsConfig); err != nil {
		js, _ := json.Marshal(mail)
		log.Printf("Failed to send email: %s", js)
		return fmt.Errorf("send email with TLS: %w", err)
	}

	return nil
}

func NewEmailer(smtpAddress string, auth smtp.Auth, tlsConfig *tls.Config) *Emailer {
	return &Emailer{
		smtpAddress: smtpAddress,
		auth:        auth,
		tlsConfig:   tlsConfig,
	}
}
