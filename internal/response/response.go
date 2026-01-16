package response

import (
	"fmt"
	"io"

	"sina.http/internal/headers"
)

const (
	StatusCodeOk     = "200 OK"
	StatusCodeBadReq = "400 Bad Request"
	StatusCodeISE    = "500 Internal Server Error"
)

func WriteStatusLine(w io.Writer, statusCode int) error {
	var msg string
	switch statusCode {
	case 200:
		msg = fmt.Sprintf("HTTP/1.1 %s", StatusCodeOk)
	case 400:
		msg = fmt.Sprintf("HTTP/1.1 %s", StatusCodeBadReq)
	case 500:
		msg = fmt.Sprintf("HTTP/1.1 %s", StatusCodeISE)
	default:
		msg = fmt.Sprintf("HTTP/1.1 %d", statusCode)
	}
	msg += "\r\n"
	n, err := w.Write([]byte(msg))
	if err != nil {
		return err
	}
	if n != len(msg) {
		return fmt.Errorf("didn't write full status line")
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("content-length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-type", "text/plain")
	return h
}

func WriteHeaders(w io.Writer, h headers.Headers) error {
	for k, v := range h {
		hdr := fmt.Appendf(nil, "%s: %s\r\n", k, v)
		n, err := w.Write(hdr)
		if err != nil {
			return err
		}
		if n != len(hdr) {
			return fmt.Errorf("didn't write full status line")
		}
	}
	_, err := w.Write([]byte("\r\n"))
	return err
}
