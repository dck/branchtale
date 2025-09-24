package config

import (
	"fmt"
	"os"
)

type Config struct {
	GitHubToken       string
	YandexGPTAPIKey   string
	YandexFolderID    string
	BranchPrefix      string
	Verbose           bool
	Interactive       bool
	ContentGeneration string
	UseAI             bool
}

func Load() (*Config, error) {
	cfg := &Config{
		GitHubToken:       os.Getenv("GITHUB_TOKEN"),
		YandexGPTAPIKey:   os.Getenv("YANDEX_GPT_API_KEY"),
		YandexFolderID:    os.Getenv("YANDEX_FOLDER_ID"),
		ContentGeneration: os.Getenv("CONTENT_GENERATION"),
		UseAI:             false,
	}

	if cfg.GitHubToken == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}

	if cfg.ContentGeneration != "" && cfg.ContentGeneration != "local" {
		if cfg.YandexGPTAPIKey == "" {
			return nil, fmt.Errorf("YANDEX_GPT_API_KEY environment variable is required")
		}

		if cfg.YandexFolderID == "" {
			return nil, fmt.Errorf("YANDEX_FOLDER_ID environment variable is required")
		}
		cfg.UseAI = true
	}

	return cfg, nil
}
