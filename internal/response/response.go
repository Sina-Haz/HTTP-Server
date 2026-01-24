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
	ResponseInit    ResponseState = "status-line"
	ResponseHeaders ResponseState = "headers"
	ResponseBody    ResponseState = "body"
	ResponseDone    ResponseState = "done"
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
		State:  ResponseInit,
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

// code can be 200, 400, 500 with messages or custom code w/ no message
func (w *Writer) WriteStatusLine(code int) error {
	err := w.stateCheck(ResponseInit)
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

	w.State = ResponseHeaders
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	err := w.stateCheck(ResponseHeaders)
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
	w.State = ResponseBody
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	err := w.stateCheck(ResponseBody)
	if err != nil {
		return 0, err
	}

	n, err := w.writer.Write(p)
	if err != nil {
		w.State = ResponseError
	}
	w.State = ResponseDone
	return n, nil
}

// Function to write one chunk of a body which is streamed via chunked-encoding
func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	err := w.stateCheck(ResponseBody)
	if err != nil {
		return 0, err
	}
	data := fmt.Sprintf("%X\r\n%s\r\n", len(p), p)
	n, err := w.writer.Write([]byte(data))
	if err != nil {
		w.State = ResponseError
	}
	return n, err
}

// Will simply write 0\r\n and line with only CRLF
// to signal that we are done sending chunks for body
func (w *Writer) WriteChunkedBodyDone() (int, error) {
	err := w.stateCheck(ResponseBody)
	if err != nil {
		return 0, err
	}
	data := "0\r\n\r\n"
	n, err := w.writer.Write([]byte(data))
	if err != nil {
		w.State = ResponseError
	}
	w.State = ResponseDone // TODO: may need to make this responseTrailers in the future, figure out how we are to handle trailers
	return n, err
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
