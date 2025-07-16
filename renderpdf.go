package mailprint

import (
	"fmt"
	"io"
	"net/mail"
	"strings"
	"time"

	"github.com/signintech/gopdf"
)

const margin = 50.0

type RenderOptions struct {
	PageFormat string
	LogoPath   string
	FontPaths  FontPathOptions
}

type FontPathOptions struct {
	Header     string
	HeaderBold string
	Content    string
}

func RenderPdf(em *Email, opts *RenderOptions, w io.Writer) error {
	pdf := gopdf.GoPdf{}
	var pageSize *gopdf.Rect
	switch strings.ToLower(opts.PageFormat) {
	case "letter":
		pageSize = gopdf.PageSizeLetter
	case "a4":
		pageSize = gopdf.PageSizeA4
	default:
		return fmt.Errorf("unsupported page size %q", opts.PageFormat)
	}
	pdf.Start(gopdf.Config{PageSize: *pageSize})
	pdf.AddPage()

	err := pdf.AddTTFFont("header-bold", opts.FontPaths.HeaderBold)
	if err != nil {
		return fmt.Errorf("failed to add font: %w", err)
	}
	err = pdf.AddTTFFont("header", opts.FontPaths.Header)
	if err != nil {
		return fmt.Errorf("failed to add font: %w", err)
	}
	err = pdf.AddTTFFont("content", opts.FontPaths.Content)
	if err != nil {
		return fmt.Errorf("failed to add font: %w", err)
	}

	avatarSize := &gopdf.Rect{W: 50, H: 50}
	// Draw avatar
	if opts.LogoPath != "" {
		// Place image in top right corner
		err := pdf.Image(opts.LogoPath, pageSize.W-margin-avatarSize.W, margin, avatarSize)
		if err != nil {
			// Silently ignore image errors for now
		}
	}

	// Draw header
	const headerFontSize = 9

	type Header struct{ Name, Value string }
	headers := []Header{
		{"Subject:", em.Subject},
		{"From:", formatAddresses(em.From)},
		{"To:", formatAddresses(em.To)},
	}
	if len(em.Cc) > 0 {
		headers = append(headers, Header{"Cc:", formatAddresses(em.Cc)})
	}
	headers = append(headers, Header{"Date:", em.Date.Format(time.RFC1123Z)})

	// Determine the width of the longest header name
	var maxNameWidth float64
	pdf.SetFont("header-bold", "", headerFontSize)
	for _, h := range headers {
		width, err := pdf.MeasureTextWidth(h.Name)
		if err != nil {
			return err
		}
		if width > maxNameWidth {
			maxNameWidth = width
		}
	}

	pdf.SetY(margin)
	valueX := margin + maxNameWidth + 5 // 5 points gap

	for _, h := range headers {
		// Draw header name (bold, right-aligned)
		pdf.SetFont("header-bold", "", headerFontSize)
		y := pdf.GetY()
		pdf.SetX(margin)
		pdf.CellWithOption(
			&gopdf.Rect{W: maxNameWidth},
			h.Name,
			gopdf.CellOption{Align: gopdf.Right},
		)

		// Draw header value (regular, wrapped)
		pdf.SetFont("header", "", headerFontSize)
		valueWidth := pageSize.W - valueX - margin
		if opts.LogoPath != "" {
			valueWidth -= avatarSize.W
		}

		pdf.SetXY(valueX, y)
		pdf.MultiCellWithOption(
			&gopdf.Rect{W: valueWidth},
			h.Value,
			gopdf.CellOption{
				BreakOption: &gopdf.BreakOption{
					Mode:           gopdf.BreakModeIndicatorSensitive,
					BreakIndicator: ' ',
				},
			},
		)
	}

	// Draw line
	currentY := pdf.GetY() + 5 // smaller gap before the line
	pdf.SetLineWidth(0.5)
	pdf.Line(margin, currentY, pageSize.W-margin, currentY)
	pdf.SetY(currentY + 20)

	// Draw body
	processBodyPdf(em.TextBody, &pdf, pageSize)

	_, err = pdf.WriteTo(w)
	return err
}

func formatAddresses(addresses []*mail.Address) string {
	var parts []string
	for _, addr := range addresses {
		if addr.Name != "" {
			parts = append(parts, fmt.Sprintf("%v <%v>", addr.Name, addr.Address))
		} else {
			parts = append(parts, addr.Address)

		}
	}
	return strings.Join(parts, ", ")
}

func processBodyPdf(input string, pdf *gopdf.GoPdf, pageSize *gopdf.Rect) {
	lines := strings.Split(input, "\n")
	lineHeight := 12.0
	maxWidth := pageSize.W - 2*margin

	// Set font for body
	pdf.SetFont("content", "", 9)
	pdf.SetX(margin)

	for _, line := range lines {
		if pdf.GetY() > pageSize.H-margin {
			pdf.AddPage()
			pdf.SetY(margin)
			pdf.SetX(margin)
		}

		var r, g, b uint8
		switch {
		case strings.HasPrefix(line, ">"):
			r, g, b = 100, 100, 100 // grey for quoted
		case strings.HasPrefix(line, "+"):
			r, g, b = 0, 128, 0 // green for added
		case strings.HasPrefix(line, "-"):
			r, g, b = 255, 0, 0 // red for removed
		default:
			r, g, b = 0, 0, 0 // black for normal
		}
		pdf.SetTextColor(r, g, b)

		pdf.CellWithOption(
			&gopdf.Rect{W: maxWidth, H: lineHeight},
			truncate(line, maxWidth, pdf),
			gopdf.CellOption{
				//Border:      gopdf.AllBorders,
				Float: gopdf.Bottom,
			},
		)
	}
}

func truncate(text string, maxWidth float64, pdf *gopdf.GoPdf) string {
	ls, _ := pdf.SplitText(text, maxWidth)
	if len(ls) == 0 {
		// Likely empty input?
		return ""
	}
	return ls[0]
}
