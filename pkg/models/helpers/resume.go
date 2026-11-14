package helpers

import (
	"bytes"
	"path"
	"strings"
)

type ResumeStatus string

const (
	ResumeStatusUploaded ResumeStatus = "UPLOADED"
	ResumeStatusParsed   ResumeStatus = "PARSED"
	ResumeStatusEmbedded ResumeStatus = "EMBEDDED"
	ResumeStatusFailed   ResumeStatus = "FAILED"

	ContentTypePDF  = "application/pdf"
	ContentTypeDOCX = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"

	extensionPDF  = ".pdf"
	extensionDOCX = ".docx"
)

var (
	pdfMagic        = []byte("%PDF-")
	zipMagic        = []byte("PK\x03\x04")
	docxBodyMarker  = []byte("word/document.xml")
	maxResumeNameLn = 255
)

func DetectResumeContentType(content []byte) string {
	switch {
	case bytes.HasPrefix(content, pdfMagic):
		return ContentTypePDF
	case bytes.HasPrefix(content, zipMagic) && bytes.Contains(content, docxBodyMarker):
		return ContentTypeDOCX
	default:
		return ""
	}
}

func ResumeExtension(contentType string) string {
	switch contentType {
	case ContentTypePDF:
		return extensionPDF
	case ContentTypeDOCX:
		return extensionDOCX
	default:
		return ""
	}
}

func SanitizeFileName(name string) string {
	base := path.Base(strings.ReplaceAll(name, "\\", "/"))
	base = strings.Map(func(character rune) rune {
		if character < 32 || character == 127 {
			return -1
		}

		return character
	}, base)
	base = strings.TrimSpace(base)

	if len(base) > maxResumeNameLn {
		base = base[len(base)-maxResumeNameLn:]
	}

	if base == "" || base == "." || base == "/" {
		return "resume"
	}

	return base
}
