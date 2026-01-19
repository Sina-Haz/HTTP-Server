package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"sina.http/internal/request"
	"sina.http/internal/response"
	"sina.http/internal/server"
)

const port = 42069

func server_handler(w io.Writer, req *request.Request) {
	resource := req.RequestLine.RequestTarget
	var resp *response.Response
	switch resource {
	case "/yourproblem":
		resp = response.CreateResponse(400, []byte("your problem is not my problem\n"))
	case "/myproblem":
		resp = response.CreateResponse(500, []byte("Oops my bad\n"))
	default:
		resp = response.CreateResponse(200, []byte("All good! \n"))
	}
	resp.Write(w)
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
