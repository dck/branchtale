package ai

import (
	"testing"
)

func TestYandexGPT_NewYandexGPT(t *testing.T) {
	apiKey := "test-api-key"
	folderID := "test-folder-id"

	yandex := NewYandexGPT(apiKey, folderID)

	if yandex.APIKey != apiKey {
		t.Errorf("Expected APIKey %s, got %s", apiKey, yandex.APIKey)
	}

	if yandex.FolderID != folderID {
		t.Errorf("Expected FolderID %s, got %s", folderID, yandex.FolderID)
	}

	if yandex.client == nil {
		t.Error("Expected client to be initialized")
	}
}

func TestYandexGPTRequest_Structure(t *testing.T) {
	req := YandexGPTRequest{
		ModelURI: "gpt://test-folder/yandexgpt-lite/latest",
		CompletionOptions: CompletionOptions{
			Stream:      false,
			Temperature: 0.3,
			MaxTokens:   500,
		},
		Messages: []Message{
			{
				Role: "user",
				Text: "test prompt",
			},
		},
	}

	if req.ModelURI != "gpt://test-folder/yandexgpt-lite/latest" {
		t.Errorf("Unexpected ModelURI: %s", req.ModelURI)
	}

	if req.CompletionOptions.Temperature != 0.3 {
		t.Errorf("Expected temperature 0.3, got %f", req.CompletionOptions.Temperature)
	}

	if len(req.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(req.Messages))
	}

	if req.Messages[0].Role != "user" {
		t.Errorf("Expected role 'user', got %s", req.Messages[0].Role)
	}
}

func TestYandexGPT_CompletionOptions(t *testing.T) {
	opts := CompletionOptions{
		Stream:      false,
		Temperature: 0.3,
		MaxTokens:   500,
	}

	if opts.Stream != false {
		t.Errorf("Expected Stream to be false, got %t", opts.Stream)
	}

	if opts.Temperature != 0.3 {
		t.Errorf("Expected Temperature to be 0.3, got %f", opts.Temperature)
	}

	if opts.MaxTokens != 500 {
		t.Errorf("Expected MaxTokens to be 500, got %d", opts.MaxTokens)
	}
}

func TestMessage_Structure(t *testing.T) {
	msg := Message{
		Role: "user",
		Text: "Generate a branch name",
	}

	if msg.Role != "user" {
		t.Errorf("Expected Role to be 'user', got %s", msg.Role)
	}

	if msg.Text != "Generate a branch name" {
		t.Errorf("Expected Text to be 'Generate a branch name', got %s", msg.Text)
	}
}
