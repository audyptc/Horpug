package usecase

import (
	"bytes"
	"net/http"
	"path"
	"strings"
	"unicode"
	"unicode/utf8"

	documentdomain "apihorpug/internal/features/document/domain"
)

// MaxUploadSize caps a single uploaded file. nginx's client_max_body_size and
// the Fiber body limit are set slightly above this to leave room for the
// multipart envelope.
const MaxUploadSize = 8 << 20

const maxOriginalNameLength = 255

// FileUpload is a file received from a client, not yet validated or stored.
type FileUpload struct {
	Name string
	Data []byte
}

// StoredFile describes a file already written to storage.
type StoredFile struct {
	Path string
	Name string
	Mime string
	Size int64
}

type fileType struct {
	mime string
	// office marks formats the sniffer can't identify precisely (a .docx is
	// just a zip, an old .doc is an opaque OLE blob), so the extension is
	// trusted alongside a generic sniffed type.
	office bool
}

// allowedFileTypes is the whitelist of extensions accepted for upload. The
// extension is only ever taken from here, never from the client's file name,
// and SVG/HTML are deliberately absent — served back from our own origin they
// could run script.
var allowedFileTypes = map[string]fileType{
	".pdf":  {mime: "application/pdf"},
	".jpg":  {mime: "image/jpeg"},
	".jpeg": {mime: "image/jpeg"},
	".png":  {mime: "image/png"},
	".gif":  {mime: "image/gif"},
	".webp": {mime: "image/webp"},
	".doc":  {mime: "application/msword", office: true},
	".docx": {mime: "application/vnd.openxmlformats-officedocument.wordprocessingml.document", office: true},
	".xls":  {mime: "application/vnd.ms-excel", office: true},
	".xlsx": {mime: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", office: true},
	".ppt":  {mime: "application/vnd.ms-powerpoint", office: true},
	".pptx": {mime: "application/vnd.openxmlformats-officedocument.presentationml.presentation", office: true},
}

// validateUpload checks size and that the content matches the claimed
// extension, and returns the extension and canonical MIME type to store.
func validateUpload(upload *FileUpload) (ext, mime string, err error) {
	if len(upload.Data) == 0 {
		return "", "", documentdomain.ErrEmptyFile
	}
	if len(upload.Data) > MaxUploadSize {
		return "", "", documentdomain.ErrFileTooLarge
	}

	ext = strings.ToLower(path.Ext(cleanFileName(upload.Name)))
	fileType, ok := allowedFileTypes[ext]
	if !ok {
		return "", "", documentdomain.ErrUnsupportedFileType
	}

	sniffed := http.DetectContentType(upload.Data)
	if fileType.office {
		if sniffed != "application/zip" && sniffed != "application/octet-stream" {
			return "", "", documentdomain.ErrUnsupportedFileType
		}
	} else if sniffed != fileType.mime {
		return "", "", documentdomain.ErrUnsupportedFileType
	}

	return ext, fileType.mime, nil
}

// cleanFileName reduces a client-supplied name to a plain, printable base
// name for display and download. It is never used to build a storage path.
func cleanFileName(name string) string {
	name = strings.ReplaceAll(name, `\`, "/")
	name = path.Base(name)

	name = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, name)
	name = strings.TrimSpace(name)

	if name == "." || name == "/" {
		return ""
	}
	for len(name) > maxOriginalNameLength || !utf8.ValidString(name) {
		_, size := utf8.DecodeLastRuneInString(name)
		name = name[:len(name)-size]
	}
	return name
}

func (s *Service) storeUpload(dormitoryFolder string, upload *FileUpload) (*StoredFile, error) {
	ext, mime, err := validateUpload(upload)
	if err != nil {
		return nil, err
	}

	key, err := s.files.Save("documents/"+dormitoryFolder, ext, bytes.NewReader(upload.Data))
	if err != nil {
		return nil, err
	}

	return &StoredFile{
		Path: key,
		Name: cleanFileName(upload.Name),
		Mime: mime,
		Size: int64(len(upload.Data)),
	}, nil
}
