package utils

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/ledongthuc/pdf"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

// Streams are measured before parsing, so a decompression bomb is refused.
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

	if reader.NumPage() > config.GetPDFMaxPages() {
		return "", fmt.Errorf("extract pdf text: %w", config.ErrDocumentTooLarge)
	}

	budget := config.GetPDFMaxDecodedBytes()
	builder := strings.Builder{}

	for number := 1; number <= reader.NumPage(); number++ {
		page := reader.Page(number)
		if page.V.IsNull() {
			continue
		}

		streams := []pdf.Value{page.V.Key("Contents")}

		for _, name := range page.Fonts() {
			streams = append(streams, page.Font(name).V.Key("ToUnicode"))
		}

		if budget, err = spendDecodedBytes(streams, budget); err != nil {
			return "", fmt.Errorf("extract pdf text: %w", err)
		}

		builder.WriteString(positionedText(page.Content().Text))

		if builder.Len() > maxChars*config.GetDocumentReadFactor() {
			break
		}
	}

	return truncateText(builder.String(), maxChars), nil
}

// Words are rebuilt from glyph positions: many PDFs (LaTeX among them) place words apart without
// storing a space, and some add invisible zero-width marker glyphs.
func positionedText(glyphs []pdf.Text) string {
	measuredFonts := map[string]bool{}

	for _, glyph := range glyphs {
		if glyph.W > 0 {
			measuredFonts[glyph.Font] = true
		}
	}

	builder := strings.Builder{}

	var previous *pdf.Text

	for index := range glyphs {
		glyph := &glyphs[index]
		if glyph.W <= 0 && measuredFonts[glyph.Font] {
			continue
		}

		if previous != nil {
			size := math.Max(glyph.FontSize, 1)
			gap := glyph.X - (previous.X + previous.W)

			switch {
			case math.Abs(glyph.Y-previous.Y) > size*config.GetPDFLineShiftRatio():
				builder.WriteString("\n")
			case previous.W > 0 && gap > size*config.GetPDFWordGapRatio() &&
				!strings.HasSuffix(previous.S, " ") && !strings.HasPrefix(glyph.S, " "):
				builder.WriteString(" ")
			}
		}

		builder.WriteString(glyph.S)
		previous = glyph
	}

	builder.WriteString("\n")

	return builder.String()
}

func spendDecodedBytes(values []pdf.Value, budget int64) (int64, error) {
	for _, value := range values {
		switch value.Kind() {
		case pdf.Array:
			parts := make([]pdf.Value, 0, value.Len())
			for index := range value.Len() {
				parts = append(parts, value.Index(index))
			}

			remaining, err := spendDecodedBytes(parts, budget)
			if err != nil {
				return 0, err
			}

			budget = remaining
		case pdf.Stream:
			stream := value.Reader()
			decoded, err := io.CopyN(io.Discard, stream, budget+1)
			_ = stream.Close()

			if decoded > budget {
				return 0, config.ErrDocumentTooLarge
			}

			if err != nil && !errors.Is(err, io.EOF) {
				return 0, fmt.Errorf("%w: %w", config.ErrUnreadableDocument, err)
			}

			budget -= decoded
		}
	}

	return budget, nil
}

func ExtractDocxText(content []byte, maxChars int) (string, error) {
	archive, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return "", fmt.Errorf("extract docx text: %w: %w", config.ErrUnreadableDocument, err)
	}

	for _, file := range archive.File {
		if file.Name != config.GetDocxBodyPath() {
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
	decoder := xml.NewDecoder(&boundedReader{reader: body, remaining: config.GetDocxMaxBodyBytes()})
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
			insideText = element.Name.Local == config.GetDocxTextElement()
			writeDocxSeparator(&builder, element.Name.Local)
		case xml.EndElement:
			insideText = false

			if element.Name.Local == config.GetDocxParagraph() {
				builder.WriteByte('\n')
			}
		case xml.CharData:
			if insideText {
				builder.Write(element)
			}
		}

		if builder.Len() > maxChars*config.GetDocumentReadFactor() {
			return "", fmt.Errorf("extract docx text: %w", config.ErrDocumentTooLarge)
		}
	}

	return truncateText(builder.String(), maxChars), nil
}

func writeDocxSeparator(builder *strings.Builder, element string) {
	switch element {
	case config.GetDocxTab():
		builder.WriteByte('\t')
	case config.GetDocxLineBreak():
		builder.WriteByte('\n')
	}
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
