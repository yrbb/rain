package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/yrbb/rain/pkg/logger/writer"
)

type JSONOptions struct {
	Level  *slog.LevelVar
	Writer writer.Writer
}

func (o JSONOptions) NewJSONHandler() slog.Handler {
	if o.Writer == nil {
		panic("missing writer")
	}

	return &JSONHandler{
		option:  o,
		attrs:   []slog.Attr{},
		grainps: []string{},
	}
}

type JSONHandler struct {
	option  JSONOptions
	attrs   []slog.Attr
	grainps []string
}

func (h *JSONHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.option.Level.Level()
}

func (h *JSONHandler) Handle(ctx context.Context, r slog.Record) error {
	message := h.formatter(&r)

	bytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return h.option.Writer.Write(r.Level, append(bytes, byte('\n')))
}

func (h *JSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &JSONHandler{
		option:  h.option,
		attrs:   appendAttrsToGroup(h.grainps, h.attrs, attrs),
		grainps: h.grainps,
	}
}

func (h *JSONHandler) WithGroup(name string) slog.Handler {
	return &JSONHandler{
		option:  h.option,
		attrs:   h.attrs,
		grainps: append(h.grainps, name),
	}
}

func (h *JSONHandler) formatter(r *slog.Record) map[string]any {
	log := map[string]any{
		"timestamp": r.Time.Format("2006-01-02 15:04:05.000000"),
		"level":     r.Level.String(),
		"message":   r.Message,
	}

	for k, v := range attrsToValue(h.attrs) {
		log[k] = v
	}

	fs := runtime.CallersFrames([]uintptr{r.PC})
	f, _ := fs.Next()
	if f.File != "" {
		_, file := filepath.Split(f.File)
		log["file"] = file + ":" + strconv.Itoa(f.Line)
	}

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

	log["fields"] = fields

	return log
}
