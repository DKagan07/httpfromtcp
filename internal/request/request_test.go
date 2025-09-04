package request

import (
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
// its useful for simulating reading a variable number of bytes per chunk from a
// network connection
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

	return n, nil
}

func TestRequestLineParse(t *testing.T) {
	t.Run("Good GET request line", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	})

	t.Run("Good GET Request Line with path", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 1,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	})

	t.Run("Good POST Request Line with path", func(t *testing.T) {
		reader := &chunkReader{
			data: "POST /coffee/order/1 HTTP/1.1\r\nHost: localhost:42068\r\n\r\n'{\"coffee_type\":\"agentic\"'}",
			numBytesPerRead: len(
				[]byte(
					"POST /coffee/order/1 HTTP/1.1\r\nHost: localhost:42068\r\n\r\n'{\"coffee_type\":\"agentic\"'}",
				),
			),
		}

		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "POST", r.RequestLine.Method)
		assert.Equal(t, "/coffee/order/1", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	})

	t.Run("Invalid number of parts in request line", func(t *testing.T) {
		reader := &chunkReader{
			data:            "/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 10,
		}

		_, err := RequestFromReader(reader)
		require.Error(t, err)
	})

	t.Run("Invalid method (out of order) in Request Line", func(t *testing.T) {
		reader := &chunkReader{
			data:            "/coffee GET HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\n\r\n",
			numBytesPerRead: 50,
		}

		_, err := RequestFromReader(reader)
		assert.Error(t, err)
	})

	t.Run("Invalid HTTP version number", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET /coffee HTTP/1.9\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\n\r\n",
			numBytesPerRead: 15,
		}

		_, err := RequestFromReader(reader)
		assert.Error(t, err)
	})

	t.Run("Handle standard headers", func(t *testing.T) {
		// Test: Standard Headers
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "localhost:42069", r.Headers["host"])
		assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
		assert.Equal(t, "*/*", r.Headers["accept"])
	})

	t.Run("Malformed headers", func(t *testing.T) {
		// Test: Malformed Header
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
			numBytesPerRead: 3,
		}

		_, err := RequestFromReader(reader)
		require.Error(t, err)
	})
}
