package main

import "testing"

func TestGetConfigPathUsesEnvironmentOverride(t *testing.T) {
	t.Setenv("CONFIG_PATH", "custom-config.yml")

	if got := getConfigPath(); got != "custom-config.yml" {
		t.Fatalf("getConfigPath() = %q, want %q", got, "custom-config.yml")
	}
}

func TestGetConfigPathDefaultsToProductionConfig(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")

	if got := getConfigPath(); got != "./config/config.prod.yml" {
		t.Fatalf("getConfigPath() = %q, want %q", got, "./config/config.prod.yml")
	}
}

func TestGetAllowedOriginsIncludesConfiguredOrigins(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", " https://debateai.example, http://localhost:5173,https://preview.example ")

	got := getAllowedOrigins()
	want := []string{
		"http://localhost:5173",
		"https://debateai.example",
		"https://preview.example",
	}
	if len(got) != len(want) {
		t.Fatalf("getAllowedOrigins() = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("getAllowedOrigins() = %q, want %q", got, want)
		}
	}
}

func TestGetAllowedOriginsKeepsLocalhostByDefault(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "")

	got := getAllowedOrigins()
	if len(got) != 1 || got[0] != "http://localhost:5173" {
		t.Fatalf("getAllowedOrigins() = %q, want only localhost frontend", got)
	}
}
