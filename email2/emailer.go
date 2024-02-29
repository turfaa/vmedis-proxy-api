package email2

import (
	"crypto/tls"
	"fmt"
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
