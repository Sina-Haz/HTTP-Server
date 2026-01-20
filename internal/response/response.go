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

type ResponseState string

const (
	responseInit    ResponseState = "status-line"
	responseHeaders ResponseState = "headers"
	responseBody    ResponseState = "body"
	responseDone    ResponseState = "done"
	ResponseError   ResponseState = "error"
)

// ResponseWriter struct will allow users to write
// The response in chunks (i.e. send off status line, then headers, then body)
// Will also keep internal state to ensure that users do this in correct order
type Writer struct {
	writer io.Writer
	State  ResponseState
}

func MakeResponseWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		State:  responseInit,
	}
}

// Will check if the response state is same as expected
// if not will set w.state to errorState and return informative error
func (w *Writer) stateCheck(expected ResponseState) error {
	if w.State != expected {
		err := fmt.Errorf("trying to write %s with response writer but state is: %s", expected, w.State)
		w.State = ResponseError
		return err
	}
	return nil
}

// code can be 200, 400, 500 with mesages or custom code w/ no message
func (w *Writer) WriteStatusLine(code int) error {
	err := w.stateCheck(responseInit)
	if err != nil {
		return err
	}

	statCode := IntToStatusCode(code)
	parts := [][]byte{
		[]byte("HTTP/1.1"), []byte(statCode), []byte("\r\n"),
	}
	statLine := bytes.Join(parts, []byte(" "))
	n, err := w.writer.Write(statLine)
	if err != nil {
		w.State = ResponseError
		return err
	}
	if n != len(statLine) {
		w.State = ResponseError
		return fmt.Errorf("didn't write full status line")
	}

	w.State = responseHeaders
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	err := w.stateCheck(responseHeaders)
	if err != nil {
		return err
	}

	for k, v := range headers {
		hdr := fmt.Appendf(nil, "%s: %s\r\n", k, v)
		n, err := w.writer.Write(hdr)
		if err != nil {
			w.State = ResponseError
			return err
		}
		if n != len(hdr) {
			w.State = ResponseError
			return fmt.Errorf("didn't write full status line")
		}
	}
	_, err = w.writer.Write([]byte("\r\n"))
	if err != nil {
		w.State = ResponseError
	}
	w.State = responseBody
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	err := w.stateCheck(responseBody)
	if err != nil {
		return 0, err
	}

	n, err := w.writer.Write(p)
	if err != nil {
		w.State = ResponseError
	}
	w.State = responseDone
	return n, nil
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

// ================================== OLD RESPONSE DESIGN
// Not going to do it this way because it would
// require entire response to be stored in memory
// before writing it back to the user, the real way internet works
// is that response is written bank usually in chunks
// in HTTP/1.1 we write those chunks directly over to the TCP connection
// But in 2.0/3.0 may be different (i.e. using multiplexing/caching idk)
// type Response struct {
// 	version []byte
// 	code    int
// 	status  StatusCode
// 	headers headers.Headers
// 	body    []byte
// }
//
// func CreateResponse(code int, body []byte) *Response {
// 	stat := IntToStatusCode(code)
// 	hdr := GetDefaultHeaders(len(body))
// 	return &Response{
// 		version: []byte("HTTP/1.1"),
// 		code:    code,
// 		status:  stat,
// 		headers: hdr,
// 		body:    body,
// 	}
// }
//
// func (r Response) Write(w io.Writer) error {
// 	err := WriteStatusLine(w, r.version, r.code)
// 	if err != nil {
// 		return err
// 	}
// 	err = WriteHeaders(w, r.headers)
// 	if err != nil {
// 		return err
// 	}
// 	_, err = w.Write(r.body)
// 	return err
// }
