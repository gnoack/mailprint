package mailprint

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
	"time"
)

type Email struct {
	Subject  string
	From     []*mail.Address
	To       []*mail.Address
	Cc       []*mail.Address
	Date     time.Time
	TextBody string
}

func Parse(r io.Reader) (*Email, error) {
	msg, err := mail.ReadMessage(r)
	if err != nil {
		return nil, fmt.Errorf("mail.ReadMessage: %w", err)
	}

	h := msg.Header
	e := Email{}

	var wd mime.WordDecoder
	e.Subject, err = wd.DecodeHeader(h.Get("Subject"))
	if err != nil {
		return nil, err
	}

	e.Date, err = h.Date()
	if err != nil {
		return nil, err
	}

	e.To, err = h.AddressList("To")
	if err != nil {
		return nil, fmt.Errorf("header 'To': %w", err)
	}

	e.From, err = h.AddressList("From")
	if err != nil {
		return nil, fmt.Errorf("header 'From': %w", err)
	}

	e.Cc, err = h.AddressList("Cc")
	if err != nil && err != mail.ErrHeaderNotPresent {
		return nil, fmt.Errorf("header 'Cc': %w", err)
	}

	e.TextBody, err = parseTextBody(msg.Header.Get("Content-Type"), msg.Body)
	if err != nil {
		return nil, err
	}

	return &e, nil
}

var unsupportedContentTypeErr = errors.New("unsupported content type")

func parseTextBody(ctHdr string, r io.Reader) (string, error) {
	var (
		contentType string
		params      map[string]string
		err         error
	)
	if ctHdr == "" {
		contentType = "text/plain"
	} else {
		contentType, params, err = mime.ParseMediaType(ctHdr)
		if err != nil {
			return "", err
		}
	}

	switch contentType {
	case "text/plain":
		b, err := io.ReadAll(r)
		if err != nil {
			return "", err
		}
		return strings.TrimSuffix(string(b), "\n"), nil
	case "text/html":
		return "", unsupportedContentTypeErr
	case "multipart/alternative":
		fallthrough
	case "multipart/related":
		fallthrough
	case "multipart/report":
		fallthrough
	case "multipart/mixed":
		mpr := multipart.NewReader(r, params["boundary"])
		for {
			part, err := mpr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				return "", err
			}
			result, err := parseTextBody(part.Header.Get("Content-Type"), part)
			if err == unsupportedContentTypeErr {
				continue
			}
			return result, err
		}
		return "", errors.New("missing plain text content")
	default:
		return "", fmt.Errorf("unknown content type %q", contentType)
	}
}
