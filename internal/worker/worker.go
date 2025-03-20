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
	PingPoint    time.Time
	PingInterval time.Duration
	Logger       *zap.Logger
	Storage      repository.Storage
	stop         chan bool
}

// NewDelWorker возвращает воркера, который будет следить за тасками на удаление
func NewDelWorker(pingInterval time.Duration, logger *zap.Logger, storage repository.Storage) *DelWorker {
	return &DelWorker{
		PingPoint:    time.Now(),
		PingInterval: pingInterval,
		Logger:       logger,
		Storage:      storage,
		stop:         make(chan bool, 1),
	}
}

// LookUp основной луп воркера
func (dw *DelWorker) LookUp() {
	dw.Logger.Info("worker: starting look up for delete tasks")

loop:
	for {
		select {
		case <-dw.stop:
			break loop
		default:
			if time.Now().After(dw.PingPoint) {
				dw.Logger.Debug("worker: ping tasks")

				ctx := context.Background()
				tasks, err := taskService.GetDeleteTasksWStatus(ctx, models.Registered, dw.Storage)
				if err != nil {
					dw.Logger.Error("worker: error getting tasks for delete", zap.Error(err))
					dw.pingAfterInterval()
					continue
				}
				if len(tasks) == 0 {
					dw.pingAfterInterval()
					continue
				}
				dw.Logger.Debug("worker: found del tasks")

				err = taskService.MarkAsDeleted(ctx, tasks, dw.Storage)
				if err != nil {
					dw.Logger.Error("worker: error updating records in storage", zap.Error(err))
					dw.pingAfterInterval()
					continue
				}
				dw.Logger.Debug("worker: mark URLs from task deleted")

				err = taskService.UpdateTasksStatus(ctx, tasks, models.Done, dw.Storage)
				if err != nil {
					dw.Logger.Error("worker: error updating tasks statuses", zap.Error(err))
					dw.pingAfterInterval()
					continue
				}
				dw.Logger.Debug("worker: update done task statuses")

				dw.pingAfterInterval()
			}
		}
	}

	dw.Logger.Debug("del worker stopped")
}

// pingAfterInterval ping tasks after interval.
func (dw *DelWorker) pingAfterInterval() {
	dw.PingPoint = time.Now().Add(dw.PingInterval)
}

// Stop worker.
func (dw *DelWorker) Stop() {
	defer func() {
		close(dw.stop)
	}()

	dw.Logger.Debug("del worker got signal for stopping")
	dw.stop <- true
}
