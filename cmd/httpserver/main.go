package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"sina.http/internal/headers"
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
		server.WriteDefaultResponse(rw, 400, []byte(msg1))
	case resource == "/myproblem":
		server.WriteDefaultResponse(rw, 500, []byte(msg2))
	case resource == "/video":
		video, err := os.ReadFile("./assets/vim.mp4")
		if err != nil {
			server.WriteDefaultResponse(rw, 500, []byte("Couldn't read the video file"))
			break
		}
		rw.WriteStatusLine(200)
		hdrs := response.GetDefaultHeaders(len(video))
		hdrs.Set("content-type", "video/mp4")
		rw.WriteHeaders(hdrs)
		rw.WriteBody(video)
	case strings.HasPrefix(resource, "/httpbin"):
		proxyHandler(rw, req) // use sub-handler
	default:
		server.WriteDefaultResponse(rw, 200, []byte(msg3))
	}
}

// Given that a request is sent to server with target = /httpbin/x -> send request to httpbin.org/x
// And use chunked-encoding to stream back the response to client
func proxyHandler(rw *response.Writer, req *request.Request) {
	suffix := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	url := fmt.Sprintf("https://httpbin.org/%s", suffix)
	resp, err := http.Get(url)
	if err != nil {
		msg := fmt.Sprintf("GET on %s failed with error: %s\n", url, err.Error())
		server.WriteDefaultResponse(rw, 400, []byte(msg))
		return
	}

	// Now read from resp.Body and stream it back using Chunked-Encoding
	rw.WriteStatusLine(200)
	hdrs := response.GetDefaultHeaders(0)
	hdrs.Remove("content-length")
	hdrs.Set("Transfer-encoding", "chunked")
	hdrs.Put("trailers", "X-Content-SHA256")
	hdrs.Put("trailers", "X-Content-Length")
	rw.WriteHeaders(hdrs)
	bufr := make([]byte, 1024)

	// For our trailers we can keep track incrementally, don't need to keep all of the data in memory
	hasher := sha256.New()
	contentLen := 0
	for {
		n, err := resp.Body.Read(bufr)
		contentLen += n
		hasher.Write(bufr[:n])
		if n == 0 || errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Fatalf("Hit err while reading response body, %s", err.Error())
		}
		bufr = bufr[:n]
		rw.WriteChunkedBody(bufr)
	}
	rw.WriteBody([]byte("0\r\n"))
	trailers := headers.NewHeaders()
	checksum := hex.EncodeToString(hasher.Sum(nil))
	trailers.Set("X-Content-SHA256", checksum)
	trailers.Set("X-Content-Length", fmt.Sprintf("%d", contentLen))
	rw.State = response.ResponseHeaders // need to manually set back state to avoid error state
	rw.WriteHeaders(trailers)
	rw.WriteBody([]byte("\r\n"))
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
