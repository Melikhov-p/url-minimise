package worker

import (
	"context"
	"time"

	"github.com/Melikhov-p/url-minimise/internal/models"
	"github.com/Melikhov-p/url-minimise/internal/repository"
	taskService "github.com/Melikhov-p/url-minimise/internal/service"
	"go.uber.org/zap"
)

// Worker интерфейс воркера
type Worker interface {
	LookUp()
}

// DelWorker воркер, который будет следить за тасками на удаление
type DelWorker struct {
	PingInterval time.Duration
	Logger       *zap.Logger
	Storage      repository.Storage
	stop         bool
}

// NewDelWorker возвращает воркера, который будет следить за тасками на удаление
func NewDelWorker(pingInterval time.Duration, logger *zap.Logger, storage repository.Storage) *DelWorker {
	return &DelWorker{
		PingInterval: pingInterval,
		Logger:       logger,
		Storage:      storage,
		stop:         false,
	}
}

// LookUp основной луп воркера
func (dw *DelWorker) LookUp() {
	dw.Logger.Info("worker: starting look up for delete tasks")
	for {
		if dw.stop {
			dw.Logger.Debug("del worker stopped")
			return
		}

		dw.Logger.Debug("worker: ping tasks")
		ctx := context.Background()
		tasks, err := taskService.GetDeleteTasksWStatus(ctx, models.Registered, dw.Storage)
		if err != nil {
			dw.Logger.Error("worker: error getting tasks for delete", zap.Error(err))
			time.Sleep(dw.PingInterval)
			continue
		}
		if len(tasks) == 0 {
			time.Sleep(dw.PingInterval)
			continue
		}
		dw.Logger.Debug("worker: found del tasks")

		err = taskService.MarkAsDeleted(ctx, tasks, dw.Storage)
		if err != nil {
			dw.Logger.Error("worker: error updating records in storage", zap.Error(err))
			time.Sleep(dw.PingInterval)
			continue
		}
		dw.Logger.Debug("worker: mark URLs from task deleted")

		err = taskService.UpdateTasksStatus(ctx, tasks, models.Done, dw.Storage)
		if err != nil {
			dw.Logger.Error("worker: error updating tasks statuses", zap.Error(err))
			time.Sleep(dw.PingInterval)
			continue
		}
		dw.Logger.Debug("worker: update done task statuses")

		time.Sleep(dw.PingInterval)
	}
}

func (dw *DelWorker) Stop() {
	dw.Logger.Debug("del worker got signal for stopping")
	dw.stop = true
}
