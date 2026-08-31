package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"time"
	"uuid"

	"log/slog"

	"github.com/ionburstcloud/go-gelf/gelf"
	sloggraylog "github.com/ionburstcloud/slog-graylog/v2"
)

func main() {
	// docker-compose up -d
	// or
	// ncat -l 12201 -u
	tlsConfig := &tls.Config{}
	gelfWriter, err := gelf.NewTLSWriter("orwell-eu2.ionburst.io:12202", tlsConfig)
	if err != nil {
		log.Fatalf("gelf.NewWriter: %s", err)
	}
	//w, _ := reflect.ValueOf(gelfWriter).Interface().(*gelf.TLSWriter)

	logger := slog.New(sloggraylog.Option{Level: slog.LevelDebug, Writer: gelfWriter}.NewGraylogHandler())
	logger = logger.With("service", "defined source")

	logger.
		With(
			slog.Group("user",
				slog.String("id", "user-123"),
				slog.Time("created_at", time.Now().AddDate(0, 0, -1)),
			),
		).
		With("environment", "dev").
		With("error", fmt.Errorf("an error")).
		Error("An error message")

	logger.Debug("A debug message", "Some_number", 14, "Some_text", "Blah blah")

	logger.With("Some_number", 14).
		With("Some_text", "Blah blah").
		With("Some_uuid", uuid.New()).
		Warn("A warning message")

	logger.With("Some_usefeul_number", 14).
		With("Some_useful_text", "Blah blah").
		With("Some_useful_uuid", uuid.New()).
		Info("An information message")

	logger.Info("Another informational", "Index", 5, "Reason", "Reason string", "UUID tag", uuid.New(), "UUID_tag", uuid.New())

	time.Sleep(time.Second * 5)

	fmt.Printf("End of example\n")
}
