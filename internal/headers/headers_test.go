package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaders(t *testing.T) {
	t.Run("Valid single header", func(t *testing.T) {
		h := NewHeaders()

		n, done, err := h.Parse([]byte("Host: localhost:42069\r\n\r\n"))
		require.NoError(t, err)
		require.NotNil(t, h)
		assert.Equal(t, "localhost:42069", h["host"])
		assert.Equal(t, 23, n)
		assert.False(t, done)
	})

	t.Run("Invalid spacing header", func(t *testing.T) {
		h := NewHeaders()

		n, done, err := h.Parse([]byte("      Host : localhost:42069           \r\n\r\n"))
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Valid headers with other present headers", func(t *testing.T) {
		h := NewHeaders()
		h["host"] = "localhost:42069"

		n, done, err := h.Parse([]byte("Content-Type: application/json\r\n\r\n"))
		require.NoError(t, err)
		assert.Equal(t, "application/json", h["content-type"])
		assert.Equal(t, 32, n)
		assert.False(t, done)
	})

	t.Run("Data starts with a CRLF", func(t *testing.T) {
		h := NewHeaders()

		n, done, err := h.Parse([]byte("\r\n"))
		require.NoError(t, err)
		assert.Equal(t, 2, n)
		assert.True(t, done)
	})

	t.Run("Invalid header character", func(t *testing.T) {
		h := NewHeaders()

		n, done, err := h.Parse([]byte("H©st: localhost:42069\r\n\r\n"))
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Multiple header keys", func(t *testing.T) {
		h := NewHeaders()
		h["set-person"] = "lane-loves-go"

		n, done, err := h.Parse([]byte("Set-Person: prime-loves-zig\r\n\r\n"))
		require.NoError(t, err)
		assert.Equal(t, "lane-loves-go, prime-loves-zig", h["set-person"])
		assert.Equal(t, 29, n)
		assert.False(t, done)
	})
}
