package headers

import (
	"bytes"
	"fmt"
	"strings"

	"sina.http/internal/request"
)

var BAD_HEADER = fmt.Errorf("Malformed HTTP header")

type Headers map[string]string

// can technically pass in by value here because maps in go are reference types so changing h
// will still update the key-value pairs in the map even though its passed by value
func (h Headers) Parse(data []byte) (int, bool, error) {
	endHeaderIdx := bytes.Index(data, bytes.Repeat(request.CRLF, 2))
	if endHeaderIdx == -1 {
		return 0, false, nil // still waiting for \r\n\r\n which signals end of HTTP headers
	}
	headers := bytes.SplitSeq(data[:endHeaderIdx], request.CRLF)
	for hdr := range headers {
		trimmed := bytes.TrimSpace(hdr)
		fn, fv, ok := strings.Cut(string(trimmed), ":")
		if !ok || strings.Contains(fn, " ") {
			return 0, false, BAD_HEADER
		}
		fv = strings.TrimSpace(fv)
		h[fn] = fv
	}
	return endHeaderIdx + len(request.CRLF)*2, true, nil
}

func NewHeaders() Headers {
	return make(Headers)
}
