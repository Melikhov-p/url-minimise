package worker

import (
	"testing"
	"time"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/logger"
	"github.com/Melikhov-p/url-minimise/internal/repository"
	"github.com/stretchr/testify/assert"
)

func BenchmarkWorker(b *testing.B) {
	log, err := logger.BuildLogger("ERROR")
	if err != nil {
		panic(err.Error())
	}
	store, err := repository.NewStorage(config.NewConfig(log, true), log)
	if err != nil {
		panic(err.Error())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		worker := NewDelWorker(5*time.Second, log, store)
		go worker.LookUp()
	}
}

func TestNewDelWorker(t *testing.T) {
	pingInterval := 15 * time.Second
	log, err := logger.BuildLogger("ERROR")
	if err != nil {
		panic(err.Error())
	}
	store, err := repository.NewStorage(config.NewConfig(log, true), log)
	if err != nil {
		panic(err.Error())
	}
	dw := NewDelWorker(pingInterval, log, store)

	assert.Equal(t, dw.PingInterval, pingInterval)
}
