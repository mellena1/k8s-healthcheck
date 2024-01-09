package healthcheck

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/mellena1/k8s-healthcheck/config"
)

type Checker struct {
	client *http.Client
	check  config.ServiceCheck
}

func NewChecker(client *http.Client, check config.ServiceCheck) Checker {
	return Checker{
		client: client,
		check:  check,
	}
}

func (c Checker) RunForever(ctx context.Context, logger *slog.Logger, frequency time.Duration) {
	lastCheckTime := time.Now().Add(-1000 * time.Minute)

	for {
		if err := ctx.Err(); err != nil {
			logger.Info("exiting checker", "error", err)
			return
		}

		if time.Since(lastCheckTime) < frequency {
			continue
		}

		if err := c.healthcheck(ctx); err != nil {
			logger.Warn("failed health check", "error", err)
		} else {
			logger.Info("health check ok")
		}
		lastCheckTime = time.Now()

		time.Sleep(1 * time.Second)
	}
}

func (c Checker) healthcheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.check.HTTPEndpoint(), nil)
	if err != nil {
		return fmt.Errorf("failed to make req: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("%s failed health check: %w", c.check, err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("non-200 status code (%d) from %s", resp.StatusCode, c.check)
	}

	req, err = http.NewRequestWithContext(ctx, http.MethodHead, c.check.HealthCheckEndpoint(), nil)
	if err != nil {
		return fmt.Errorf("failed to make healh check req: %w", err)
	}
	resp, err = c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send ping to healthchecks.io: %w", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("non-200 status code (%d) from healthchecks.io", resp.StatusCode)
	}

	return nil
}
