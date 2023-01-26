package wool

import (
	"golang.org/x/exp/slog"
	"sync/atomic"
)

var logger atomic.Value

func init() {
	logger.Store(slog.Default().WithGroup("wool"))
}

func Logger() *slog.Logger {
	return logger.Load().(*slog.Logger)
}

func SetLogger(l *slog.Logger) {
	logger.Store(l.WithGroup("wool"))
}
