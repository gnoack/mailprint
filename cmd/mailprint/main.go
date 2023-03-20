package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnoack/mailprint"
	"github.com/gnoack/picon"
	"github.com/landlock-lsm/go-landlock/landlock"
	llsys "github.com/landlock-lsm/go-landlock/landlock/syscall"
	"gopkg.in/pipe.v2"
)

var (
	pageFormat = flag.String("page_format", "A4", "Page format (A4, letter, ...), as supported by groff(1)")
	cc         = flag.Bool("cc", true, "Whether to show the CC headers.")
	facePicon  = flag.Bool("face.picon", true, "Whether to look up picon profile pictures")
	outFormat  = flag.String("output.format", "pdf", "Output format (one of 'pdf', 'mom')")
)

var usage = `Mailprint Deluxe! 📬

Profile pictures (picons)
	The picons archive can be found at https://kinzler.com/picons/ftp/,
	or installed with the "picons" package on Debian.

	The unpacked picons will be looked up in ~/.picons or /usr/share/picons.

Other output formats
	By default, Mailprint runs groff for you and looks up profile pictures.
	You can turn off both by setting the following flags:

	    --face.picon=false --output.format=mom

	The groff output is in "Mom" format and may contain UTF-8.
	To format it as PDF, run it through preconv and groff like so:

	    mailprint --face.picon --output.format=mom | preconv | groff -Tpdf -mom

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

	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	tmpdir, err := os.MkdirTemp("", "Mailprint*")
	if err != nil {
		return err
	}
	os.Setenv("TMPDIR", tmpdir) // Used by ImageMagick
	defer os.RemoveAll(tmpdir)

	logoPdfOut, err := os.CreateTemp(tmpdir, "MailprintFace*.pdf")
	if err != nil {
		return err
	}
	defer os.Remove(logoPdfOut.Name())

	err = landlock.V3.BestEffort().RestrictPaths(
		// Deleting tmpdir
		landlock.PathAccess(llsys.AccessFSRemoveFile, filepath.Dir(tmpdir)),
		// Icon lookup
		landlock.RODirs(
			"/usr/share/picons", // where it's installed on Debian
			filepath.Join(os.Getenv("HOME"), ".picons"),
		).IgnoreIfMissing(),
		// General binary invocations
		landlock.RODirs(strings.Split(os.Getenv("PATH"), ":")...).IgnoreIfMissing(),
		landlock.RODirs("/usr", "/lib"),
		// ImageMagick
		landlock.RWDirs(tmpdir),
		landlock.RODirs("/etc"),
		// Groff (to read the PDF profile image)
		landlock.ROFiles(logoPdfOut.Name()),
	)
	if err != nil {
		return fmt.Errorf("landlock: %w", err)
	}

	email, err := mailprint.Parse(os.Stdin)
	if err != nil {
		return err
	}

	if len(email.From) < 1 {
		return fmt.Errorf("Missing 'From' header")
	}
	ok, err := lookupIconPdf(email.From[0].Address, logoPdfOut)
	if err != nil {
		return fmt.Errorf("Looking up face: %w", err)
	}
	logoPdf := logoPdfOut.Name()
	if !ok {
		logoPdf = ""
	}

	if !*cc {
		email.Cc = nil
	}

	var groffbuf bytes.Buffer
	err = mailprint.Render(email, *pageFormat, logoPdf, &groffbuf)
	if err != nil {
		return fmt.Errorf("mailprint.Render: %w", err)
	}

	switch *outFormat {
	case "pdf":
		o, err := pipe.CombinedOutput(pipe.Line(
			pipe.Read(&groffbuf),
			pipe.Exec("preconv"),
			pipe.Exec("groff", "-mom", "-Tpdf"),
			pipe.Write(os.Stdout),
		))
		if err != nil {
			return fmt.Errorf("Pipeline failed: %v", string(o))
		}
	case "mom":
		_, err := io.Copy(os.Stdout, &groffbuf)
		if err != nil {
			return fmt.Errorf("io.Copy: %w", err)
		}
	default:
		return fmt.Errorf("unknown output format %q", *outFormat)
	}
	return nil
}

// Look up the face and write a PDF to w. Return true if w was written.
func lookupIconPdf(email string, w io.Writer) (ok bool, err error) {
	if !*facePicon {
		return false, nil
	}

	filename, ok := picon.Lookup(email)
	if filename == "" || !ok {
		return false, nil
	}

	o, err := pipe.CombinedOutput(pipe.Line(
		pipe.ReadFile(filename),
		pipe.Exec("convert", "--", "-", "PDF:-"),
		pipe.Write(w),
	))
	if err != nil {
		return false, fmt.Errorf("convert: %w; %v", err, string(o))
	}
	return true, nil
}
