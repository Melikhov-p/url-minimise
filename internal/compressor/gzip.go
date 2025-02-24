// Пакет compress предоставляет утилиты для сжатия и декомпрессии HTTP-ответов с использованием gzip.
package compress

import (
	"compress/gzip"
	"io"
	"net/http"
)

const statusCodeThreshHold = 300

// CompressWriter — обёртка вокруг http.ResponseWriter, которая сжимает ответ с использованием gzip.
type CompressWriter struct {
	w  http.ResponseWriter
	gw *gzip.Writer
}

// NewCompressWriter создаёт новый экземпляр CompressWriter.
// Инициализирует gzip-писатель с помощью переданного http.ResponseWriter.
func NewCompressWriter(w http.ResponseWriter) *CompressWriter {
	return &CompressWriter{w: w, gw: gzip.NewWriter(w)}
}

// Header возвращает заголовки базового http.ResponseWriter.
func (c *CompressWriter) Header() http.Header {
	return c.w.Header()
}

// Write записывает сжатые данные в ответ.
func (c *CompressWriter) Write(p []byte) (int, error) {
	return c.gw.Write(p)
}

// WriteHeader устанавливает HTTP-код статуса и добавляет заголовок "Content-Encoding: gzip",
// если код статуса ниже порогового значения.
func (c *CompressWriter) WriteHeader(statusCode int) {
	if statusCode < statusCodeThreshHold {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip-писатель, сбрасывая оставшиеся данные.
func (c *CompressWriter) Close() error {
	return c.gw.Close()
}

// CompressReader — обёртка вокруг io.ReadCloser, которая декомпрессирует ответ с использованием gzip.
type CompressReader struct {
	r  io.ReadCloser
	gr *gzip.Reader
}

// NewCompressReader создаёт новый экземпляр CompressReader.
// Инициализирует gzip-читатель с помощью переданного io.ReadCloser.
func NewCompressReader(r io.ReadCloser) (*CompressReader, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &CompressReader{r: r, gr: gr}, nil
}

// Read считывает декомпрессированные данные из ответа.
func (c *CompressReader) Read(p []byte) (int, error) {
	return c.gr.Read(p)
}

// Close закрывает базовый io.ReadCloser и gzip-читатель.
func (c *CompressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}

	return c.gr.Close()
}
