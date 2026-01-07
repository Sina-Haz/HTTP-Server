package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sina.http/internal/request"
)

func TestHelloWorld(t *testing.T) {
	// t.Fatal("not implemented")
}

func TestHeaderParsing(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid header + extra whitespace
	headers = NewHeaders()
	data = []byte("       Host: localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, len(data)-len(request.CRLF), n)
	assert.False(t, done)

	// Test: Valid done + existing headers
	data = []byte("\r\n{message body here}")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Test: valid multiple headers + existing headers
	headers = NewHeaders()
	data = []byte("       Host: localhost:42069       \r\nanother_header: fooooo    \r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.False(t, done)

	// Test: invalid character in the field name
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)

	// Test: field-name is case-insensitive
	assert.Equal(t, headers.Get("HOST"), headers.Get("Host"))

	// Test: properly handles multiple of the same field-name by concatenating them together
	data = []byte("       Host: google.com       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069, google.com", headers.Get("Host"))
	assert.False(t, done)

}
