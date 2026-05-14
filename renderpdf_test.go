package mailprint

import (
	"bytes"
	"net/mail"
	"strings"
	"testing"
	"time"
)

// Regression test: emails with no To header (only Cc) must parse and render.
func TestCcOnlyEmail(t *testing.T) {
	const raw = "From: sender@example.com\r\n" +
		"Cc: recipient@example.com\r\n" +
		"Subject: CC only\r\n" +
		"Date: Mon, 02 Jan 2006 15:04:05 -0700\r\n" +
		"\r\n" +
		"Body text.\r\n"

	em, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(em.To) != 0 {
		t.Errorf("expected empty To, got %v", em.To)
	}
	if len(em.Cc) != 1 {
		t.Fatalf("expected 1 Cc address, got %d", len(em.Cc))
	}

	var buf bytes.Buffer
	opts := &RenderOptions{
		PageFormat: "A4",
		FontPaths:  Fonts(t),
	}
	if err := RenderPdf(em, opts, &buf); err != nil {
		t.Fatalf("RenderPdf failed: %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) {
		t.Error("output is not a PDF")
	}
}

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
