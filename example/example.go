package main

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"log/slog"

	"github.com/ionburstcloud/go-gelf/gelf"
	sloggraylog "github.com/ionburstcloud/slog-graylog/v2"
)

func main() {
	// docker-compose up -d
	// or
	// ncat -l 12201 -u
	gelfWriter, err := gelf.NewTCPWriter("orwell-internal.ionburst.io:12202")
	if err != nil {
		log.Fatalf("gelf.NewWriter: %s", err)
	}
	tcpw, _ := reflect.ValueOf(gelfWriter).Interface().(*gelf.TCPWriter)

	logger := slog.New(sloggraylog.Option{Level: slog.LevelDebug, Writer: tcpw}.NewGraylogHandler())
	logger = logger.With("release", "v1.0.0")

	logger.
		With(
			slog.Group("user",
				slog.String("id", "user-123"),
				slog.Time("created_at", time.Now().AddDate(0, 0, -1)),
			),
		).
		With("environment", "dev").
		With("error", fmt.Errorf("an error")).
		Error("A message")
}
