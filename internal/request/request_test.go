package request

import (
	"io"
	"testing"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n
	if n > cr.numBytesPerRead {
		n = cr.numBytesPerRead
		cr.pos -= n - cr.numBytesPerRead
	}
	return n, nil
}

// func TestRequestLineParse(t *testing.T) {
// 	// Test: Good GET Request line
// 	r, err := RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
// 	require.NoError(t, err)
// 	require.NotNil(t, r)
// 	assert.Equal(t, "GET", r.RequestLine.Method)
// 	assert.Equal(t, "/", r.RequestLine.RequestTarget)
// 	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

// 	// Test: Good GET Request line with path
// 	r, err = RequestFromReader(strings.NewReader("GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
// 	require.NoError(t, err)
// 	require.NotNil(t, r)
// 	assert.Equal(t, "GET", r.RequestLine.Method)
// 	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
// 	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

// 	// Test: Good POST Request line with path
// 	r, err = RequestFromReader(strings.NewReader("POST /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\nContent-Type: application/json\r\nContent-Length: 22\r\n\r\n{\"flavor\":\"dark mode\"}"))
// 	require.NoError(t, err)
// 	require.NotNil(t, r)
// 	assert.Equal(t, "POST", r.RequestLine.Method)
// 	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
// 	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

// 	// Test: Invalid number of parts in request line
// 	_, err = RequestFromReader(strings.NewReader("/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
// 	require.Error(t, err)

// 	// Test: Invalid method
// 	r, err = RequestFromReader(strings.NewReader("GETTO /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
// 	require.Error(t, err)

// 	// Test: Invalid method (out of order) Request line
// 	_, err = RequestFromReader(strings.NewReader("/coffee POST HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
// 	require.Error(t, err)

// }

// func TestStreamRequestLineParse(t *testing.T) {
// 	reader := &chunkReader{data: "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n", numBytesPerRead: 5}
// 	r, err := RequestFromReader(reader)
// 	require.NoError(t, err)
// 	require.NotNil(t, r)
// 	assert.Equal(t, "GET", r.RequestLine.Method)
// 	assert.Equal(t, "/", r.RequestLine.RequestTarget)
// 	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

// }

func TestRequestParsing(t *testing.T) {
	tests := []struct {
		name            string
		request         string
		numBytesPerRead int
		wantMethod      string
		wantTarget      string
		wantVersion     string
		wantErr         bool
	}{
		{
			name:            "Valid GET request",
			request:         "GET /home HTTP/1.1\r\n",
			numBytesPerRead: 4, // Simulate reading 4 bytes at a time
			wantMethod:      "GET",
			wantTarget:      "/home",
			wantVersion:     "1.1",
			wantErr:         false,
		},
		{
			name:            "Valid POST request",
			request:         "POST /submit HTTP/1.1\r\n",
			numBytesPerRead: 8, // Simulate reading 8 bytes at a time
			wantMethod:      "POST",
			wantTarget:      "/submit",
			wantVersion:     "1.1",
			wantErr:         false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Step 1: Create chunkReader for current test case
			reader := &chunkReader{data: tc.request, numBytesPerRead: tc.numBytesPerRead}

			// Step 2: Call RequestFromReader
			r, err := RequestFromReader(reader)
			requestline := r.RequestLine
			// Step 3: Validate errors
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("did not expect error but got: %v", err)
			}

			// Step 4: Validate parsed request fields
			if requestline.Method != tc.wantMethod {
				t.Errorf("expected method %s, got %s", tc.wantMethod, requestline.Method)
			}
			if requestline.RequestTarget != tc.wantTarget {
				t.Errorf("expected target %s, got %s", tc.wantTarget, requestline.RequestTarget)
			}
			if requestline.HttpVersion != tc.wantVersion {
				t.Errorf("expected version %s, got %s", tc.wantVersion, requestline.HttpVersion)
			}
		})
	}
}
