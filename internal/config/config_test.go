package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Ensure env is clean for keys we validate
	_ = os.Unsetenv("SERVER_HOST")
	_ = os.Unsetenv("SERVER_PORT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Fatalf("unexpected server host: %s", cfg.Server.Host)
	}
	if cfg.Server.Port != "8080" {
		t.Fatalf("unexpected server port: %s", cfg.Server.Port)
	}
}
