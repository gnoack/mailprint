# Mailprint Deluxe!

[*Mailprint*](https://gnoack.github.io/mailprint) is a replacement for
`muttprint`: It reads an email on stdin and writes a PDF on stdout.

![](https://gnoack.github.io/mailprint/mailprint.png)

## System requirements

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
