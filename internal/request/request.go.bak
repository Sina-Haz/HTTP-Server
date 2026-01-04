package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

var BAD_REQ_LINE = fmt.Errorf("malformed request line")
var UNSUPPORTED_HTTP_VERSION = fmt.Errorf("Unsupported http version")
var CRLF = []byte("\r\n")

// using string enum for better readability
const (
	initState  = "init"
	finalState = "done"
)

type Request struct {
	RequestLine RequestLine
	state       string
}

// data should accumulate until we return nonzero int, then set data = data[n:] to avoid reparsing
func (r *Request) parse(data []byte) (int, error) {
	var n int
	if r.RequestLine == (RequestLine{}) {
		rl, n, err := ParseRequestLine(data)
		if err != nil {
			return n, errors.Join(fmt.Errorf("unable to parse request line, data passed was: %q", data), err)
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *rl
		r.state = finalState
	}
	return n, nil
}

type RequestLine struct {
	Method        string // ex. GET
	RequestTarget string // ex. /admin/login
	HttpVersion   string // ex. HTTP/1.1
}

func newRequest() *Request {
	return &Request{
		state: initState,
	}
}

// reqline = method SP request-target SP HTTP-version
func ParseRequestLine(msg []byte) (*RequestLine, int, error) {
	idx := bytes.Index(msg, CRLF)
	if idx == -1 {
		return nil, 0, nil // not an error but no request line detected
	}
	reqline := string(msg[:idx])

	parts := strings.Split(reqline, " ")
	if len(parts) != 3 {
		return nil, idx, BAD_REQ_LINE
	}

	http, version, _ := strings.Cut(parts[2], "/")
	if http != "HTTP" || version != "1.1" {
		return nil, idx, UNSUPPORTED_HTTP_VERSION
	}

	rl := RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   version,
	}

	return &rl, idx, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, 1024)
	req := newRequest()
	bufIdx := 0
	for req.state != finalState {
		n, err := reader.Read(buf[bufIdx:])
		// TODO: handle this error better
		if err != nil {
			return nil, errors.Join(fmt.Errorf("read error while parsing request"), err)
		}
		bufIdx += n
		parsed, err := req.parse(buf)
		if err != nil {
			return nil, err
		}
		buf = buf[parsed:]
	}

	return req, nil
}
