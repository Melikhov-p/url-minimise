package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestBuildLogger(t *testing.T) {
	tests := []struct {
		level    string
		expected zapcore.Level
		wantErr  bool
	}{
		{"debug", zapcore.DebugLevel, false},
		{"info", zapcore.InfoLevel, false},
		{"warn", zapcore.WarnLevel, false},
		{"error", zapcore.ErrorLevel, false},
		{"fatal", zapcore.FatalLevel, false},
		{"panic", zapcore.PanicLevel, false},
		{"invalid", zapcore.Level(0), true}, // неверный уровень
	}

	for _, tt := range tests {
		logger, err := BuildLogger(tt.level)

		if tt.wantErr {
			assert.Error(t, err, "Expected error for level %s", tt.level)
		} else {
			assert.NoError(t, err, "Unexpected error for level %s", tt.level)
			assert.Equal(t, tt.expected, logger.Level(), "Log level mismatch for level %s", tt.level)
		}
	}
}
