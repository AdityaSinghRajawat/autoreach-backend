package helpers

import (
	"archive/zip"
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func BuildPDF(t *testing.T, lines ...string) []byte {
	t.Helper()

	stream := strings.Builder{}
	stream.WriteString("BT\n/F1 12 Tf\n72 720 Td\n14 TL\n")

	for _, line := range lines {
		escaped := strings.NewReplacer(`\`, `\\`, `(`, `\(`, `)`, `\)`).Replace(line)
		stream.WriteString("(" + escaped + ") Tj\nT*\n")
	}

	stream.WriteString("ET\n")

	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R " +
			"/Resources << /Font << /F1 5 0 R >> >> >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", stream.Len(), stream.String()),
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>",
	}

	document := bytes.Buffer{}
	document.WriteString("%PDF-1.4\n")

	offsets := make([]int, 0, len(objects))
	for index, object := range objects {
		offsets = append(offsets, document.Len())
		fmt.Fprintf(&document, "%d 0 obj\n%s\nendobj\n", index+1, object)
	}

	crossReferenceOffset := document.Len()
	fmt.Fprintf(&document, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)

	for _, offset := range offsets {
		fmt.Fprintf(&document, "%010d 00000 n \n", offset)
	}

	fmt.Fprintf(
		&document,
		"trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n",
		len(objects)+1,
		crossReferenceOffset,
	)

	return document.Bytes()
}

func BuildDocx(t *testing.T, paragraphs ...string) []byte {
	t.Helper()

	body := strings.Builder{}
	body.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	body.WriteString(
		`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`,
	)

	for _, paragraph := range paragraphs {
		escaped := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(paragraph)
		body.WriteString(`<w:p><w:r><w:t xml:space="preserve">` + escaped + `</w:t></w:r></w:p>`)
	}

	body.WriteString(`</w:body></w:document>`)

	archive := bytes.Buffer{}
	writer := zip.NewWriter(&archive)

	contentTypes, err := writer.Create("[Content_Types].xml")
	require.NoError(t, err)

	_, err = contentTypes.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><Types/>`))
	require.NoError(t, err)

	document, err := writer.Create("word/document.xml")
	require.NoError(t, err)

	_, err = document.Write([]byte(body.String()))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	return archive.Bytes()
}

func DoMultipart(
	t *testing.T,
	handler http.Handler,
	path, field, fileName string,
	content []byte,
	headers map[string]string,
) Exchange {
	t.Helper()

	body := bytes.Buffer{}
	writer := multipart.NewWriter(&body)

	if field != "" {
		part, err := writer.CreateFormFile(field, fileName)
		require.NoError(t, err)

		_, err = part.Write(content)
		require.NoError(t, err)
	}

	require.NoError(t, writer.Close())

	request := httptest.NewRequest(http.MethodPost, path, &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	for name, value := range headers {
		request.Header.Set(name, value)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	return Exchange{Request: request, Recorder: recorder, Body: recorder.Body.Bytes()}
}
