package headers

import (
	"bytes"
	"fmt"
	"strings"
)

var BAD_HEADER = fmt.Errorf("Malformed HTTP header")
var crlf = []byte("\r\n")

type Headers map[string]string

// Will take in the data and only parse one header at a time
// Returns number of bytes parsed, done == parsed all headers, error
// can pass in Headers by value as maps are reference types so h is copy of ptr to hashmap
func (h Headers) Parse(data []byte) (int, bool, error) {
	crlfIdx := bytes.Index(data, crlf)
	if crlfIdx == -1 {
		return 0, false, nil
	}
	if crlfIdx == 0 {
		return len(crlf), true, nil
	}
	parsedN := crlfIdx + len(crlf)
	hdr := string(data[:crlfIdx])
	trimmed := strings.TrimSpace(hdr) // trim whitespace before and after field name & field val
	fn, fv, ok := strings.Cut(trimmed, ":")
	if !ok || strings.Contains(fn, " ") {
		return 0, false, BAD_HEADER
	}
	// trim optional whitespace before and after field val
	fv = strings.TrimSpace(fv)
	h[fn] = fv

	return parsedN, false, nil
}

func NewHeaders() Headers {
	return make(Headers)
}
