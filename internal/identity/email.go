package identity

import (
	"context"
	"fmt"
	"net/smtp"
	"net/url"
	"strings"
)

// VerificationMailer is deliberately narrow: it receives no password, API
// key, session, or event data—only a recipient and one-time confirmation URL.
type VerificationMailer interface {
	SendVerification(context.Context, string, string) error
}

type SMTPMailer struct{ addr, from, username, password string }

func NewSMTPMailer(addr, from, username, password string) SMTPMailer {
	return SMTPMailer{addr: addr, from: from, username: username, password: password}
}

func (m SMTPMailer) SendVerification(_ context.Context, recipient, verificationURL string) error {
	if m.addr == "" || m.from == "" {
		return fmt.Errorf("SMTP mailer is not configured")
	}
	header := "To: " + recipient + "\r\nSubject: Confirm your TravelingHub email\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n"
	message := []byte(header + "Open this one-time link within 24 hours to confirm your email:\r\n" + verificationURL + "\r\n")
	var auth smtp.Auth
	if m.username != "" {
		host, _, _ := strings.Cut(m.addr, ":")
		auth = smtp.PlainAuth("", m.username, m.password, host)
	}
	return smtp.SendMail(m.addr, auth, m.from, []string{recipient}, message)
}

func verificationURL(origin, token string) (string, error) {
	base, err := url.Parse(origin)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return "", fmt.Errorf("invalid verification origin")
	}
	base.Path = "/v1/web/verify-email"
	query := base.Query()
	query.Set("token", token)
	base.RawQuery = query.Encode()
	return base.String(), nil
}
