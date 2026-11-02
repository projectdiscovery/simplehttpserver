package httpserver

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type countingBody struct {
	r    io.Reader
	read int64
}

func (c *countingBody) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.read += int64(n)
	return n, err
}

func (c *countingBody) Close() error { return nil }

// TestLogLayerBoundsChunkedBody checks that a body without Content-Length
// (chunked) is not read past the dump limit before the next handler runs,
// and that the next handler still gets the whole body.
func TestLogLayerBoundsChunkedBody(t *testing.T) {
	EnableVerbose = true
	defer func() { EnableVerbose = false }()

	const limit = 1024
	payload := bytes.Repeat([]byte("A"), 1<<20)
	body := &countingBody{r: bytes.NewReader(payload)}

	server := &HTTPServer{options: &Options{MaxDumpBodySize: limit}}
	var readBeforeHandler int64
	var got []byte
	handler := server.loglayer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		readBeforeHandler = body.read
		got, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusUnauthorized)
	}))

	req := httptest.NewRequest(http.MethodPost, "http://example.com/", nil)
	req.Body = body
	req.ContentLength = -1
	req.TransferEncoding = []string{"chunked"}
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if readBeforeHandler > limit {
		t.Fatalf("loglayer read %d bytes before the next handler, limit %d", readBeforeHandler, limit)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("next handler got %d bytes, want %d", len(got), len(payload))
	}
}

// TestLogLayerSkipsBodyWithoutLog checks that the body is not read when no
// log needs it.
func TestLogLayerSkipsBodyWithoutLog(t *testing.T) {
	body := &countingBody{r: bytes.NewReader(bytes.Repeat([]byte("A"), 4096))}
	server := &HTTPServer{options: &Options{}}
	var readBeforeHandler int64 = -1
	handler := server.loglayer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		readBeforeHandler = body.read
	}))
	req := httptest.NewRequest(http.MethodPost, "http://example.com/", nil)
	req.Body = body
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if readBeforeHandler != 0 {
		t.Fatalf("loglayer read %d bytes with no log enabled", readBeforeHandler)
	}
}
