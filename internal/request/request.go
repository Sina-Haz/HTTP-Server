package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"sina.http/internal/headers"
)

var BAD_REQ_LINE = fmt.Errorf("malformed request line")
var UNSUPPORTED_HTTP_VERSION = fmt.Errorf("Unsupported http version")
var CRLF = []byte("\r\n")

// using string enum for better readability
const (
	initState   = "init"
	headerState = "headers"
	finalState  = "done"
)

type Request struct {
	RequestLine RequestLine
	headers     headers.Headers
	state       string
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		switch r.state {
		case initState:
			rl, n, err := parseRequestLine(data[read:])

			if err != nil {
				return n, errors.Join(fmt.Errorf("unable to parse request line, data passed was: %q", data[read:]), err)
			}
			if n == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += n
			r.state = headerState
		case headerState:
			n, done, err := r.headers.Parse(data[read:])
			if err != nil {
				return n, errors.Join(fmt.Errorf("unable to parse headers data passed was: %q", data[read:]), err)
			}
			if n == 0 {
				break outer // need to read more
			}
			read += n
			if done == true {
				r.state = finalState
			}
		case finalState:
			break outer
		}
	}
	return read, nil
}

type RequestLine struct {
	Method        string // ex. GET
	RequestTarget string // ex. /admin/login
	HttpVersion   string // ex. HTTP/1.1
}

func newRequest() *Request {
	return &Request{
		headers: headers.NewHeaders(),
		state:   initState,
	}
}

// reqline = method SP request-target SP HTTP-version
func parseRequestLine(msg []byte) (*RequestLine, int, error) {
	idx := bytes.Index(msg, CRLF)
	if idx == -1 {
		return nil, 0, nil // not an error but no request line detected
	}
	reqline := string(msg[:idx])

	parts := strings.Split(reqline, " ")
	if len(parts) != 3 {
		return nil, idx, BAD_REQ_LINE
	}

	http, version, found := strings.Cut(parts[2], "/")
	if !found || http != "HTTP" || version != "1.1" {
		return nil, idx, UNSUPPORTED_HTTP_VERSION
	}

	rl := RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   version,
	}

	return &rl, idx + len(CRLF), nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, 1024)
	req := newRequest()
	bufLen := 0
	for req.state != finalState {
		n, err := reader.Read(buf[bufLen:])
		// TODO: handle this error better
		if err != nil {
			return nil, errors.Join(fmt.Errorf("reader.Read() error while parsing request"), err)
		}
		bufLen += n
		parsed, err := req.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}
		// overwrite parsed data instead of setting buf = buf[parsed:]
		// which would cause buffer to get smaller every iteration
		copy(buf, buf[parsed:bufLen])
		bufLen -= parsed
	}
	if string(buf[:bufLen]) != "" {
		return req, fmt.Errorf("Request reached final state but parsed data != read data, here is remaining data in buffer: \n%s\n", string(buf[:bufLen]))
	}

	return req, nil
}
