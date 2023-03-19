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
)

var usage = `Mailprint Deluxe!

Faces!
	The picons archive can be found at https://kinzler.com/picons/ftp/,
	or installed with the "picons" package on Debian.

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
	logoPdfOut, err := os.CreateTemp("", "MailprintFace*.pdf")
	if err != nil {
		return err
	}
	defer os.Remove(logoPdfOut.Name())

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

	// XXX: Move the Landlock call before mail parsing and profile picture lookup.
	err = landlock.V3.BestEffort().RestrictPaths(
		landlock.RODirs(strings.Split(os.Getenv("PATH"), ":")...).IgnoreIfMissing(),
		landlock.RODirs("/usr", "/lib"),
		landlock.ROFiles(logoPdfOut.Name()),
		landlock.PathAccess(llsys.AccessFSRemoveFile, filepath.Dir(logoPdfOut.Name())),
	)
	if err != nil {
		return fmt.Errorf("landlock: %w", err)
	}

	if !*cc {
		email.Cc = nil
	}

	var groffbuf bytes.Buffer
	err = mailprint.Render(email, *pageFormat, logoPdf, &groffbuf)
	if err != nil {
		return fmt.Errorf("mailprint.Render: %w", err)
	}

	o, err := pipe.CombinedOutput(pipe.Line(
		pipe.Read(&groffbuf),
		pipe.Exec("preconv"),
		pipe.Exec("groff", "-mom", "-Tpdf"),
		pipe.Write(os.Stdout),
	))
	if err != nil {
		return fmt.Errorf("Pipeline failed: %v", string(o))
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
