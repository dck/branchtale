package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type YandexGPT struct {
	APIKey   string
	FolderID string
	client   *http.Client
}

type YandexGPTRequest struct {
	ModelURI          string            `json:"modelUri"`
	CompletionOptions CompletionOptions `json:"completionOptions"`
	Messages          []Message         `json:"messages"`
}

type CompletionOptions struct {
	Stream      bool    `json:"stream"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"maxTokens"`
}

type Message struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type YandexGPTResponse struct {
	Result Result `json:"result"`
}

type Result struct {
	Alternatives []Alternative `json:"alternatives"`
}

type Alternative struct {
	Message Message `json:"message"`
	Status  string  `json:"status"`
}

func NewYandexGPT(apiKey, folderID string) *YandexGPT {
	return &YandexGPT{
		APIKey:   apiKey,
		FolderID: folderID,
		client:   &http.Client{},
	}
}

func (y *YandexGPT) GeneratePRTitle(ctx context.Context, diff string) (string, error) {
	prompt := fmt.Sprintf(
		"Generate a concise and descriptive pull request title based on the following git diff. "+
			"The title should be in imperative mood, start with a verb, and be under 70 characters:\n\n%s\n\n"+
			"Return only the title, no additional text.",
		diff,
	)

	return y.generateText(ctx, prompt)
}

func (y *YandexGPT) GeneratePRDescription(ctx context.Context, diff string) (string, error) {
	prompt := fmt.Sprintf(
		"Generate a pull request description based on the following git diff.\n\n"+
			"Format the response in markdown:\n\n%s",
		diff,
	)

	return y.generateText(ctx, prompt)
}

func (y *YandexGPT) GenerateBranchName(ctx context.Context, diff string) (string, error) {
	prompt := "Generate a short branch name based on the following git diff. The name should:\n" +
		"- Be descriptive but concise\n" +
		"- Use kebab-case (lowercase with hyphens)\n" +
		"- Be under 40 characters\n" +
		"- Not include any prefixes\n\n" +
		diff +
		"\n\nReturn only the branch name, no additional text."

	return y.generateText(ctx, prompt)
}

func (y *YandexGPT) generateText(ctx context.Context, prompt string) (string, error) {
	reqBody := YandexGPTRequest{
		ModelURI: fmt.Sprintf("gpt://%s/yandexgpt-lite/latest", y.FolderID),
		CompletionOptions: CompletionOptions{
			Stream:      false,
			Temperature: 0.3,
			MaxTokens:   500,
		},
		Messages: []Message{
			{
				Role: "user",
				Text: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://llm.api.cloud.yandex.net/foundationModels/v1/completion", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", y.APIKey))

	resp, err := y.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response YandexGPTResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Result.Alternatives) == 0 {
		return "", fmt.Errorf("no alternatives in response")
	}

	return response.Result.Alternatives[0].Message.Text, nil
}
