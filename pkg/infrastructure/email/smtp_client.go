package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/smtp"
	"net/textproto"

	"github.com/unclaim/chegonado.git/internal/shared/config"
	"github.com/unclaim/chegonado.git/internal/shared/ports"
)

// SMTPClient реализует интерфейс ports.EmailSender для отправки писем через SMTP.
type SMTPClient struct {
	cfg *config.SMTPConfig
}

// NewSMTPClient создает новый экземпляр SMTPClient.
func NewSMTPClient(cfg *config.SMTPConfig) *SMTPClient {
	return &SMTPClient{cfg: cfg}
}

// SendEmail отправляет простое письмо без вложений.
func (s *SMTPClient) SendEmail(to, subject string, body io.Reader) error {
	return s.SendEmailWithAttachments(to, subject, body, nil)
}

// SendEmailWithAttachments отправляет письмо с вложениями.
// SendEmailWithAttachments отправляет письмо с вложениями.
func (s *SMTPClient) SendEmailWithAttachments(to, subject string, body io.Reader, attachments []ports.Attachment) error {
	smtpServer := fmt.Sprintf("%s:%s", s.cfg.Host, s.cfg.Port)
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)

	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	boundary := writer.Boundary()

	header := make(textproto.MIMEHeader)
	header.Set("From", fmt.Sprintf("%s <%s>", s.cfg.FromName, s.cfg.From))
	header.Set("To", to)
	header.Set("Subject", subject)
	header.Set("MIME-Version", "1.0")
	header.Set("Content-Type", fmt.Sprintf("multipart/mixed; boundary=%s", boundary))

	// Используем fmt.Fprintf вместо buf.WriteString(fmt.Sprintf())
	for k, v := range header {
		fmt.Fprintf(buf, "%s: %s\r\n", k, v[0])
	}
	buf.WriteString("\r\n")

	htmlPartHeader := make(textproto.MIMEHeader)
	htmlPartHeader.Set("Content-Type", "text/html; charset=UTF-8")
	htmlPartWriter, err := writer.CreatePart(htmlPartHeader)
	if err != nil {
		return fmt.Errorf("ошибка создания части письма: %w", err)
	}
	if _, err := io.Copy(htmlPartWriter, body); err != nil {
		return fmt.Errorf("ошибка копирования тела письма: %w", err)
	}

	for _, attachment := range attachments {
		attachmentPartHeader := make(textproto.MIMEHeader)
		attachmentPartHeader.Set("Content-Type", fmt.Sprintf("%s; name=\"%s\"", attachment.ContentType, attachment.Filename))
		attachmentPartHeader.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", attachment.Filename))
		attachmentPartHeader.Set("Content-Transfer-Encoding", "base64")

		attachmentPartWriter, err := writer.CreatePart(attachmentPartHeader)
		if err != nil {
			return fmt.Errorf("ошибка создания части вложения: %w", err)
		}

		encoder := base64.NewEncoder(base64.StdEncoding, attachmentPartWriter)
		if _, err := encoder.Write(attachment.Data); err != nil {
			return fmt.Errorf("ошибка кодирования вложения: %w", err)
		}
		if err := encoder.Close(); err != nil {
			return fmt.Errorf("error closing base64 encoder: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("error closing multipart writer: %w", err)
	}

	err = smtp.SendMail(smtpServer, auth, s.cfg.From, []string{to}, buf.Bytes())
	if err != nil {
		return fmt.Errorf("ошибка отправки письма: %w", err)
	}

	return nil
}
