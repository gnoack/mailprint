package mailprint

import (
	"embed"
	"io"
	"net/mail"
	"strings"
	"text/template"
	"time"
)

//go:embed templates/*
var templateFiles embed.FS

var tmpl = template.Must(template.New("main").Funcs(template.FuncMap{
	"processBody": processBody,
	"escape":      groffEscape,
}).ParseFS(templateFiles, "templates/*.mom"))

func Render(em *Email, pageFormat, logoPdf string, w io.Writer) error {
	return tmpl.ExecuteTemplate(w, "main", struct {
		// Page
		PageFormat  string
		LogoPdfPath string
		// Headers
		Subject string
		From    []*mail.Address
		To      []*mail.Address
		Cc      []*mail.Address
		Date    time.Time
		// Content
		TextBody string
	}{
		PageFormat:  pageFormat,
		LogoPdfPath: logoPdf,
		Subject:     em.Subject,
		From:        em.From,
		To:          em.To,
		Cc:          em.Cc,
		Date:        em.Date,
		TextBody:    em.TextBody,
	})
}

func groffEscape(s string) string {
	prefix := ""
	if strings.HasPrefix(s, ".") {
		prefix = `\&`
	}
	return prefix + strings.ReplaceAll(s, `\`, `\\`)
}

func processBody(input string) string {
	var out strings.Builder

	color := "black"
	setColor := func(c string) {
		if color == c {
			return
		}
		//out.WriteString(`\*[` + c + `]`)
		out.WriteString(".COLOR " + c + "\n")
		color = c
	}
	pos := 0
	for _, b := range []byte(input) {
		if pos == 0 {
			switch b {
			case '>':
				setColor("quotedmail")
			case '+':
				setColor("green")
			case '-':
				setColor("red")
			default:
				setColor("black")
			}
			if b == '.' {
				// Needs escaping
				out.WriteString(`\&`)
			}
		}
		if b == '\\' {
			// Needs escaping
			out.WriteString(`\`)
		}
		switch b {
		case '\t':
			for {
				out.WriteByte(' ')
				pos++

				if pos%8 == 0 {
					break
				}
			}
		case '\n':
			out.WriteByte(b)
			pos = 0
		default:
			out.WriteByte(b)
			pos++
		}
	}

	return out.String()
}
