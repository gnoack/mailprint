# Mailprint Deluxe!

Installation:

```
go install github.com/gnoack/mailprint/...@latest
```

In your muttrc:

```
set print_command="mailprint | lpr"
```

(or pipe it to a PDF viewer if you want to look at it first)
