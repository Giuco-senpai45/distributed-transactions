package main

import (
	"context"
	"dt/app"
	"dt/utils/log"
	"embed"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
)

//go:embed db/migrations/mvcc/*.sql
var mvcc_migrations embed.FS

//go:embed db/migrations/app/*.sql
var app_migrations embed.FS

func main() {
	log.Instantiate()

	migrations := []fs.FS{mvcc_migrations, app_migrations}
	a := app.New(migrations)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := a.Start(ctx); err != nil {
		log.Error("failed to start server", slog.Any("error", err))
	}
}
