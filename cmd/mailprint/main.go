package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gnoack/mailprint"
	"github.com/gnoack/picon"
	"github.com/landlock-lsm/go-landlock/landlock"
)

var (
	pageFormat = flag.String("page_format", "A4", "Page format (A4, letter, ...)")
	cc         = flag.Bool("cc", true, "Whether to show the CC headers.")
	facePicon  = flag.Bool("face.picon", true, "Whether to look up picon profile pictures")
)

var usage = `Mailprint Deluxe! 📬

Profile pictures (picons)
	The picons archive can be found at https://kinzler.com/picons/ftp/,
	or installed with the "picons" package on Debian.

	The unpacked picons will be looked up in ~/.picons or /usr/share/picons.

Happy printing! 💜
`

func main() {
	flag.Usage = func() {
		o := flag.CommandLine.Output()
		fmt.Fprintln(o, usage)
		fmt.Fprintf(o, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	// TODO: Maybe make this configurable in the future.
	var (
		headerFont  = "Serif"
		contentFont = "Mono"
	)
	fontPaths, err := mailprint.FindFonts(headerFont, contentFont)
	if err != nil {
		log.Fatalf("Font lookup: %v", err)
	}

	if err := enableSandbox(fontPaths); err != nil {
		log.Fatalf("Failed to enable opportunistic Landlock sandbox: %v", err)
	}

	if err := run(fontPaths); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func enableSandbox(fontPaths mailprint.FontPathOptions) error {
	home := os.Getenv("HOME")

	return landlock.V5.BestEffort().Restrict(
		// Icon lookup
		landlock.RODirs(
			"/usr/share/picons", // where it's installed on Debian
			filepath.Join(home, ".picons"),
		).IgnoreIfMissing(),
		// Fonts
		landlock.ROFiles(
			fontPaths.Content,
			fontPaths.Header,
			fontPaths.HeaderBold,
		).IgnoreIfMissing(),
	)
}

func run(fontPaths mailprint.FontPathOptions) error {
	email, err := mailprint.Parse(os.Stdin)
	if err != nil {
		return err
	}

	if len(email.From) < 1 {
		return fmt.Errorf("missing 'From' header")
	}
	logoPath, ok := lookupIconPath(email.From[0].Address)
	if !ok {
		logoPath = ""
	}

	if !*cc {
		email.Cc = nil
	}

	opts := &mailprint.RenderOptions{
		PageFormat: *pageFormat,
		LogoPath:   logoPath,
		FontPaths:  fontPaths,
	}
	err = mailprint.RenderPdf(email, opts, os.Stdout)
	if err != nil {
		return fmt.Errorf("mailprint.RenderPdf: %w", err)
	}

	return nil
}

// Look up the face and return the path to the image.
func lookupIconPath(email string) (path string, ok bool) {
	if !*facePicon {
		return "", false
	}

	filename, ok := picon.Lookup(email)
	if filename == "" || !ok {
		return "", false
	}
	return filename, true
}
