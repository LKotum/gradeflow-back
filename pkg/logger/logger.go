package logger

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"gradeflow/internal/config"
)

type rotatingWriter struct {
	path       string
	maxSize    int64
	maxBackups int
	maxAge     time.Duration

	mu   sync.Mutex
	file *os.File
	size int64
}

func newRotatingWriter(path string, maxSizeMB, maxBackups, maxAgeDays int) (*rotatingWriter, error) {
	if path == "" {
		return nil, errors.New("log path is empty")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir %s: %w", dir, err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("stat log file: %w", err)
	}
	rw := &rotatingWriter{
		path:       path,
		file:       f,
		size:       info.Size(),
		maxSize:    int64(maxSizeMB) * 1024 * 1024,
		maxBackups: maxBackups,
	}
	if rw.maxSize <= 0 {
		rw.maxSize = 20 * 1024 * 1024
	}
	if rw.maxBackups <= 0 {
		rw.maxBackups = 10
	}
	if maxAgeDays > 0 {
		rw.maxAge = time.Duration(maxAgeDays) * 24 * time.Hour
	}

	rw.mu.Lock()
	defer rw.mu.Unlock()
	if rw.size >= rw.maxSize {
		if err := rw.rotateLocked(); err != nil {
			return nil, err
		}
	}
	return rw, nil
}

func (w *rotatingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		if err := w.openLocked(); err != nil {
			return 0, err
		}
	}
	if w.maxSize > 0 && w.size+int64(len(p)) > w.maxSize {
		if err := w.rotateLocked(); err != nil {
			return 0, err
		}
	}
	n, err := w.file.Write(p)
	w.size += int64(n)
	return n, err
}

func (w *rotatingWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}

func (w *rotatingWriter) rotateLocked() error {
	if w.file == nil {
		return w.openLocked()
	}
	if err := w.file.Close(); err != nil {
		return err
	}
	timestamp := time.Now().UTC().Format("20060102-150405")
	rotated := fmt.Sprintf("%s.%s", w.path, timestamp)
	if err := os.Rename(w.path, rotated); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("rotate rename: %w", err)
	}
	if err := w.openLocked(); err != nil {
		return err
	}
	if err := w.purgeOldBackupsLocked(); err != nil {
		return err
	}
	return nil
}

func (w *rotatingWriter) openLocked() error {
	f, err := os.OpenFile(w.path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	w.file = f
	w.size = 0
	return nil
}

func (w *rotatingWriter) purgeOldBackupsLocked() error {
	pattern := fmt.Sprintf("%s.*", w.path)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("glob backups: %w", err)
	}
	if len(matches) == 0 {
		return nil
	}
	type backupInfo struct {
		path    string
		modTime time.Time
	}
	backups := make([]backupInfo, 0, len(matches))
	now := time.Now()
	for _, p := range matches {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if w.maxAge > 0 && now.Sub(info.ModTime()) > w.maxAge {
			_ = os.Remove(p)
			continue
		}
		backups = append(backups, backupInfo{path: p, modTime: info.ModTime()})
	}
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].modTime.After(backups[j].modTime)
	})
	for idx := w.maxBackups; idx < len(backups); idx++ {
		_ = os.Remove(backups[idx].path)
	}
	return nil
}

var (
	globalLogger *slog.Logger
	globalOnce   sync.Once
	rotWriter    *rotatingWriter
)

// Init configures the global logger using provided logging configuration.
func Init(cfg config.LoggingConfig) (*slog.Logger, error) {
	var initErr error
	globalOnce.Do(func() {
		var writers []io.Writer
		if cfg.Path != "" {
			var err error
			rotWriter, err = newRotatingWriter(cfg.Path, cfg.MaxSizeMB, cfg.MaxBackups, cfg.MaxAgeDays)
			if err != nil {
				initErr = err
				return
			}
			writers = append(writers, rotWriter)
		}
		if cfg.AlsoStdout || len(writers) == 0 {
			writers = append(writers, os.Stdout)
		}
		output := io.MultiWriter(writers...)
		handler := slog.NewJSONHandler(output, &slog.HandlerOptions{
			AddSource: true,
			Level:     parseLevel(cfg.Level),
			ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
				if attr.Key == slog.TimeKey {
					t := attr.Value.Time()
					attr.Value = slog.StringValue(t.UTC().Format(time.RFC3339Nano))
				}
				return attr
			},
		})
		globalLogger = slog.New(handler)
	})
	if initErr != nil {
		return nil, initErr
	}
	if globalLogger == nil {
		return slog.Default(), nil
	}
	return globalLogger, nil
}

func parseLevel(level string) slog.Leveler {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "info", "":
		return slog.LevelInfo
	default:
		return slog.LevelInfo
	}
}

// L returns the configured global logger.
func L() *slog.Logger {
	if globalLogger != nil {
		return globalLogger
	}
	return slog.Default()
}

// Sync flushes buffered log data and closes the rotating writer if configured.
func Sync() {
	if rotWriter != nil {
		_ = rotWriter.Close()
	}
}

func Debug(msg string, args ...any) {
	L().Debug(msg, args...)
}

func Info(msg string, args ...any) {
	L().Info(msg, args...)
}

func Warn(msg string, args ...any) {
	L().Warn(msg, args...)
}

func Error(msg string, args ...any) {
	L().Error(msg, args...)
}

func Fatal(msg string, args ...any) {
	L().Error(msg, args...)
	Sync()
	os.Exit(1)
}
