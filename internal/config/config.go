package config

import (
	"fmt"
	"os"
)

type Config struct {
	GitHubToken     string
	YandexGPTAPIKey string
	YandexFolderID  string
	BranchPrefix    string
	MainBranch      string
	Verbose         bool
	Interactive     bool
}

func Load() (*Config, error) {
	cfg := &Config{
		GitHubToken:     os.Getenv("GITHUB_TOKEN"),
		YandexGPTAPIKey: os.Getenv("YANDEX_GPT_API_KEY"),
		YandexFolderID:  os.Getenv("YANDEX_FOLDER_ID"),
		MainBranch:      getEnvOrDefault("MAIN_BRANCH", "main"),
	}

	if cfg.GitHubToken == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}

	if cfg.YandexGPTAPIKey == "" {
		return nil, fmt.Errorf("YANDEX_GPT_API_KEY environment variable is required")
	}

	if cfg.YandexFolderID == "" {
		return nil, fmt.Errorf("YANDEX_FOLDER_ID environment variable is required")
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
