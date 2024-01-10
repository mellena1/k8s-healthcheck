package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/mellena1/k8s-healthcheck/config"
	"github.com/mellena1/k8s-healthcheck/healthcheck"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.ReadConfigFromFile("config.json")
	if err != nil {
		logger.Error("failed to read config", "error", err)
		os.Exit(1)
	}

	httpClient := makeClient()

	ctx := context.Background()
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt)

	wg := runCheckersForever(ctx, logger, httpClient, cfg)

	startHttpServer(logger)

	wg.Wait()
	logger.Info("exiting...")
}

// runs an http server with a health check endpoint
func startHttpServer(logger *slog.Logger) {
	go func() {
		err := http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("ok"))
		}))
		if err != nil {
			logger.Error("http server failed", "error", err)
		}
	}()
}

func makeClient() *http.Client {
	client := retryablehttp.NewClient()
	client.Logger = nil
	client.RetryMax = 2
	client.HTTPClient.Timeout = 5 * time.Second
	return client.StandardClient()
}

func runCheckersForever(ctx context.Context, logger *slog.Logger, client *http.Client, cfg config.Config) *sync.WaitGroup {
	wg := &sync.WaitGroup{}

	for _, sc := range cfg.Checks {
		wg.Add(1)
		go func(check config.ServiceCheck) {
			checkLogger := logger.With("check", check.String())
			checkLogger.Info("starting checker")
			sc := healthcheck.NewChecker(client, check)
			sc.RunForever(ctx, checkLogger)
			wg.Done()
		}(sc)
	}

	return wg
}
