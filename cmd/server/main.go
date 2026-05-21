package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/sanin7k/conveyor/internal/api"
	"github.com/sanin7k/conveyor/internal/dispatcher"
	"github.com/sanin7k/conveyor/internal/store"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	db, err := sql.Open(
		"postgres",
		"postgres://conveyor:conveyor@localhost:5432/conveyor?sslmode=disable",
	)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	s := store.New(db)

	d := dispatcher.New(ctx, s, 10, 4)

	mux := http.NewServeMux()

	h := api.NewHandler(d, s)
	h.RegisterRoutes(mux)

	server := http.Server{
		Addr: ":8080",
		Handler: mux,
	}

	d.Start()

	go func() {
		log.Println("server listening on :8080")

		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	fmt.Println()
	log.Println("signal received")
	log.Println("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	server.Shutdown(shutdownCtx)

	d.Shutdown()
}

