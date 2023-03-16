package mailprint

import (
	"io"
	"net/mail"
	"strings"
	"text/template"
	"time"
)

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{
	"processBody": processBody,
	"escape":      groffEscape,
}).Parse(`
{{define "person" -}}
  {{if .Name -}}
    {{.Name | escape}} <{{.Address | escape}}>
  {{- else -}}
    {{.Address | escape}}
  {{- end -}}
{{end -}}

{{- define "main" -}}
.TITLE "{{.Subject | escape}}"
.DOCTYPE    LETTER
.PRINTSTYLE TYPESET
.PAPER      {{.PageFormat}}
.
.T_MARGIN   2.5c
.B_MARGIN   3c
.
.LS 11
.
.PAGENUM_POS BOTTOM CENTER
.PAGENUM_HYPHENS OFF
.
.FOOTER_RECTO C "\*[PAGE#]"
.DOCHEADER OFF
.
.QUOTE_INDENT 0
.QUOTE_SIZE -4
.
.NEWCOLOR red         RGB #880000
.NEWCOLOR green       RGB #008800
.NEWCOLOR quotedmail  RGB #444488
.
.START
.FAMILY T
.PT_SIZE 9
\# XXX No idea why this is needed... :-/
.sp 0.0001c
\# Header
{{if .LogoPdfPath -}}
.PDF_IMAGE -R {{.LogoPdfPath}} 48p 48p
{{end -}}
.sp |2.5c
.SILENT
.PAD "\*[ST1]\fBSubject:\fR\*[ST1X]\*[FWD 6p]\*[ST2]#\*[ST2X]\*[FWD 54p]"
.SILENT OFF
.ST 1 R
.ST 2 L QUAD
.TAB 1
\fBSubject:\fR
.TN
{{.Subject | escape}}
{{range .From -}}
.TAB 1
\fBFrom:\fR
.TN
{{block "person" .}}{{end}}
{{end -}}
{{range .To -}}
.TAB 1
\fBTo:\fR
.TN
{{block "person" .}}{{end}}
{{end -}}
{{range .Cc -}}
.TAB 1
\fBCc:\fR
.TN
{{block "person" .}}{{end}}
{{end -}}
{{if .Date -}}
.TAB 1
\fBDate:\fR
.TN
{{.Date}}
{{end -}}
.TQ
.DRH
.sp 2
\# xxx why is this needed for the text body?
.\" Body
.QUOTE
.CODE
{{.TextBody | processBody}}
.QUOTE OFF
{{end -}}
`))

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
			for pos%8 != 0 {
				out.WriteByte(' ')
				pos++
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
