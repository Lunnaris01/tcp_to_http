package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestHeaderLineParse(t *testing.T) {

	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Valid double header Full Test.
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\n")
	data_2 := []byte("Content: media/json\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)
	n, done, err = headers.Parse(data_2)
	assert.Equal(t, "media/json", headers["content"])
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 21, n)
	assert.False(t, done)
	n, done, err = headers.Parse([]byte("\r\n"))
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Test: Valid double header same key.
	headers = NewHeaders()
	data = []byte("Username: Lunnaris\r\n")
	data_2 = []byte("Username: Julian\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "Lunnaris", headers["username"])
	assert.Equal(t, 20, n)
	assert.False(t, done)
	n, done, err = headers.Parse(data_2)
	assert.Equal(t, "Lunnaris, Julian", headers["username"])
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 18, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Partial header
	headers = NewHeaders()
	data = []byte("Host: localhost:")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Illegal rune in key
	headers = NewHeaders()
	data = []byte("Höst: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

}
