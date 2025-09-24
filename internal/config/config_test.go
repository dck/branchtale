package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Save original env vars
	originalGitHubToken := os.Getenv("GITHUB_TOKEN")
	originalYandexAPIKey := os.Getenv("YANDEX_GPT_API_KEY")
	originalYandexFolderID := os.Getenv("YANDEX_FOLDER_ID")

	defer func() {
		os.Setenv("GITHUB_TOKEN", originalGitHubToken)
		os.Setenv("YANDEX_GPT_API_KEY", originalYandexAPIKey)
		os.Setenv("YANDEX_FOLDER_ID", originalYandexFolderID)
	}()

	t.Run("success with all required env vars", func(t *testing.T) {
		os.Setenv("GITHUB_TOKEN", "test-github-token")
		os.Setenv("YANDEX_GPT_API_KEY", "test-yandex-key")
		os.Setenv("YANDEX_FOLDER_ID", "test-folder-id")
		os.Setenv("CONTENT_GENERATION", "local")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if cfg.GitHubToken != "test-github-token" {
			t.Errorf("expected GitHubToken 'test-github-token', got '%s'", cfg.GitHubToken)
		}

		if cfg.YandexGPTAPIKey != "test-yandex-key" {
			t.Errorf("expected YandexGPTAPIKey 'test-yandex-key', got '%s'", cfg.YandexGPTAPIKey)
		}

		if cfg.YandexFolderID != "test-folder-id" {
			t.Errorf("expected YandexFolderID 'test-folder-id', got '%s'", cfg.YandexFolderID)
		}

		if cfg.ContentGeneration != "local" {
			t.Errorf("expected ContentGeneration 'local', got '%s'", cfg.ContentGeneration)
		}

		if cfg.UseAI {
			t.Errorf("expected UseAI to be false, got true")
		}
	})

	t.Run("custom main branch", func(t *testing.T) {
		os.Setenv("GITHUB_TOKEN", "test-github-token")
		os.Setenv("YANDEX_GPT_API_KEY", "test-yandex-key")
		os.Setenv("YANDEX_FOLDER_ID", "test-folder-id")
		os.Setenv("CONTENT_GENERATION", "local")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if cfg.ContentGeneration != "local" {
			t.Errorf("expected ContentGeneration 'local', got '%s'", cfg.ContentGeneration)
		}

		if cfg.UseAI {
			t.Errorf("expected UseAI to be false, got true")
		}
	})

	t.Run("ai mode enabled", func(t *testing.T) {
		os.Setenv("GITHUB_TOKEN", "test-github-token")
		os.Setenv("YANDEX_GPT_API_KEY", "test-yandex-key")
		os.Setenv("YANDEX_FOLDER_ID", "test-folder-id")
		os.Setenv("CONTENT_GENERATION", "yandex")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if cfg.ContentGeneration != "yandex" {
			t.Errorf("expected ContentGeneration 'yandex', got '%s'", cfg.ContentGeneration)
		}

		if !cfg.UseAI {
			t.Errorf("expected UseAI to be true, got false")
		}
	})

	t.Run("missing GITHUB_TOKEN", func(t *testing.T) {
		os.Unsetenv("GITHUB_TOKEN")
		os.Setenv("YANDEX_GPT_API_KEY", "test-yandex-key")
		os.Setenv("YANDEX_FOLDER_ID", "test-folder-id")
		os.Setenv("CONTENT_GENERATION", "yandex")

		_, err := Load()
		if err == nil {
			t.Fatal("expected error for missing GITHUB_TOKEN")
		}

		expected := "GITHUB_TOKEN environment variable is required"
		if err.Error() != expected {
			t.Errorf("expected error '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("missing YANDEX_GPT_API_KEY", func(t *testing.T) {
		os.Setenv("GITHUB_TOKEN", "test-github-token")
		os.Unsetenv("YANDEX_GPT_API_KEY")
		os.Setenv("YANDEX_FOLDER_ID", "test-folder-id")
		os.Setenv("CONTENT_GENERATION", "yandex")

		_, err := Load()
		if err == nil {
			t.Fatal("expected error for missing YANDEX_GPT_API_KEY")
		}

		expected := "YANDEX_GPT_API_KEY environment variable is required"
		if err.Error() != expected {
			t.Errorf("expected error '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("missing YANDEX_FOLDER_ID", func(t *testing.T) {
		os.Setenv("GITHUB_TOKEN", "test-github-token")
		os.Setenv("YANDEX_GPT_API_KEY", "test-yandex-key")
		os.Unsetenv("YANDEX_FOLDER_ID")
		os.Setenv("CONTENT_GENERATION", "yandex")

		_, err := Load()
		if err == nil {
			t.Fatal("expected error for missing YANDEX_FOLDER_ID")
		}

		expected := "YANDEX_FOLDER_ID environment variable is required"
		if err.Error() != expected {
			t.Errorf("expected error '%s', got '%s'", expected, err.Error())
		}
	})
}
