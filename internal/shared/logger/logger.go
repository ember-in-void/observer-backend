package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// ANSI цвета для консоли
const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
	// colorYellow = "\033[33m"
	// colorBlue   = "\033[34m"
	// colorPurple = "\033[35m"
	colorCyan = "\033[36m"
	colorGray = "\033[90m"
	// colorWhite  = "\033[97m"

	// Bold variants
	colorBoldRed   = "\033[1;31m"
	colorBoldGreen = "\033[1;32m"
	// colorBoldYellow = "\033[1;33m"
	colorBoldCyan = "\033[1;36m"
)

// Logger - интерфейс для логгера, позволяет легко мокать в тестах
type Logger interface {
	Info(msg ...any)
	Infof(format string, args ...any)
	Warn(msg ...any)
	Warnf(format string, args ...any)
	Error(msg ...any)
	Errorf(format string, args ...any)
	WithField(key string, value any) Logger
	WithFields(fields map[string]any) Logger
}

// Format тип формата логов
type Format string

const (
	FormatPretty Format = "pretty" // Красивый цветной формат (по умолчанию)
	FormatJSON   Format = "json"   // JSON формат
	FormatPlain  Format = "plain"  // Простой текст без цветов
)

// CustomLogger - основная реализация логгера
type CustomLogger struct {
	consoleWriter io.Writer
	fileWriter    io.Writer
	fields        map[string]any
	format        Format
	useColors     bool
	mu            sync.RWMutex
}

// Config - конфигурация логгера
type Config struct {
	LogDir      string // Директория для логов (по умолчанию "logs")
	WriteToFile bool   // Писать ли в файлы (по умолчанию true)
	Format      Format // Формат вывода (pretty, json, plain)
	UseColors   bool   // Использовать цвета в консоли
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() Config {
	return Config{
		LogDir:      "logs",
		WriteToFile: false,
		Format:      FormatPretty,
		UseColors:   true,
	}
}

// NewCustomLogger создаёт новый логгер с конфигурацией по умолчанию
func NewCustomLogger() (*CustomLogger, error) {
	return NewCustomLoggerWithConfig(DefaultConfig())
}

// NewCustomLoggerWithConfig создаёт новый логгер с указанной конфигурацией
func NewCustomLoggerWithConfig(cfg Config) (*CustomLogger, error) {
	var fileWriter io.Writer

	if cfg.WriteToFile {
		err := os.MkdirAll(cfg.LogDir, 0o755)
		if err != nil {
			return nil, fmt.Errorf("failed to create logs directory: %w", err)
		}

		// Один файл для всех логов - проще искать
		logFile, err := os.OpenFile(
			filepath.Join(cfg.LogDir, "app.log"),
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0o666,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		fileWriter = logFile
	}

	logger := &CustomLogger{
		consoleWriter: os.Stdout,
		fileWriter:    fileWriter,
		fields:        make(map[string]any),
		format:        cfg.Format,
		useColors:     cfg.UseColors,
	}

	return logger, nil
}

// Info логирует сообщение с уровнем INFO
func (l *CustomLogger) Info(msg ...any) {
	l.log("INFO", fmt.Sprint(msg...))
}

// Infof логирует форматированное сообщение с уровнем INFO
func (l *CustomLogger) Infof(format string, args ...any) {
	l.log("INFO", fmt.Sprintf(format, args...))
}

// Warn логирует сообщение с уровнем WARN
func (l *CustomLogger) Warn(msg ...any) {
	l.log("WARN", fmt.Sprint(msg...))
}

// Warnf логирует форматированное сообщение с уровнем WARN
func (l *CustomLogger) Warnf(format string, args ...any) {
	l.log("WARN", fmt.Sprintf(format, args...))
}

// Error логирует сообщение с уровнем ERROR
func (l *CustomLogger) Error(msg ...any) {
	l.log("ERROR", fmt.Sprint(msg...))
}

// Errorf логирует форматированное сообщение с уровнем ERROR
func (l *CustomLogger) Errorf(format string, args ...any) {
	l.log("ERROR", fmt.Sprintf(format, args...))
}

// WithField возвращает новый логгер с добавленным полем
func (l *CustomLogger) WithField(key string, value any) Logger {
	newLogger := l.clone()
	newLogger.fields[key] = value
	return newLogger
}

// WithFields возвращает новый логгер с добавленными полями
func (l *CustomLogger) WithFields(fields map[string]any) Logger {
	newLogger := l.clone()
	for k, v := range fields {
		newLogger.fields[k] = v
	}
	return newLogger
}

// clone создаёт копию логгера
func (l *CustomLogger) clone() *CustomLogger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newFields := make(map[string]any, len(l.fields))
	for k, v := range l.fields {
		newFields[k] = v
	}

	return &CustomLogger{
		consoleWriter: l.consoleWriter,
		fileWriter:    l.fileWriter,
		fields:        newFields,
		format:        l.format,
		useColors:     l.useColors,
	}
}

// log выводит лог
func (l *CustomLogger) log(level, msg string) {
	l.mu.RLock()
	fields := l.fields
	l.mu.RUnlock()

	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}
	shortFile := filepath.Base(file)

	now := time.Now()

	// Вывод в консоль
	if l.consoleWriter != nil {
		var output string
		switch l.format {
		case FormatJSON:
			output = l.formatJSON(now, level, msg, shortFile, line, fields)
		case FormatPlain:
			output = l.formatPlain(now, level, msg, shortFile, line, fields)
		default:
			output = l.formatPretty(now, level, msg, shortFile, line, fields, l.useColors)
		}
		fmt.Fprintln(l.consoleWriter, output)
	}

	// Вывод в файл (всегда plain без цветов)
	if l.fileWriter != nil {
		output := l.formatPlain(now, level, msg, shortFile, line, fields)
		fmt.Fprintln(l.fileWriter, output)
	}
}

// formatPretty форматирует лог в красивом цветном виде
func (l *CustomLogger) formatPretty(t time.Time, level, msg, file string, line int, fields map[string]any, colors bool) string {
	var levelColor, msgColor, resetColor, grayColor, cyanColor string

	if colors {
		resetColor = colorReset
		grayColor = colorGray
		cyanColor = colorCyan

		switch level {
		case "INFO":
			levelColor = colorBoldGreen
			msgColor = colorGreen
		case "WARN":
			levelColor = colorBoldCyan
			msgColor = cyanColor
		case "ERROR":
			levelColor = colorBoldRed
			msgColor = colorRed
		}
	}

	// Формат времени: 15:04:05.000
	timeStr := t.Format("15:04:05.000")

	// Форматируем уровень с фиксированной шириной
	levelStr := fmt.Sprintf("%-5s", level)

	// Базовая строка
	result := fmt.Sprintf("%s%s%s %s%s%s %s%s:%d%s %s▸%s %s%s%s",
		grayColor, timeStr, resetColor,
		levelColor, levelStr, resetColor,
		grayColor, file, line, resetColor,
		cyanColor, resetColor,
		msgColor, msg, resetColor,
	)

	// Добавляем поля
	if len(fields) > 0 {
		result += " "
		for k, v := range fields {
			result += fmt.Sprintf("%s%s%s=%s%v%s ", cyanColor, k, resetColor, grayColor, v, resetColor)
		}
	}

	return result
}

// formatJSON форматирует лог в JSON
func (l *CustomLogger) formatJSON(t time.Time, level, msg, file string, line int, fields map[string]any) string {
	entry := map[string]any{
		"time":    t.Format(time.RFC3339Nano),
		"level":   level,
		"message": msg,
		"caller":  fmt.Sprintf("%s:%d", file, line),
	}

	// Добавляем дополнительные поля
	for k, v := range fields {
		entry[k] = v
	}

	data, _ := json.Marshal(entry)
	return string(data)
}

// formatPlain форматирует лог в простом текстовом виде
func (l *CustomLogger) formatPlain(t time.Time, level, msg, file string, line int, fields map[string]any) string {
	timeStr := t.Format("2006-01-02 15:04:05.000")

	result := fmt.Sprintf("%s [%-5s] [%s:%d] %s", timeStr, level, file, line, msg)

	if len(fields) > 0 {
		result += " |"
		for k, v := range fields {
			result += fmt.Sprintf(" %s=%v", k, v)
		}
	}

	return result
}

// NopLogger - логгер-заглушка для тестов
type NopLogger struct{}

func NewNopLogger() *NopLogger                               { return &NopLogger{} }
func (l *NopLogger) Info(msg ...any)                         {}
func (l *NopLogger) Infof(format string, args ...any)        {}
func (l *NopLogger) Warn(msg ...any)                         {}
func (l *NopLogger) Warnf(format string, args ...any)        {}
func (l *NopLogger) Error(msg ...any)                        {}
func (l *NopLogger) Errorf(format string, args ...any)       {}
func (l *NopLogger) WithField(key string, value any) Logger  { return l }
func (l *NopLogger) WithFields(fields map[string]any) Logger { return l }
