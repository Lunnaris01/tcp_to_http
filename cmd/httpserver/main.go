package main

import (
	"crypto/sha256"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

const port = 42069

func main() {
	server, err := server.Serve(handler, port)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	log.Print("handling request")
	if req.RequestLine.RequestTarget == "/yourproblem" {
		handler400(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		handler500(w, req)
		return
	}
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		log.Print("/httbin route")
		handlerChunkBody(w, req)
		return
	}
	handler200(w, req)
	return
}

func handlerChunkBody(w *response.Writer, req *request.Request) {
	proxyAddr := "https://httpbin.org"
	proxyPath := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	proxyRoute := proxyAddr + proxyPath
	log.Printf("Trying to forward data from %s", proxyRoute)
	resp, err := http.Get(proxyRoute)
	log.Printf("Got response successfully %v, %v", resp, err)
	if err != nil {
		log.Printf("Failed to read from httbin.org")
		handler500(w, req)
		return
	}
	log.Print("Writing Response Line (Chunked)")
	err = w.WriteStatusLine(response.StatusOk)
	if err != nil {
		log.Printf("Error Writing Status Line: %v", err)
	}
	log.Print("Writing Header (Chunked)")
	myHeaders := response.GetDefaultChunkedHeaders()
	myHeaders.AddKey("Trailer", "X-Content-SHA256")
	myHeaders.AddKey("Trailer", "X-Content-Length")
	err = w.WriteHeaders(myHeaders)
	if err != nil {
		log.Printf("Error Writing Headers: %v", err)
	}

	readbytes := 0
	buffer := make([]byte, 32) // 32 bytes at a time
	hashBuffer := make([]byte, 1024)
	for {
		if len(hashBuffer) <= (readbytes + 32) {
			newBuffer := make([]byte, len(hashBuffer)*2)
			copy(newBuffer, hashBuffer)
			hashBuffer = newBuffer
		}
		n, err := resp.Body.Read(buffer)
		copy(hashBuffer[readbytes:readbytes+n], buffer[:n])
		readbytes += n

		// If we read any bytes, write them as a chunk
		if n > 0 {
			// Only write n bytes from the buffer - not the entire buffer!
			log.Printf("Writing Chunk of length %d", n)
			_, writeErr := w.WriteChunkedBody(buffer[:n])
			if writeErr != nil {
				log.Printf("Error writing chunk: %v", writeErr)
				break
			}
		}

		// Handle end of stream or errors
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from response: %v", err)
			}
			break
		}
	}
	w.WriteChunkedBodyDone()

	trailers := headers.NewHeaders()
	hashedBody := sha256.Sum256(hashBuffer[:readbytes])

	trailers.AddKey("X-Content-SHA256", fmt.Sprintf("%x", hashedBody))
	log.Print(readbytes)
	trailers.AddKey("X-Content-Length", strconv.Itoa(readbytes))

	w.WriteTrailers(trailers)
	return

}

func handler400(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusBadRequest)
	body := []byte(`<html>
<head>
<title>400 Bad Request</title>
</head>
<body>
<h1>Bad Request</h1>
<p>Your request honestly kinda sucked.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.SetKey("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
	return
}

func handler500(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusInternalServerError)
	body := []byte(`<html>
<head>
<title>500 Internal Server Error</title>
</head>
<body>
<h1>Internal Server Error</h1>
<p>Okay, you know what? This one is on me.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.SetKey("Content-Type", "text/html")
	for key, val := range h {
		log.Printf("%s: %s", key, val)
	}
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler200(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusOk)
	body := []byte(`<html>
<head>
<title>200 OK</title>
</head>
<body>
<h1>Success!</h1>
<p>Your request was an absolute banger.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.SetKey("content-type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
	return
}
