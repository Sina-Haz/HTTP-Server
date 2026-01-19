package response

import (
	"bytes"
	"fmt"
	"io"

	"sina.http/internal/headers"
)

// Status code = string(%d %s, code, reason-phrase)
type StatusCode string

const (
	OK                    StatusCode = "200 OK"
	BAD_REQUEST           StatusCode = "400 Bad Request"
	INTERNAL_SERVER_ERROR StatusCode = "500 Internal Server Error"
)

func IntToStatusCode(code int) StatusCode {
	var Status StatusCode
	switch code {
	case 200:
		Status = OK
	case 400:
		Status = BAD_REQUEST
	case 500:
		Status = INTERNAL_SERVER_ERROR
	default:
		Status = StatusCode(fmt.Sprintf("%d ", code))
	}
	return Status
}

type Response struct {
	version []byte
	code    int
	status  StatusCode
	headers headers.Headers
	body    []byte
}

func CreateResponse(code int, body []byte) *Response {
	stat := IntToStatusCode(code)
	hdr := GetDefaultHeaders(len(body))
	return &Response{
		version: []byte("HTTP/1.1"),
		code:    code,
		status:  stat,
		headers: hdr,
		body:    body,
	}
}

func (r Response) Write(w io.Writer) error {
	err := WriteStatusLine(w, r.version, r.code)
	if err != nil {
		return err
	}
	err = WriteHeaders(w, r.headers)
	if err != nil {
		return err
	}
	_, err = w.Write(r.body)
	return err
}

// version should be HTTP/1.1 (others unimplemented)
// code can be 200, 400, 500 with mesages or custom code w/ no message
func WriteStatusLine(w io.Writer, version []byte, code int) error {
	statCode := IntToStatusCode(code)
	parts := [][]byte{
		version, []byte(" "), []byte(statCode), []byte("\r\n"),
	}
	statLine := bytes.Join(parts, []byte(""))
	n, err := w.Write(statLine)
	if err != nil {
		return err
	}
	if n != len(statLine) {
		return fmt.Errorf("didn't write full status line")
	}
	return nil
}

// Adds default headers such as content length, connection, Content-type
// Other headers to add later are:
// Content-Encoding -> tells client how to decode compressed/encoded content
// Date -> useful for caching
// Cache-Control -> Tells client how to handle caching for requests and responses
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
