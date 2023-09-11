package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/yrbb/rain/pkg/logger/writer"
)

const (
	ansiReset          = "\033[0m"
	ansiFaint          = "\033[2m"
	ansiResetFaint     = "\033[22m"
	ansiBrightRed      = "\033[91m"
	ansiBrightGreen    = "\033[92m"
	ansiBrightYellow   = "\033[93m"
	ansiBrightRedFaint = "\033[91;2m"
)

type TextOptions struct {
	Level  *slog.LevelVar
	Writer writer.Writer
}

func (o TextOptions) NewTextHandler() slog.Handler {
	if o.Writer == nil {
		panic("missing writer")
	}

	return &TextHandler{
		option:  o,
		attrs:   []slog.Attr{},
		grainps: []string{},
	}
}

type TextHandler struct {
	mu sync.Mutex

	option  TextOptions
	attrs   []slog.Attr
	grainps []string
}

func (h *TextHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.option.Level.Level()
}

func (h *TextHandler) Handle(ctx context.Context, r slog.Record) error {
	buf := newBuffer()
	defer buf.Free()

	if !r.Time.IsZero() {
		h.appendTime(buf, r.Time)
		buf.WriteByte(' ')
	}

	h.appendLevel(buf, r.Level)
	buf.WriteByte(' ')

	// write source
	fs := runtime.CallersFrames([]uintptr{r.PC})
	f, _ := fs.Next()
	if f.File != "" {
		h.appendSource(buf, f)
		buf.WriteByte(' ')
	}

	// write message
	buf.WriteString(r.Message)
	buf.WriteByte(' ')

	// if len(h.attrs) > 0 {
	// 	fields := map[string]any{}
	// 	for k, v := range attrsToValue(h.attrs) {
	// 		fields[k] = v
	// 	}
	//
	// 	bts, err := json.Marshal(fields)
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	buf.Write(bts)
	// 	buf.WriteByte(' ')
	// }

	// write attributes
	fields := map[string]any{}
	r.Attrs(func(attr slog.Attr) bool {
		for k, v := range attrsToValue([]slog.Attr{attr}) {
			if k != "!BADKEY" {
				fields[k] = v
				continue
			}

			if mv, ok := v.(map[string]any); ok {
				for key, val := range mv {
					fields[key] = val
				}

				continue
			}

			fields[k] = v
		}

		return true
	})
	if len(fields) > 0 {
		bts, err := json.Marshal(fields)
		if err != nil {
			return err
		}
		buf.Write(bts)
		buf.WriteByte(' ')
	}

	if len(*buf) == 0 {
		return nil
	}
	(*buf)[len(*buf)-1] = '\n'

	h.mu.Lock()
	defer h.mu.Unlock()

	return h.option.Writer.Write(r.Level, *buf)
}

func (h *TextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TextHandler{
		option:  h.option,
		attrs:   appendAttrsToGroup(h.grainps, h.attrs, attrs),
		grainps: h.grainps,
	}
}

func (h *TextHandler) WithGroup(name string) slog.Handler {
	return &TextHandler{
		option:  h.option,
		attrs:   h.attrs,
		grainps: append(h.grainps, name),
	}
}

func (h *TextHandler) appendTime(buf *buffer, t time.Time) {
	buf.WriteString(ansiFaint)
	*buf = t.AppendFormat(*buf, "01-02 15:04:05")
	buf.WriteString(ansiReset)
}

func (h *TextHandler) appendLevel(buf *buffer, level slog.Level) {
	delta := func(buf *buffer, val slog.Level) {
		if val == 0 {
			return
		}
		buf.WriteByte('+')
		*buf = strconv.AppendInt(*buf, int64(val), 10)
	}

	switch {
	case level < slog.LevelInfo:
		buf.WriteString("DEBUG")
		delta(buf, level-slog.LevelDebug)
	case level < slog.LevelWarn:
		buf.WriteString(ansiBrightGreen)
		buf.WriteString("INFO")
		delta(buf, level-slog.LevelInfo)
		buf.WriteString(ansiReset)
	case level < slog.LevelError:
		buf.WriteString(ansiBrightYellow)
		buf.WriteString("WARN")
		delta(buf, level-slog.LevelWarn)
		buf.WriteString(ansiReset)
	default:
		buf.WriteString(ansiBrightRed)
		buf.WriteString("ERROR")
		delta(buf, level-slog.LevelError)
		buf.WriteString(ansiReset)
	}
}

func (h *TextHandler) appendSource(buf *buffer, f runtime.Frame) {
	_, file := filepath.Split(f.File)

	buf.WriteString(ansiFaint)
	buf.WriteString(file)
	buf.WriteByte(':')
	buf.WriteString(strconv.Itoa(f.Line))
	buf.WriteString(ansiReset)
}
