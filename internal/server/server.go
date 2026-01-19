package server

import (
	"bytes"
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
type Handler func(w io.Writer, req *request.Request)

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

// Only actually write to connection here, we pass handler a bytes buffer
// and let it write to the buffer, once it returns we will write the buffer to the connection
func (s *Server) handle(conn io.ReadWriteCloser, handler Handler) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	bufr := bytes.NewBuffer(nil) // makes a dynamic array buffer to write to
	if err != nil {
		log.Println("Parsing the request caused the following error, sending 400", err.Error())
		response.WriteStatusLine(conn, []byte("HTTP/1.1"), 400)
		return
	}
	handler(bufr, req)
	conn.Write(bufr.Bytes())
}
