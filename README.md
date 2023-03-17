# Mailprint Deluxe!

*Mailprint* is a replacement for `muttprint`: It reads an email on
stdin and writes a PDF on stdout.

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
