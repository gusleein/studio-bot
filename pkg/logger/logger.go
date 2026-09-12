package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger — обёртка над *zap.Logger с удобными методами.
type Logger struct {
	zap *zap.Logger
}

// New создаёт production-логгер с JSON-форматом вывода.
func New() (*Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	z, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return &Logger{zap: z}, nil
}

// NewDevelopment создаёт development-логгер с human-friendly форматом.
func NewDevelopment() (*Logger, error) {
	z, err := zap.NewDevelopment(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return &Logger{zap: z}, nil
}

// Sync сбрасывает буферизованные записи лога.
func (l *Logger) Sync() error {
	return l.zap.Sync()
}

// Zap возвращает исходный *zap.Logger для использования в сторонних библиотеках.
func (l *Logger) Zap() *zap.Logger {
	return l.zap
}

// With возвращает новый Logger с дополнительными полями.
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{zap: l.zap.With(fields...)}
}

// Info логирует сообщение уровня INFO.
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

// Error логирует сообщение уровня ERROR.
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

// Debug логирует сообщение уровня DEBUG.
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

// Fatal логирует сообщение уровня FATAL и завершает процесс.
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}

// Warn логирует сообщение уровня WARN.
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

// Infof логирует форматированное сообщение уровня INFO.
func (l *Logger) Infof(template string, args ...any) {
	l.zap.Sugar().Infof(template, args...)
}

// Errorf логирует форматированное сообщение уровня ERROR.
func (l *Logger) Errorf(template string, args ...any) {
	l.zap.Sugar().Errorf(template, args...)
}

// Debugf логирует форматированное сообщение уровня DEBUG.
func (l *Logger) Debugf(template string, args ...any) {
	l.zap.Sugar().Debugf(template, args...)
}
