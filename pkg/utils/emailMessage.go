package utils

import (
	"bufio"
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"regexp"
	"slices"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

func writeQuotedPrintable(message *bytes.Buffer, text string) error {
	writer := quotedprintable.NewWriter(message)
	if _, err := writer.Write([]byte(strings.ReplaceAll(text, "\n", "\r\n"))); err != nil {
		return err
	}

	return writer.Close()
}

func NewEmailMessageID(fromAddress string) string {
	_, domain, found := strings.Cut(fromAddress, "@")
	if !found || domain == "" {
		domain = "localhost"
	}

	return "<" + GenerateUUID() + "@" + domain + ">"
}

// Plain text by default: it keeps cold email out of spam folders.
func BuildEmailMessage(email sharedTypes.OutgoingEmail) ([]byte, error) {
	var message bytes.Buffer

	headers := [][2]string{
		{"From", (&mail.Address{Name: email.FromName, Address: email.FromAddress}).String()},
		{"To", (&mail.Address{Name: email.ToName, Address: email.ToAddress}).String()},
		{"Subject", mime.QEncoding.Encode("utf-8", email.Subject)},
		{"Date", email.Date.Format("Mon, 02 Jan 2006 15:04:05 -0700")},
		{"Message-ID", email.MessageID},
	}

	if email.InReplyTo != "" {
		headers = append(headers, [2]string{"In-Reply-To", email.InReplyTo},
			[2]string{"References", email.InReplyTo})
	}

	if email.ListUnsubscribe != "" {
		headers = append(
			headers,
			[2]string{config.GetListUnsubscribeHeader(), email.ListUnsubscribe},
		)
	}

	if email.OneClickUnsubscribe {
		headers = append(headers, [2]string{
			config.GetListUnsubscribePostHeader(),
			config.GetListUnsubscribeOneClickValue(),
		})
	}

	headers = append(headers, [2]string{"MIME-Version", "1.0"})

	if email.HTMLBody == "" {
		headers = append(headers,
			[2]string{"Content-Type", "text/plain; charset=utf-8"},
			[2]string{"Content-Transfer-Encoding", "quoted-printable"},
		)
	} else {
		headers = append(headers, [2]string{
			"Content-Type",
			`multipart/alternative; boundary="` + config.GetEmailAlternativeBoundary() + `"`,
		})
	}

	for _, header := range headers {
		message.WriteString(header[0])
		message.WriteString(": ")
		message.WriteString(header[1])
		message.WriteString("\r\n")
	}

	message.WriteString("\r\n")

	if email.HTMLBody == "" {
		if err := writeQuotedPrintable(&message, email.Body); err != nil {
			return nil, err
		}

		return message.Bytes(), nil
	}

	for _, part := range [][2]string{{"text/plain", email.Body}, {"text/html", email.HTMLBody}} {
		message.WriteString("--")
		message.WriteString(config.GetEmailAlternativeBoundary())
		message.WriteString("\r\n")
		message.WriteString("Content-Type: ")
		message.WriteString(part[0])
		message.WriteString("; charset=utf-8\r\n")
		message.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")

		if err := writeQuotedPrintable(&message, part[1]); err != nil {
			return nil, err
		}

		message.WriteString("\r\n")
	}

	message.WriteString("--")
	message.WriteString(config.GetEmailAlternativeBoundary())
	message.WriteString("--\r\n")

	return message.Bytes(), nil
}

var (
	messageIDPatternI       = regexp.MustCompile(`<[^<>\s]+@[^<>\s]+>`)
	originalRecipientRegexI = regexp.MustCompile(
		`(?im)^(?:final|original)-recipient:\s*rfc822;\s*(\S+)`,
	)
)

func ParseInboundEmail(uid uint32, raw []byte) (sharedTypes.InboundEmail, error) {
	message, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return sharedTypes.InboundEmail{}, err
	}

	inbound := sharedTypes.InboundEmail{
		UID:           uid,
		MessageID:     strings.TrimSpace(message.Header.Get("Message-ID")),
		ReferencedIDs: referencedIDs(message.Header),
		IsAutoReply:   isAutoReply(message.Header),
		IsUnsubscribeRequest: isUnsubscribeSubject(
			decodeHeader(message.Header.Get("Subject")),
		),
	}

	if from, err := mail.ParseAddress(message.Header.Get("From")); err == nil {
		inbound.FromAddress = strings.ToLower(from.Address)
	}

	mediaType, params, _ := mime.ParseMediaType(message.Header.Get("Content-Type"))
	isReport := mediaType == "multipart/report" &&
		strings.EqualFold(params["report-type"], "delivery-status")

	if !isReport && !isBounceSender(inbound.FromAddress) {
		return inbound, nil
	}

	inbound.IsBounce = true
	body, _ := io.ReadAll(message.Body)

	if isReport && params["boundary"] != "" {
		body = flattenParts(body, params["boundary"])
	}

	inbound.BouncedMessageIDs = messageIDPatternI.FindAllString(string(body), -1)
	for _, match := range originalRecipientRegexI.FindAllStringSubmatch(string(body), -1) {
		inbound.BouncedAddresses = append(inbound.BouncedAddresses, strings.ToLower(match[1]))
	}

	return inbound, nil
}

func referencedIDs(header mail.Header) []string {
	ids := messageIDPatternI.FindAllString(
		header.Get("In-Reply-To")+" "+header.Get("References"),
		-1,
	)
	slices.Sort(ids)

	return slices.Compact(ids)
}

func isAutoReply(header mail.Header) bool {
	autoSubmitted := strings.ToLower(strings.TrimSpace(header.Get("Auto-Submitted")))
	precedence := strings.ToLower(header.Get("Precedence"))

	return (autoSubmitted != "" && autoSubmitted != "no") ||
		header.Get("X-Autoreply") != "" ||
		header.Get("X-Autorespond") != "" ||
		precedence == "auto_reply" || precedence == "bulk" || precedence == "junk"
}

func isBounceSender(address string) bool {
	local, _, _ := strings.Cut(address, "@")

	return slices.Contains(config.GetBounceSenderPrefixes(), local)
}

func flattenParts(body []byte, boundary string) []byte {
	var flattened bytes.Buffer

	reader := multipart.NewReader(bytes.NewReader(body), boundary)
	for {
		part, err := reader.NextPart()
		if err != nil {
			break
		}

		content, _ := io.ReadAll(bufio.NewReader(part))
		flattened.Write(content)
		flattened.WriteString("\n")
	}

	return flattened.Bytes()
}

func isUnsubscribeSubject(subject string) bool {
	subject = strings.ToLower(strings.TrimSpace(subject))

	for trimmed := true; trimmed; {
		trimmed = false

		for _, prefix := range config.GetReplySubjectPrefixes() {
			if strings.HasPrefix(subject, prefix) {
				subject = strings.TrimSpace(strings.TrimPrefix(subject, prefix))
				trimmed = true
			}
		}
	}

	return subject == config.GetUnsubscribeKeyword()
}

func decodeHeader(value string) string {
	decoded, err := new(mime.WordDecoder).DecodeHeader(value)
	if err != nil {
		return value
	}

	return decoded
}
