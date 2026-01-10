package request

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
// Use ptr to chunkReader struct because the Read call is modifying it and to avoid copying everytime we call
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := min(cr.pos+cr.numBytesPerRead, len(cr.data))

	// copy chunk reader data over some range into p
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func checkRequestLineCorrect(method string, target string, r *Request, err error, t *testing.T) {
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, method, r.RequestLine.Method)
	assert.Equal(t, target, r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

// Example test structure
func TestChunkReader(t *testing.T) {
	reader := &chunkReader{
		data:            "Hello, World! This is a test.",
		numBytesPerRead: 5,
		pos:             0,
	}

	readData, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fmt.Println("testing reader readAll():", string(readData))

	expected := "Hello, World! This is a test."
	if string(readData) != expected {
		t.Errorf("got %q, want %q", string(readData), expected)
	}
}

func TestRequestLineParse(t *testing.T) {
	// Test: Good GET Request line
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	checkRequestLineCorrect("GET", "/", r, err, t)

	// Test: Good GET Request line with path
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}
	r, err = RequestFromReader(reader)
	checkRequestLineCorrect("GET", "/coffee", r, err, t)

	// Test: Good GET with reader reading full string all at once
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 0, // Will be set below
	}
	reader.numBytesPerRead = len(reader.data)
	r, err = RequestFromReader(reader)
	checkRequestLineCorrect("GET", "/coffee", r, err, t)

	// Test: Invalid number of parts in request line
	reader = &chunkReader{
		data:            "/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: bad version
	reader = &chunkReader{
		data:            "GET /coffee HTTP-1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)
}

func TestRequestsWithHeaders(t *testing.T) {
	// Test: good request with normal headers
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\n			User-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	assert.Equal(t, r.headers.Get("HOST"), "localhost:42069")
	assert.Equal(t, r.headers.Get("user-agent"), "curl/7.81.0")
	assert.Equal(t, r.headers.Get("ACCEPT"), "*/*")

	// Test: headers with invalid characters
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\n@ccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: invalid whitespace in header
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept	: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: valid duplicate field name headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	assert.Equal(t, r.headers.Get("accept"), "*/*, */*")

}
