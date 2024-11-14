package compressor

import (
	"compress/gzip"
	"io"
	"net/http"
)

type GzipCompressWriter struct {
	http.ResponseWriter
	gw *gzip.Writer
}

func NewGzipCompressWriter(w http.ResponseWriter) *GzipCompressWriter {
	return &GzipCompressWriter{
		ResponseWriter: w,
		gw:             gzip.NewWriter(w),
	}
}

func (c *GzipCompressWriter) Close() error {
	return c.gw.Close()
}

type GzipCompressReader struct {
	io.ReadCloser
	zr *gzip.Reader
}

func NewGzipCompressReader(r io.ReadCloser) (*GzipCompressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &GzipCompressReader{
		ReadCloser: r,
		zr:         zr,
	}, nil
}
