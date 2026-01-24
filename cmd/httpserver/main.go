package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"sina.http/internal/request"
	"sina.http/internal/response"
	"sina.http/internal/server"
)

const port = 42069
const msg1 = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`
const msg2 = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`
const msg3 = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

func serverHandler(rw *response.Writer, req *request.Request) {
	resource := req.RequestLine.RequestTarget
	hdrs := response.GetDefaultHeaders(0) // will change content len manually and update content type
	hdrs.Set("content-type", "text/html")
	switch {
	case resource == "/yourproblem":
		server.WriteDefaultResponse(rw, 400, msg1)
	case resource == "/myproblem":
		server.WriteDefaultResponse(rw, 500, msg2)
	case strings.HasPrefix(resource, "/httpbin"):
		proxyHandler(rw, req) // use sub-handler
	default:
		server.WriteDefaultResponse(rw, 200, msg3)
	}
}

// Given that a request is sent to server with target = /httpbin/x -> send request to httpbin.org/x
// And use chunked-encoding to stream back the response to client
func proxyHandler(rw *response.Writer, req *request.Request) {
	suffix := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	url := fmt.Sprintf("https://httpbin.org/%s", suffix)
	resp, err := http.Get(url)
	if err != nil {
		server.WriteDefaultResponse(rw, 400, fmt.Sprintf("GET on %s failed with error: %s\n", url, err.Error()))
		return
	}

	// Now read from resp.Body and stream it back using Chunked-Encoding
	rw.WriteStatusLine(200)
	hdrs := response.GetDefaultHeaders(0)
	hdrs.Remove("content-length")
	hdrs.Set("Transfer-encoding", "chunked")
	rw.WriteHeaders(hdrs)
	bufr := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(bufr)
		if n == 0 || errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Fatalf("Hit err while reading response body, %s", err.Error())
		}
		bufr = bufr[:n]
		rw.WriteChunkedBody(bufr)
	}
	rw.WriteChunkedBodyDone()

}

// Given that a request is sent to server with target = /httpbin/x -> send request to httpbin.org/x
// And use chunked-encoding to stream back the response to client
func proxyHandler2(rw *response.Writer, req *request.Request) {
	// first assert that request target has form httpbin/x
	bad_req_body := "Bad request, expected target to be to /httpbin/x where x is an integer\r\n"
	target := req.RequestLine.RequestTarget
	parts := strings.Split(target, "/")
	if len(parts) != 2 {
		server.WriteDefaultResponse(rw, 400, bad_req_body)
		return
	}
	n, err := strconv.Atoi(parts[1])
	if err != nil {
		server.WriteDefaultResponse(rw, 400, bad_req_body)
		return
	}
	resp, err := http.Get(fmt.Sprintf("https://httpbin.org/%d", n))
	if err != nil {
		server.WriteDefaultResponse(rw, 500, fmt.Sprintf("Error calling httpbin.org with x=%d, error message: %s", n, err.Error()))
		return
	}

	// Now read from resp.Body and stream it back using Chunked-Encoding
	rw.WriteStatusLine(200)
	hdrs := response.GetDefaultHeaders(0)
	hdrs.Remove("content-length")
	hdrs.Set("Transfer-encoding", "chunked")
	rw.WriteHeaders(hdrs)
	bufr := make([]byte, 32)
	for {
		n, err := resp.Body.Read(bufr)
		if err != nil {
			rw.WriteChunkedBodyDone()
			log.Fatalf("Hit err while reading response body, %s", err.Error())
		}
		if n == 0 {
			rw.WriteChunkedBodyDone()
			break
		}
		bufr = bufr[:n]
		rw.WriteChunkedBody(bufr)
	}
}

func main() {
	server, err := server.Serve(port, serverHandler)
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
