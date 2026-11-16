package mocks

import (
	"context"
	"sync"

	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

type SentMail struct {
	To      string
	Subject string
	Body    string
}

type SystemMailer struct {
	mutex sync.Mutex
	Sent  []SentMail
}

var _ sharedTypes.SystemMailer = (*SystemMailer)(nil)

func (m *SystemMailer) Send(
	_ context.Context,
	toAddress string,
	email sharedTypes.AccountEmail,
) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Sent = append(m.Sent, SentMail{To: toAddress, Subject: email.Subject, Body: email.Text})

	return nil
}

func (m *SystemMailer) SentTo(address string) []SentMail {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sent := []SentMail{}

	for _, mail := range m.Sent {
		if mail.To == address {
			sent = append(sent, mail)
		}
	}

	return sent
}
