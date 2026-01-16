package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"sina.http/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port)
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
