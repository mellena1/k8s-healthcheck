package healthcheck

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
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

func (c Checker) RunForever(ctx context.Context, logger *slog.Logger) {
	ticker := time.NewTicker(time.Duration(c.check.CheckFrequency))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("exiting checker", "error", ctx.Err())
			return
		case <-ticker.C:
			if err := c.healthcheck(ctx); err != nil {
				logger.Warn("failed health check", "error", err)
			} else {
				logger.Info("health check ok")
			}
		}
	}
}

func (c Checker) healthcheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.check.HTTPEndpoint(), nil)
	if err != nil {
		return fmt.Errorf("failed to make req: %w", err)
	}

	for k, v := range c.check.ExtraHeaders {
		if strings.ToLower(k) == "host" {
			// host is handled differently in net/http
			// see: https://stackoverflow.com/a/50559020
			req.Host = v
		} else {
			req.Header.Set(k, v)
		}
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
