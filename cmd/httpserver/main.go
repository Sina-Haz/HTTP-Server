package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"sina.http/internal/request"
	"sina.http/internal/response"
	"sina.http/internal/server"
)

const port = 42069

func server_handler(rw *response.Writer, req *request.Request) {
	resource := req.RequestLine.RequestTarget
	var body []byte
	switch resource {
	case "/yourproblem":
		body = []byte("your problem is not my problem\n")
		rw.WriteStatusLine(400)
		rw.WriteHeaders(response.GetDefaultHeaders(len(body)))
	case "/myproblem":
		body = []byte("Oops my bad\n")
		rw.WriteStatusLine(500)
		rw.WriteHeaders(response.GetDefaultHeaders(len(body)))
	default:
		body = []byte("All good! \n")
		rw.WriteStatusLine(200)
		rw.WriteHeaders(response.GetDefaultHeaders(len(body)))
	}
	rw.WriteBody(body)
}

func main() {
	server, err := server.Serve(port, server_handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM) // relays SIGINT (ctrl C)+SIGNTERM (used by system tools like PKILL)signals to the channel
	<-sigChan                                               // blocks and waits for a signal to arrive to the channel (will if ctrl C or kill input by terminal b/c line above)
	log.Println("Server gracefully stopped")
}
