package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"sina.http/internal/request"
	"sina.http/internal/response"
)

type StatusCode = response.StatusCode

type HandlerError struct {
	status StatusCode
}

func (h HandlerError) Error() string {
	return fmt.Sprintf("HTTP/1.1 %s\r\n", string(h.status))
}

func MakeHandlerError(code int, msg string) *HandlerError {
	status := response.IntToStatusCode(code)
	return &HandlerError{
		status: status,
	}
}

// Assume that you write a proper response object to w if you return nil no error
type Handler func(w *response.Writer, req *request.Request)

// bind+listen to a port -> in a loop accept connections and handle each in a goroutine -> do until closed
type Server struct {
	closed   atomic.Bool
	listener net.Listener
}

func newServer(listener net.Listener) *Server {
	srv := &Server{closed: atomic.Bool{}, listener: listener}
	return srv
}

// Sets up a listener at specified port
// Returns a new Server and sets that server to listen in a separate goroutine
func Serve(port uint16, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	srv := newServer(listener)
	go srv.listen(handler)
	return srv, nil
}

// Sets closed to true and closes the server's listener binded to its port
func (s *Server) Close() error {
	s.closed.Store(true)
	err := s.listener.Close()
	return err
}

func (s *Server) listen(handler Handler) {
	for s.closed.Load() != true {
		conn, err := s.listener.Accept()

		// TODO: May need to change this to a 500 internal server error or something
		if err != nil {
			log.Fatalf("Server could not accept incoming connection, see error:\n%v ", err)
		}
		go s.handle(conn, handler)
	}
}

// This function takes in connection, reads to parse the request
// Upon parsing it will pass the request and new response.Writer to
// the user-defined handler. If error occurs in parsing -> sends back 400 w/ explanation
// If response.Writer goes to error state it will write back a 500 response
func (s *Server) handle(conn io.ReadWriteCloser, handler Handler) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	responseWriter := response.MakeResponseWriter(conn)
	if err != nil {
		body := fmt.Sprintf("Parsing the request caused the following error, sending 400: %s", err.Error())
		WriteDefaultResponse(responseWriter, 400, []byte(body))
		return
	}
	handler(responseWriter, req)
	if responseWriter.State == response.ResponseError {
		body := "User defined handler function incorrectly called response writer"
		WriteDefaultResponse(responseWriter, 500, []byte(body)) // TODO: may not be correct b/c what if the user-defined handler already wrote some stuff to the connection? may send back unparseable data
	}
}

// Convenience function for writing back a simple response without streaming data or using any special headers
// Will use response writer to write back a response with specified status and body and default headers
func WriteDefaultResponse(w *response.Writer, status int, body []byte) {
	w.State = response.ResponseInit
	w.WriteStatusLine(status)
	w.WriteHeaders(response.GetDefaultHeaders(len(body)))
	w.WriteBody(body)
}
