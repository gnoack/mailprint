# Mailprint Deluxe!

Mailprint is a replacement for `muttprint`: It reads an email on stdin
and writes a PDF on stdout.

Installation:

```
go install github.com/gnoack/mailprint/...@latest
```

In your muttrc:

```
set print_command="mailprint | lpr"
```

(or pipe it to a PDF viewer if you want to look at it first)
