package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"log/slog"

	"github.com/ionburstcloud/go-gelf/gelf"
	sloggraylog "github.com/ionburstcloud/slog-graylog/v2"
)

func main() {
	// docker-compose up -d
	// or
	// ncat -l 12201 -u
	tlsConfig := &tls.Config{}
	gelfWriter, err := gelf.NewTLSWriter("orwell-internal.ionburst.io:12202", tlsConfig)
	if err != nil {
		log.Fatalf("gelf.NewWriter: %s", err)
	}
	//w, _ := reflect.ValueOf(gelfWriter).Interface().(*gelf.TLSWriter)

	logger := slog.New(sloggraylog.Option{Level: slog.LevelDebug, Writer: gelfWriter}.NewGraylogHandler())
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

	time.Sleep(time.Second * 5)

	fmt.Printf("End of example\n")
}
