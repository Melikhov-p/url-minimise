package compress

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
)

const statusCodeThreshHold = 300

type CompressWriter struct {
	w  http.ResponseWriter
	gw *gzip.Writer
}

func NewCompressWrite(w http.ResponseWriter) *CompressWriter {
	return &CompressWriter{w: w, gw: gzip.NewWriter(w)}
}

func (c *CompressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *CompressWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	if err != nil {
		return n, fmt.Errorf("failed to write to gzip writer: %w", err)
	}
	return n, nil
}

func (c *CompressWriter) WriteHeader(statusCode int) {
	if statusCode < statusCodeThreshHold {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *CompressWriter) Close() error {
	err := c.gw.Close()
	if err != nil {
		return fmt.Errorf("error closing writer %w", err)
	}
	return nil
}

type CompressReader struct {
	r  io.ReadCloser
	gr *gzip.Reader
}

func NewCompressReader(r io.ReadCloser) (*CompressReader, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("error creating new reader %w", err)
	}

	return &CompressReader{r: r, gr: gr}, nil
}

func (c *CompressReader) Read(p []byte) (int, error) {
	n, err := c.gr.Read(p)
	if err != nil {
		return n, fmt.Errorf("failed to read from gzip: %w", err)
	}
	return n, nil
}

func (c *CompressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return fmt.Errorf("error closing reader %w", err)
	}

	return nil
}
