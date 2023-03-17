# Mailprint Deluxe!

[*Mailprint*](https://gnoack.github.io/mailprint) is a replacement for
`muttprint`: It reads an email on stdin and writes a PDF on stdout.

## System requirements

* GNU roff (`groff`): to format the mail to PDF
* ImageMagick's `convert`: to convert profile pictures to PDF format

*groff* is usually already installed on Linux,
because it is used to format man pages.
ImageMagick is a popular package for doing
image manipulation on the command line.

## Installation and setup

Installation:

```
go install github.com/gnoack/mailprint/...@latest
```

In your muttrc:

```
set print_command="mailprint | lpr"
```

Alternatively, print it into a PDF viewer which first lets you look at
the result:

```
set print_command="mailprint | zathura -"
```
