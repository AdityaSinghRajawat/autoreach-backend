package shared

type MultipartFile struct {
	FieldName   string
	FileName    string
	ContentType string
	Content     []byte
}
