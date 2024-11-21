package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// BuildLogger строит zap.Logger с необходимым уровнем логирования.
func BuildLogger(level string) (*zap.Logger, error) {
	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("failed to parse atomic level: %w", err)
	}
	// создаём новую конфигурацию логера
	cfg := zap.NewProductionConfig()
	// устанавливаем уровень
	cfg.Level = lvl
	// создаём логер на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build config %v", err)
	}
	return zl, nil
}
