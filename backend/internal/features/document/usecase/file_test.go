package usecase

import (
	"bytes"
	"errors"
	"testing"

	documentdomain "apihorpug/internal/features/document/domain"
)

var (
	pdfBytes  = []byte("%PDF-1.4\n1 0 obj\n<<>>\nendobj\n")
	pngBytes  = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00")
	zipBytes  = []byte("PK\x03\x04\x14\x00\x00\x00\x08\x00")
	htmlBytes = []byte("<!DOCTYPE html><html><script>alert(1)</script></html>")
)

func TestValidateUpload(t *testing.T) {
	tests := []struct {
		name     string
		file     FileUpload
		wantExt  string
		wantMime string
		wantErr  error
	}{
		{"pdf", FileUpload{Name: "lease.pdf", Data: pdfBytes}, ".pdf", "application/pdf", nil},
		{"png upper-case ext", FileUpload{Name: "ID.PNG", Data: pngBytes}, ".png", "image/png", nil},
		{"docx is a zip", FileUpload{Name: "contract.docx", Data: zipBytes}, ".docx", allowedFileTypes[".docx"].mime, nil},
		{"html renamed to pdf", FileUpload{Name: "evil.pdf", Data: htmlBytes}, "", "", documentdomain.ErrUnsupportedFileType},
		{"html renamed to png", FileUpload{Name: "evil.png", Data: htmlBytes}, "", "", documentdomain.ErrUnsupportedFileType},
		{"png renamed to pdf", FileUpload{Name: "photo.pdf", Data: pngBytes}, "", "", documentdomain.ErrUnsupportedFileType},
		{"html html", FileUpload{Name: "page.html", Data: htmlBytes}, "", "", documentdomain.ErrUnsupportedFileType},
		{"svg not allowed", FileUpload{Name: "logo.svg", Data: []byte("<svg xmlns='http://www.w3.org/2000/svg'/>")}, "", "", documentdomain.ErrUnsupportedFileType},
		{"exe", FileUpload{Name: "run.exe", Data: []byte("MZ\x90\x00")}, "", "", documentdomain.ErrUnsupportedFileType},
		{"no extension", FileUpload{Name: "README", Data: pdfBytes}, "", "", documentdomain.ErrUnsupportedFileType},
		{"empty", FileUpload{Name: "a.pdf", Data: nil}, "", "", documentdomain.ErrEmptyFile},
		{"too large", FileUpload{Name: "big.pdf", Data: append(append([]byte{}, pdfBytes...), make([]byte, MaxUploadSize)...)}, "", "", documentdomain.ErrFileTooLarge},
		{"windows path", FileUpload{Name: `C:\scans\lease.pdf`, Data: pdfBytes}, ".pdf", "application/pdf", nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ext, mime, err := validateUpload(&tc.file)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if ext != tc.wantExt || mime != tc.wantMime {
				t.Fatalf("got (%q, %q), want (%q, %q)", ext, mime, tc.wantExt, tc.wantMime)
			}
		})
	}
}

func TestCleanFileName(t *testing.T) {
	tests := map[string]string{
		"lease.pdf":           "lease.pdf",
		`C:\scans\lease.pdf`:  "lease.pdf",
		"../../etc/passwd":    "passwd",
		"สัญญาเช่า.pdf":       "สัญญาเช่า.pdf",
		"bad\x00name\r\n.pdf": "badname.pdf",
		"..":                  "..",
		"/":                   "",
		"":                    "",
	}
	for in, want := range tests {
		if got := cleanFileName(in); got != want {
			t.Errorf("cleanFileName(%q) = %q, want %q", in, got, want)
		}
	}

	long := string(bytes.Repeat([]byte("ก"), 200)) + ".pdf" // 600+ bytes
	if got := cleanFileName(long); len(got) > maxOriginalNameLength {
		t.Errorf("long name not truncated: %d bytes", len(got))
	}
}
