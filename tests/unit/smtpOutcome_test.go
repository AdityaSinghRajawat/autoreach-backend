package unit

import (
	"bufio"
	"context"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

// fakeSMTPServer speaks just enough SMTP to accept one message. acceptData decides whether it
// answers the end of DATA; either way it drops the connection instead of answering QUIT.
func fakeSMTPServer(t *testing.T, acceptData bool) int {
	t.Helper()

	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	serve := func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()

		reader := bufio.NewReader(conn)
		write := func(line string) { _, _ = conn.Write([]byte(line + "\r\n")) }

		write("220 fake ESMTP")

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}

			switch command := strings.ToUpper(strings.TrimSpace(line)); {
			case strings.HasPrefix(command, "EHLO"):
				write("250-fake")
				write("250 AUTH PLAIN")
			case strings.HasPrefix(command, "AUTH"):
				write("235 authenticated")
			case strings.HasPrefix(command, "MAIL"), strings.HasPrefix(command, "RCPT"):
				write("250 ok")
			case command == "DATA":
				write("354 go ahead")

				for {
					body, err := reader.ReadString('\n')
					if err != nil || body == ".\r\n" {
						break
					}
				}

				if !acceptData {
					return
				}

				write("250 queued")
			case command == "QUIT":
				return
			}
		}
	}

	go serve()

	return listener.Addr().(*net.TCPAddr).Port
}

func sendThrough(t *testing.T, port int) error {
	t.Helper()

	server := sharedTypes.EmailServer{
		Host:     "127.0.0.1",
		Port:     port,
		Security: string(modelHelpers.MailSecurityNone),
	}

	return utils.NewEmailUtil().SendEmail(context.Background(), sharedTypes.EmailAccount{
		Address:  "alex@example.com",
		Username: "alex",
		Password: "secret",
		SMTP:     server,
		IMAP:     server,
	}, sharedTypes.OutgoingEmail{
		FromAddress: "alex@example.com",
		ToAddress:   "riley@acme.com",
		Subject:     "Hello",
		Body:        "Hi Riley",
		MessageID:   utils.NewEmailMessageID("alex@example.com"),
		Date:        time.Now(),
	})
}

func TestAcceptedEmailIsSentEvenWhenQuitFails(t *testing.T) {
	setup.Config(t)

	err := sendThrough(t, fakeSMTPServer(t, true))

	assert.NoError(
		t,
		err,
		"the server queued the message, so a dropped QUIT must not trigger a resend",
	)
}

func TestLosingTheConnectionAtTheEndOfDataLeavesTheOutcomeUnknown(t *testing.T) {
	setup.Config(t)

	err := sendThrough(t, fakeSMTPServer(t, false))

	require.Error(t, err)
	assert.ErrorIs(
		t,
		err,
		config.ErrSendOutcomeUnknown,
		"the message may have been queued, so it is not retried",
	)
	assert.NotContains(t, strconv.Quote(err.Error()), "riley@acme.com")
}
