package utils

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

const (
	docxBodyPath       = "word/document.xml"
	docxTextElement    = "t"
	docxParagraph      = "p"
	docxTab            = "tab"
	docxLineBreak      = "br"
	documentReadFactor = 4
	docxMaxBodyBytes   = 64 << 20
)

func ExtractPDFText(content []byte, maxChars int) (text string, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("extract pdf text: %w: %v", config.ErrUnreadableDocument, recovered)
		}
	}()

	reader, err := pdf.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return "", fmt.Errorf("extract pdf text: %w: %w", config.ErrUnreadableDocument, err)
	}

	plain, err := reader.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("extract pdf text: %w: %w", config.ErrUnreadableDocument, err)
	}

	return readBoundedText(plain, maxChars)
}

func ExtractDocxText(content []byte, maxChars int) (string, error) {
	archive, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return "", fmt.Errorf("extract docx text: %w: %w", config.ErrUnreadableDocument, err)
	}

	for _, file := range archive.File {
		if file.Name != docxBodyPath {
			continue
		}

		body, err := file.Open()
		if err != nil {
			return "", fmt.Errorf("extract docx text: %w: %w", config.ErrUnreadableDocument, err)
		}

		defer func() { _ = body.Close() }()

		return decodeDocxBody(body, maxChars)
	}

	return "", fmt.Errorf("extract docx text: %w", config.ErrUnreadableDocument)
}

func decodeDocxBody(body io.Reader, maxChars int) (string, error) {
	decoder := xml.NewDecoder(&boundedReader{reader: body, remaining: docxMaxBodyBytes})
	builder := strings.Builder{}
	insideText := false

	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}

		if errors.Is(err, config.ErrDocumentTooLarge) {
			return "", fmt.Errorf("extract docx text: %w", err)
		}

		if err != nil {
			return "", fmt.Errorf("extract docx text: %w: %w", config.ErrUnreadableDocument, err)
		}

		switch element := token.(type) {
		case xml.StartElement:
			insideText = element.Name.Local == docxTextElement
			writeDocxSeparator(&builder, element.Name.Local)
		case xml.EndElement:
			insideText = false

			if element.Name.Local == docxParagraph {
				builder.WriteByte('\n')
			}
		case xml.CharData:
			if insideText {
				builder.Write(element)
			}
		}

		if builder.Len() > maxChars*documentReadFactor {
			return "", fmt.Errorf("extract docx text: %w", config.ErrDocumentTooLarge)
		}
	}

	return truncateText(builder.String(), maxChars), nil
}

func writeDocxSeparator(builder *strings.Builder, element string) {
	switch element {
	case docxTab:
		builder.WriteByte('\t')
	case docxLineBreak:
		builder.WriteByte('\n')
	}
}

func readBoundedText(reader io.Reader, maxChars int) (string, error) {
	raw, err := io.ReadAll(io.LimitReader(reader, int64(maxChars)*documentReadFactor))
	if err != nil {
		return "", fmt.Errorf("read document text: %w: %w", config.ErrUnreadableDocument, err)
	}

	return truncateText(string(raw), maxChars), nil
}

func truncateText(text string, maxChars int) string {
	characters := []rune(strings.ToValidUTF8(text, " "))
	if len(characters) > maxChars {
		characters = characters[:maxChars]
	}

	return strings.TrimSpace(string(characters))
}

type boundedReader struct {
	reader    io.Reader
	remaining int64
}

func (r *boundedReader) Read(buffer []byte) (int, error) {
	if r.remaining <= 0 {
		return 0, config.ErrDocumentTooLarge
	}

	if int64(len(buffer)) > r.remaining {
		buffer = buffer[:r.remaining]
	}

	read, err := r.reader.Read(buffer)
	r.remaining -= int64(read)

	return read, err
}
