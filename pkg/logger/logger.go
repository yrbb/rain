package logger

import (
	"log/slog"
	"runtime"
	"strings"
	"time"

	"github.com/yrbb/rain/pkg/logger/handler"
	"github.com/yrbb/rain/pkg/logger/writer"
)

var (
	lWriter writer.Writer
	gLevel  *slog.LevelVar
	dLogger *slog.Logger
	mLogger *slog.Logger
)

func Init(project string, lvl string, path string, splitTime time.Duration) {
	lWriter = writer.NewFileWriter(project, path, splitTime)

	gLevel = &slog.LevelVar{}
	WithLevel(lvl)

	textHandler := handler.TextOptions{
		Level:  gLevel,
		Writer: writer.NewConsoleWriter(),
	}.NewTextHandler()

	jsonHandler := handler.JSONOptions{
		Level:  gLevel,
		Writer: lWriter,
	}.NewJSONHandler().WithAttrs([]slog.Attr{slog.String("project", project)})

	dLogger = slog.New(jsonHandler)
	mLogger = slog.New(handler.NewMultiHandler(textHandler, jsonHandler))
	slog.SetDefault(dLogger)

	runtime.SetFinalizer(lWriter, func(w writer.Writer) {
		w.Close()
	})
}

func WithDebug(d bool) {
	if !d {
		slog.SetDefault(dLogger)
		return
	}

	slog.SetDefault(mLogger)
}

func WithLevel(lvl string) {
	switch strings.ToLower(lvl) {
	case "debug":
		gLevel.Set(slog.LevelDebug)
	case "info":
		gLevel.Set(slog.LevelInfo)
	case "warn":
		gLevel.Set(slog.LevelWarn)
	case "error":
		gLevel.Set(slog.LevelError)
	}
}

func WithSplitTime(st time.Duration) {
	lWriter.(*writer.LevelWriter).WithSplitTime(st)
}

func M() *slog.Logger {
	return mLogger
}

func Close() {
	lWriter.Close()
}
