package utils

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

type (
	loggerKey        struct{}
	correlationIDKey struct{}
	requestIDKey     struct{}
)

var (
	loggerI    *slog.Logger
	loggerOnce sync.Once

	emailAddressPattern = regexp.MustCompile(
		`[A-Za-z0-9._%+\-]+@([A-Za-z0-9\-]+(?:\.[A-Za-z0-9\-]+)+)`,
	)
)

// Mail server replies quote addresses; the domain stays for debugging.
func redactEmailAddresses(_ []string, attr slog.Attr) slog.Attr {
	switch attr.Value.Kind() {
	case slog.KindString:
		attr.Value = slog.StringValue(RedactEmailAddresses(attr.Value.String()))
	case slog.KindAny:
		if err, isError := attr.Value.Any().(error); isError {
			attr.Value = slog.StringValue(RedactEmailAddresses(err.Error()))
		}
	}

	return attr
}

func RedactEmailAddresses(text string) string {
	return emailAddressPattern.ReplaceAllString(text, config.GetRedactedEmailReplacement())
}

type correlationHandler struct {
	slog.Handler
}

func (h correlationHandler) Handle(ctx context.Context, record slog.Record) error {
	if correlationID := GetCorrelationID(ctx); correlationID != "" {
		record.AddAttrs(slog.String("correlation_id", correlationID))
	}

	return h.Handler.Handle(ctx, record)
}

func (h correlationHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return correlationHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h correlationHandler) WithGroup(name string) slog.Handler {
	return correlationHandler{Handler: h.Handler.WithGroup(name)}
}

func InitLogger(mode string) *slog.Logger {
	loggerOnce.Do(func() {
		handler := slog.NewJSONHandler(
			logWriter(mode),
			&slog.HandlerOptions{
				Level:       parseLogLevel(config.GetLogLevel()),
				ReplaceAttr: redactEmailAddresses,
			},
		)
		loggerI = slog.New(correlationHandler{Handler: handler}).With(
			slog.String("service", config.GetServiceName()),
			slog.String("mode", mode),
		)
		slog.SetDefault(loggerI)
	})

	return loggerI
}

func logWriter(mode string) io.Writer {
	logDir := config.GetLogDir()
	if logDir == "" {
		return os.Stdout
	}

	return io.MultiWriter(os.Stdout, &lumberjack.Logger{
		Filename:   filepath.Join(logDir, mode+".log"),
		MaxSize:    config.GetLogMaxSizeMB(),
		MaxBackups: config.GetLogMaxBackups(),
		MaxAge:     config.GetLogMaxAgeDays(),
		Compress:   true,
	})
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func GetLogger(ctx context.Context) *slog.Logger {
	if logger, found := ctx.Value(loggerKey{}).(*slog.Logger); found && logger != nil {
		return logger
	}

	if loggerI != nil {
		return loggerI
	}

	return slog.Default()
}

func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey{}, correlationID)
}

func GetCorrelationID(ctx context.Context) string {
	correlationID, _ := ctx.Value(correlationIDKey{}).(string)

	return correlationID
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

func GetRequestID(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey{}).(string)

	return requestID
}

func LogDebug(ctx context.Context, message string, keyValues ...any) {
	GetLogger(ctx).DebugContext(ctx, message, keyValues...)
}

func LogInfo(ctx context.Context, message string, keyValues ...any) {
	GetLogger(ctx).InfoContext(ctx, message, keyValues...)
}

func LogWarn(ctx context.Context, message string, keyValues ...any) {
	GetLogger(ctx).WarnContext(ctx, message, keyValues...)
}

func LogError(ctx context.Context, err error, keyValues ...any) {
	GetLogger(ctx).ErrorContext(ctx, err.Error(), keyValues...)
}

func LogInfoWithoutCtx(message string, keyValues ...any) {
	LogInfo(context.Background(), message, keyValues...)
}

func LogErrorWithoutCtx(err error, keyValues ...any) {
	LogError(context.Background(), err, keyValues...)
}
