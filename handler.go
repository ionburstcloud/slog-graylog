package sloggraylog

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"github.com/ionburstcloud/go-gelf/gelf"
	slogcommon "github.com/samber/slog-common"
)

type Option struct {
	// log level (default: debug)
	Level slog.Leveler

	// connection to graylog
	Writer gelf.Writer

	// optional: customize json payload builder
	Converter Converter
	// optional: fetch attributes from context
	AttrFromContext []func(ctx context.Context) []slog.Attr

	// optional: see slog.HandlerOptions
	AddSource   bool
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr

	Facility string

	// internal
	hostname string
}

func (o Option) NewGraylogHandler() slog.Handler {
	if o.Level == nil {
		o.Level = slog.LevelDebug
	}

	if o.Writer == nil {
		panic("missing graylog connections")
	}
	if o.Facility == "" {
		o.Facility = fmt.Sprintf("%s/%s", name, version)
	}

	if o.Converter == nil {
		o.Converter = DefaultConverter
	}

	if o.AttrFromContext == nil {
		o.AttrFromContext = []func(ctx context.Context) []slog.Attr{}
	}

	if hostname, err := os.Hostname(); err == nil {
		o.hostname = hostname
	}

	return &GraylogHandler{
		option: o,
		attrs:  []slog.Attr{},
		groups: []string{},
	}
}

var _ slog.Handler = (*GraylogHandler)(nil)

type GraylogHandler struct {
	option Option
	attrs  []slog.Attr
	groups []string
}

func (h *GraylogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.option.Level.Level()
}

func (h *GraylogHandler) Handle(ctx context.Context, record slog.Record) error {
	fromContext := slogcommon.ContextExtractor(ctx, h.option.AttrFromContext)
	extra := h.option.Converter(h.option.AddSource, h.option.ReplaceAttr, append(h.attrs, fromContext...), h.groups, &record)

	msg := &gelf.Message{
		Version:  "1.1",
		Host:     h.option.hostname,
		Short:    short(&record),
		Full:     strings.TrimSpace(record.Message),
		TimeUnix: makeTimestamp(),
		Level:    LogLevels[record.Level],
		Facility: h.option.Facility,
		Extra:    extra,
	}

	// non-blocking
	go func() {
		_ = h.option.Writer.WriteMessage(msg)
	}()

	return nil
}

func (h *GraylogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &GraylogHandler{
		option: h.option,
		attrs:  slogcommon.AppendAttrsToGroup(h.groups, h.attrs, attrs...),
		groups: h.groups,
	}
}

func (h *GraylogHandler) WithGroup(name string) slog.Handler {
	// https://cs.opensource.google/go/x/exp/+/46b07846:slog/handler.go;l=247
	if name == "" {
		return h
	}

	return &GraylogHandler{
		option: h.option,
		attrs:  h.attrs,
		groups: append(h.groups, name),
	}
}

func short(record *slog.Record) string {
	msg := strings.TrimSpace(record.Message)
	if i := strings.IndexRune(msg, '\n'); i > 0 {
		return msg[:i]
	}

	return msg
}

func makeTimestamp() float64 {
	var millisecInt int64 = time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
	var secF float64 = float64(millisecInt) / float64(1000)
	var secString = fmt.Sprintf("%.3f", secF)
	result, err := strconv.ParseFloat(secString, 64)
	if err != nil {
		fmt.Printf("ParseFloat error: %s\n", err.Error())
	}

	return result
}
