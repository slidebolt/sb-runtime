package runtime

import (
	"log"
	"log/slog"
	"os"
)

// SetupLogger configures structured JSON logging to stderr at DEBUG level.
// Called automatically by Run() before dispatching commands.
func SetupLogger(binaryID string) *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})).With("binary", binaryID)

	slog.SetDefault(logger)

	// Redirect standard log package to slog so log.Println etc. also
	// produce structured JSON on stderr.
	log.SetFlags(0)
	log.SetOutput(&slogWriter{logger: logger})

	return logger
}

// slogWriter adapts slog.Logger for use as an io.Writer so the standard
// log package emits structured JSON instead of plain text.
type slogWriter struct {
	logger *slog.Logger
}

func (w *slogWriter) Write(p []byte) (int, error) {
	// Trim trailing newline that log.Println adds.
	msg := string(p)
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}
	w.logger.Info(msg)
	return len(p), nil
}
