package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type duration time.Duration

func (d *duration) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		b = b[1 : len(b)-1]
	}

	dur, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}

	*d = duration(dur)
	return nil
}

type ServiceCheck struct {
	Namespace    string            `json:"namespace"`
	Service      string            `json:"service"`
	Port         int               `json:"port"`
	Path         string            `json:"path"`
	ExtraHeaders map[string]string `json:"extraHeaders"`

	HealthCheckUUID string   `json:"healthCheckUUID"`
	CheckFrequency  duration `json:"checkFrequency"`
}

func (sc ServiceCheck) HTTPEndpoint() string {
	return fmt.Sprintf("http://%s.%s.svc.cluster.local:%d%s", sc.Service, sc.Namespace, sc.Port, sc.Path)
}

func (sc ServiceCheck) HealthCheckEndpoint() string {
	return fmt.Sprintf("https://hc-ping.com/%s", sc.HealthCheckUUID)
}

func (sc ServiceCheck) String() string {
	return sc.HTTPEndpoint()
}

type Config struct {
	Checks []ServiceCheck `json:"checks"`
}

func ReadConfigFromFile(filepath string) (Config, error) {
	contents, err := os.ReadFile(filepath)
	if err != nil {
		return Config{}, fmt.Errorf("error reading file %q: %w", filepath, err)
	}

	cfg := Config{}
	err = json.Unmarshal(contents, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("error unmarshalling to json: %w", err)
	}

	return cfg, nil
}
