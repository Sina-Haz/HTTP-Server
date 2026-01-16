package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"sina.http/internal/response"
)

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
func Serve(port uint16) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	srv := newServer(listener)
	go srv.listen()
	return srv, nil
}

// Sets closed to true and closes the server's listener binded to its port
func (s *Server) Close() error {
	s.closed.Store(true)
	err := s.listener.Close()
	return err
}

func (s *Server) listen() {
	for s.closed.Load() != true {
		conn, err := s.listener.Accept()

		// TODO: May need to change this to a 500 internal server error or something
		if err != nil {
			log.Fatalf("Server could not accept incoming connection, see error:\n%v ", err)
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn io.ReadWriteCloser) {
	statusCode := 200
	content := "Hello World\r\n"
	h := response.GetDefaultHeaders(len(content))
	// write status line then headers, then content
	err := response.WriteStatusLine(conn, statusCode)
	err = response.WriteHeaders(conn, h)
	_, err = conn.Write([]byte(content))
	if err != nil {
		log.Fatalf("error occurred while handling server connection: %v", err)
	}
	conn.Close()
}

// response := "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello World!"
// n, err := conn.Write([]byte(response))
// if err != nil {
// 	log.Fatalf("Error happened while writing response: \n%d", err)
// }
// if n != len(response) {
// 	fmt.Print("Error: server didn't write back full response")
// }

// func (s Server) Serve(port int) {
// 	strPort := fmt.Sprintf("%d", port)
// 	listener, err := net.Listen("tcp", strPort)
// 	if err != nil {
// 		log.Fatal("error", err)
// 	}
// 	defer listener.Close()
// 	fmt.Println()
// wait for listenek to connect
// conn, err := listener.Accept()
// }
