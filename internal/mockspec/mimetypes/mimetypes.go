package mimetypes

// Common MIME Type Constants
const (
	ContentTypeJSON           = "application/json"
	ContentTypeHTML           = "text/html"
	ContentTypePlain          = "text/plain"
	ContentTypeXML            = "application/xml"
	ContentTypeFormURLEncoded = "application/x-www-form-urlencoded"
	ContentTypeMultipart      = "multipart/form-data"
	ContentTypeJavaScript     = "application/javascript"
	ContentTypeOctetStream    = "application/octet-stream"
	ContentTypePNG            = "image/png"
	ContentTypeJPEG           = "image/jpeg"
	ContentTypeGIF            = "image/gif"
	ContentTypeCSV            = "text/csv"
	ContentTypePDF            = "application/pdf"
	ContentTypeZIP            = "application/zip"
	ContentTypeSVG            = "image/svg+xml"
	ContentTypeWebP           = "image/webp"
)

// ExtensionToContentType is a map of file extensions to content types
var ExtensionToContentType = map[string]string{
	".json": ContentTypeJSON,
	".html": ContentTypeHTML,
	".txt":  ContentTypePlain,
	".xml":  ContentTypeXML,
	".js":   ContentTypeJavaScript,
	".png":  ContentTypePNG,
	".jpg":  ContentTypeJPEG,
	".jpeg": ContentTypeJPEG,
	".gif":  ContentTypeGIF,
	".csv":  ContentTypeCSV,
	".pdf":  ContentTypePDF,
	".zip":  ContentTypeZIP,
	".svg":  ContentTypeSVG,
	".webp": ContentTypeWebP,
}
