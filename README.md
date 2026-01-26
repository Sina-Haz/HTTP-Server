# Custom HTTP/1.1 Server in Go

A complete HTTP/1.1 server implementation written in Go from scratch. This project demonstrates understanding of HTTP protocol fundamentals while providing a functional web server with advanced features like chunked transfer encoding and proxy functionality.

## 🚀 Features

- **Complete HTTP/1.1 Protocol**: Full implementation of HTTP/1.1 with proper request/response handling
- **Custom TCP Listener**: Built-in TCP server with connection management
- **Proxy Functionality**: Forward requests to external services (httpbin.org)
- **File Serving**: Serve static files (videos, HTML, etc.)
- **Chunked Transfer Encoding**: Support for streaming responses
- **Graceful Shutdown**: Proper server shutdown handling
- **UDP Testing Client**: Included UDP sender for development/testing
- **Standalone TCP Listener**: Dedicated tool for testing request parsing

## 📋 Requirements

- Go 1.19 or higher
- No external dependencies (standard library only)

## 🛠️ Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd http_server
```

2. Build the applications:
```bash
# Build the main HTTP server
go build -o httpserver cmd/httpserver/main.go

# Build the TCP listener (for testing)
go build -o tcplistener cmd/tcplistener/main.go

# Build the UDP sender (for testing)
go build -o udpsender cmd/udpsender/main.go
```

## 🎮 Usage

### Main HTTP Server

```bash
./httpserver
```

The server starts on port `42069` with the following endpoints:

- `/yourproblem` - Returns 400 Bad Request
- `/myproblem` - Returns 500 Internal Server Error  
- `/video` - Serves `assets/vim.mp4`
- `/httpbin/*` - Proxies requests to httpbin.org with chunked encoding

### TCP Listener (Testing Tool)

```bash
./tcplistener
```

Standalone TCP listener for testing HTTP request parsing without running the full server.

### UDP Sender (Testing Tool)

```bash
./udpsender
```

Simple UDP client to send test messages to the HTTP server on localhost:42069.

## 📁 Project Structure

```
├── cmd/
│   ├── httpserver/     # Main HTTP server application
│   ├── tcplistener/    # TCP testing tool
│   └── udpsender/      # UDP testing tool
├── internal/
│   ├── server/         # Server management and connections
│   ├── request/        # HTTP request parsing
│   ├── response/       # HTTP response handling
│   └── headers/        # HTTP header management
└── assets/             # Static files (videos, etc.)
```

## ⚠️ Limitations

### Security Concerns
- **Path Traversal Risk**: Static file serving uses relative paths without validation
- **No Input Validation**: Request targets are not sanitized
- **SSRF Vulnerability**: Proxy functionality could be abused for server-side request forgery
- **No HTTPS Support**: Only HTTP protocol implemented

### Performance Issues
- **Memory Inefficiency**: Entire request bodies loaded into memory instead of streaming
- **Fixed Buffer Size**: Uses static 1024-byte buffers without dynamic resizing
- **No Connection Pooling**: Unlimited connections without proper limits

### Code Quality
- **Error Handling**: Mix of `log.Fatalf` and error returning (can crash server)
- **Complex State Management**: Manual state manipulation prone to errors
- **Limited Testing**: Minimal test coverage for error scenarios
- **TODO Items**: Several incomplete features and potential bugs

### Missing Features
- No authentication or authorization
- No rate limiting
- No structured logging
- No metrics or monitoring
- No configuration management
- No middleware support

## 🐛 TODO / Future Improvements

### High Priority
- [ ] Implement proper error handling (replace `log.Fatalf`)
- [ ] Add input validation and sanitization
- [ ] Implement HTTPS support with TLS
- [ ] Add path validation for file serving

### Medium Priority  
- [ ] Implement streaming request/response handling
- [ ] Add connection pooling and limits
- [ ] Implement structured logging
- [ ] Add comprehensive test coverage

### Low Priority
- [ ] Add configuration management (environment variables)
- [ ] Implement middleware pattern
- [ ] Add rate limiting
- [ ] Create proper documentation with examples
- [ ] Add benchmark tests

## 🧪 Testing

Run tests using:
```bash
go test ./...
```

## 📝 License

This project is for educational purposes. Not recommended for production use due to security and performance limitations.

## 🤝 Contributing

Feel free to submit issues and enhancement requests. Please ensure:
- Code follows Go best practices
- All TODO items are addressed before production use
- Security considerations are thoroughly reviewed

---

**Note**: This is a learning project demonstrating HTTP protocol implementation. For production use, consider using established HTTP servers like Go's standard `net/http` package or other battle-tested solutions.