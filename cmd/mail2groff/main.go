package main

import (
	"flag"
	"log"
	"os"

	"github.com/gnoack/mailprint"
)

var (
	cc         = flag.Bool("cc", true, "Show CC headers")
	pageFormat = flag.String("page_format", "A4", "Page format")
	logoPdf    = flag.String("logo_pdf", "", "PDF file with top-right logo")
)

func main() {
	flag.Parse()

	em, err := mailprint.Parse(os.Stdin)
	if err != nil {
		log.Fatalf("Parsing mail: %v", err)
	}

	// Clear out the CC list if it's too annoying to look at.
	if !*cc {
		em.Cc = nil
	}

	err = mailprint.Render(em, *pageFormat, *logoPdf, os.Stdout)
	if err != nil {
		log.Fatalf("Rendering error: %v", err)
	}
}
