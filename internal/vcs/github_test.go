package vcs

import (
	"testing"
)

func TestParseGitHubURL(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		expectedOwner string
		expectedRepo  string
		expectError   bool
	}{
		{
			name:          "HTTPS URL",
			url:           "https://github.com/owner/repo.git",
			expectedOwner: "owner",
			expectedRepo:  "repo",
			expectError:   false,
		},
		{
			name:          "HTTPS URL without .git",
			url:           "https://github.com/owner/repo",
			expectedOwner: "owner",
			expectedRepo:  "repo",
			expectError:   false,
		},
		{
			name:          "SSH URL",
			url:           "git@github.com:owner/repo.git",
			expectedOwner: "owner",
			expectedRepo:  "repo",
			expectError:   false,
		},
		{
			name:        "invalid URL",
			url:         "not-a-url",
			expectError: true,
		},
		{
			name:        "non-GitHub URL",
			url:         "https://gitlab.com/owner/repo.git",
			expectError: true,
		},
		{
			name:        "invalid GitHub URL format",
			url:         "https://github.com/owner",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := ParseGitHubURL(tt.url)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for URL %s, but got none", tt.url)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error for URL %s: %v", tt.url, err)
				return
			}

			if owner != tt.expectedOwner {
				t.Errorf("expected owner '%s', got '%s'", tt.expectedOwner, owner)
			}

			if repo != tt.expectedRepo {
				t.Errorf("expected repo '%s', got '%s'", tt.expectedRepo, repo)
			}
		})
	}
}
