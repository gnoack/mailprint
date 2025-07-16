package mailprint

import (
	"bytes"
	"net/mail"
	"testing"
	"time"
)

func TestRenderPdfBasic(t *testing.T) {
	// Create a sample Email struct
	fromAddr, _ := mail.ParseAddress("sender@example.com")
	toAddr, _ := mail.ParseAddress("recipient@example.com")

	em := &Email{
		Subject:  "Test Subject",
		From:     []*mail.Address{fromAddr},
		To:       []*mail.Address{toAddr},
		Cc:       []*mail.Address{},
		Date:     time.Now(),
		TextBody: "This is a short test email body.\n\nIt has multiple lines.",
	}

	// Create a buffer to write the PDF output to
	var buf bytes.Buffer

	// Call RenderPdf
	opts := &RenderOptions{
		PageFormat: "A4",
		FontPaths:  Fonts(t),
	}
	err := RenderPdf(em, opts, &buf)
	if err != nil {
		t.Fatalf("RenderPdf failed: %v", err)
	}

	// Get the PDF content from the buffer
	pdfContent := buf.Bytes()

	// Assert that the output is not empty
	if len(pdfContent) == 0 {
		t.Error("RenderPdf produced empty output")
	}

	// Assert that the output has a PDF header
	if !bytes.HasPrefix(pdfContent, []byte("%PDF-")) {
		t.Errorf("RenderPdf output does not start with PDF header. Got: %s...", string(pdfContent[:10]))
	}
}

func Fonts(t *testing.T) FontPathOptions {
	f, err := FindFonts("Liberation Serif", "Liberation Mono")
	if err != nil {
		t.Fatalf("Finding fonts: %v", err)
	}
	return f
}
