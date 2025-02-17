package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompressWriter(t *testing.T) {
	// Создаем тестовый HTTP-ответ
	rec := httptest.NewRecorder()
	cw := NewCompressWriter(rec)

	// Устанавливаем заголовок и пишем данные
	cw.Header().Set("Content-Type", "text/plain")
	data := []byte("Hello, Gzip!")
	_, err := cw.Write(data)
	assert.NoError(t, err)

	// Устанавливаем код статуса и закрываем писатель
	cw.WriteHeader(http.StatusOK)
	assert.NoError(t, cw.Close())

	// Проверяем, что заголовок Content-Encoding установлен
	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))

	// Проверяем, что данные сжаты корректно
	gr, err := gzip.NewReader(rec.Body)
	assert.NoError(t, err)
	decompressedData, err := io.ReadAll(gr)
	assert.NoError(t, err)
	assert.Equal(t, data, decompressedData)
}

func TestCompressReader(t *testing.T) {
	// Создаем сжатые данные
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	data := []byte("Hello, Gzip!")
	_, err := gw.Write(data)
	assert.NoError(t, err)
	assert.NoError(t, gw.Close())

	// Создаем CompressReader
	cr, err := NewCompressReader(io.NopCloser(&buf))
	assert.NoError(t, err)

	// Читаем и декомпрессируем данные
	decompressedData, err := io.ReadAll(cr)
	assert.NoError(t, err)
	assert.Equal(t, data, decompressedData)

	// Закрываем читатель
	assert.NoError(t, cr.Close())
}
