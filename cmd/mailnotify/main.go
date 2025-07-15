package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/gnoack/mailprint"
	"github.com/gnoack/picon"
)

var usage = `
mailnotify - send a notification about an email

Pipe the email on stdin!
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
	email, err := mailprint.Parse(os.Stdin)
	if err != nil {
		return err
	}

	iconPath, iconOK := picon.Lookup(email.From[0].Address)

	args := []string{}
	if iconOK {
		args = append(args, "-i", iconPath)
	}
	args = append(args, email.Subject)

	return exec.Command("notify-send", args...).Run()
}
