package configs

const (
	MaxTemplateFileSize = 5 // In MB
)

var (
	// TemplateFileType is map MIME type to extension
	// Ref: https://mimesniff.spec.whatwg.org/
	TemplateFileType = map[string]string{
		"application/msword": "doc",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": "docx",
		"application/octet-stream": "doc",
		"application/zip":          "docx",
		"application/pdf":          "pdf",
	}
)
