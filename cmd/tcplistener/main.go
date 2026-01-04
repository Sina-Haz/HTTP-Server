package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"sina.http/internal/request"
)

const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("error", err)
	}
	defer listener.Close()
	fmt.Println("Listening for TCP traffic on port", port)
	// wait for listenek to connect
	conn, err := listener.Accept()
	if err != nil {
		log.Fatal("error", err)
	}
	r, err := request.RequestFromReader(conn)
	fmt.Println("Request line:")
	fmt.Printf("- Method: %s\n", r.RequestLine.Method)
	fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
	fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)
}

// So input f implements Read() and Close() that's all io.ReadCloser
func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)
	go func() {
		defer f.Close()
		defer close(lines)
		currentLineContents := ""
		for {
			b := make([]byte, 8)
			n, err := f.Read(b)
			if err != nil {
				if currentLineContents != "" {
					lines <- currentLineContents
				}
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("error: %s\n", err.Error())
				return
			}
			str := string(b[:n])
			parts := strings.Split(str, "\n")
			for i := 0; i < len(parts)-1; i++ {
				lines <- fmt.Sprintf("%s%s", currentLineContents, parts[i])
				currentLineContents = ""
			}
			currentLineContents += parts[len(parts)-1]
		}
	}()
	return lines
}
