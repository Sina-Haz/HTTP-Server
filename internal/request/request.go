package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
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
	bodyState   = "body"
	finalState  = "done"
)

type Request struct {
	RequestLine RequestLine
	headers     headers.Headers
	Body        []byte
	state       string
}

func (r Request) Print() {
	fmt.Println("Request line:")
	fmt.Printf("- Method: %s\n", r.RequestLine.Method)
	fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
	fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)
	fmt.Println("Headers:")
	for key, value := range r.headers {
		fmt.Printf("- %s: %s\n", key, value)
	}
	fmt.Printf("Body:\n %s\n", string(r.Body))
	fmt.Println()
}

func (r *Request) onePass(unparsed_data []byte) (int, error) {
	var parsedN int
	switch r.state {
	case initState:
		// when use :=, overwrites n to be in local scope instead of updating the n from function level
		// Therefore we will refactor to add n to function level-val which will be renamed parsedN
		rl, n, err := parseRequestLine(unparsed_data)

		if err != nil {
			return n, errors.Join(fmt.Errorf("unable to parse request line, data passed was: %q", unparsed_data), err)
		}
		if n == 0 {
			break
		}
		r.RequestLine = *rl
		r.state = headerState
		parsedN = n
	case headerState:
		n, done, err := r.headers.Parse(unparsed_data)
		if err != nil {
			return n, errors.Join(fmt.Errorf("unable to parse headers data passed was: %q", unparsed_data), err)
		}
		if n == 0 {
			break
		}
		if done == true {
			r.state = bodyState
		}
		parsedN = n
	case bodyState:
		content_len := r.headers.Get("content-length")
		if content_len == "" || content_len == "0" {
			r.state = finalState // assume no body to parse and we will finish
		}
		conLen, err := strconv.Atoi(content_len)
		if err != nil {
			return parsedN, errors.Join(fmt.Errorf("content-length header value could not be parsed as a string, header value = %q", r.headers.Get("content-length")), err)
		}
		if len(unparsed_data) < conLen {
			break // parsedN should be 0 and tells us to read more bytes in. If body shorter than conLen then call to reader.Read() will eventually hit EOF
		}
		r.Body = unparsed_data[:conLen]
		r.state = finalState
	case finalState:
		break
	default:
		panic("Skill issue")
	}
	return parsedN, nil
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
	for r.state != finalState {
		parsedN, err := r.onePass(data[read:])
		if err != nil {
			return read, err
		}
		if parsedN == 0 {
			break // signal that we need to pop up to RequestFromReader fn and read more data
		}
		read += parsedN
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
