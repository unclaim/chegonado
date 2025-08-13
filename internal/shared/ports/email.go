package ports

import "io"

// Attachment представляет собой файл-вложение для письма.
type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// EmailSender определяет интерфейс для отправки электронных писем.
type EmailSender interface {
	// SendEmail отправляет простое письмо с указанными параметрами.
	SendEmail(to, subject string, body io.Reader) error

	// SendEmailWithAttachments отправляет письмо с вложениями.
	SendEmailWithAttachments(to, subject string, body io.Reader, attachments []Attachment) error
}
