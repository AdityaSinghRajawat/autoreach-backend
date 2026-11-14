package helpers

import (
	"bytes"
	"path"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

type ResumeStatus string

const (
	ResumeStatusUploaded ResumeStatus = "UPLOADED"
	ResumeStatusParsed   ResumeStatus = "PARSED"
	ResumeStatusEmbedded ResumeStatus = "EMBEDDED"
	ResumeStatusFailed   ResumeStatus = "FAILED"

	ContentTypePDF  = "application/pdf"
	ContentTypeDOCX = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
)

func DetectResumeContentType(content []byte) string {
	switch {
	case bytes.HasPrefix(content, config.GetPDFMagic()):
		return ContentTypePDF
	case bytes.HasPrefix(content, config.GetZipMagic()) && bytes.Contains(content, []byte(config.GetDocxBodyPath())):
		return ContentTypeDOCX
	default:
		return ""
	}
}

func ResumeExtension(contentType string) string {
	switch contentType {
	case ContentTypePDF:
		return config.GetPDFExtension()
	case ContentTypeDOCX:
		return config.GetDOCXExtension()
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

	if len(base) > config.GetMaxFileNameLength() {
		base = base[len(base)-config.GetMaxFileNameLength():]
	}

	if base == "" || base == "." || base == "/" {
		return config.GetFallbackFileName()
	}

	return base
}
