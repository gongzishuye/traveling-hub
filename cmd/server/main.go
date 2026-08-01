package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/faria/traveling-hub/api"
	"github.com/faria/traveling-hub/internal/journey"
	appplatform "github.com/faria/traveling-hub/internal/platform/app"
	"github.com/faria/traveling-hub/internal/platform/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	application, err := appplatform.New(context.Background(), cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer application.Close()
	worker, err := journey.NewWorker(application.DB, application.Journeys, cfg.JourneyWorkerInterval, cfg.JourneyWorkerBatch)
	if err != nil {
		log.Fatal(err)
	}
	workerContext, stopWorker := context.WithCancel(context.Background())
	defer stopWorker()
	go worker.Run(workerContext)
	server := &http.Server{Addr: cfg.HTTPAddr, Handler: api.NewRouter(application), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Printf("travelinghub listening on %s", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
	stopWorker()
}
