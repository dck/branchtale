package vcs

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/go-github/v74/github"
	"golang.org/x/oauth2"
)

type Provider interface {
	CreatePullRequest(ctx context.Context, req *PullRequestRequest) (*PullRequestResponse, error)
}

type PullRequestRequest struct {
	Owner       string
	Repo        string
	Title       string
	Description string
	HeadBranch  string
	BaseBranch  string
}

type PullRequestResponse struct {
	URL    string
	Number int
}

type GitHubProvider struct {
	client *github.Client
}

func NewGitHubProvider(token string) *GitHubProvider {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	return &GitHubProvider{client: client}
}

func (g *GitHubProvider) CreatePullRequest(ctx context.Context, req *PullRequestRequest) (*PullRequestResponse, error) {
	pr := &github.NewPullRequest{
		Title: &req.Title,
		Head:  &req.HeadBranch,
		Base:  &req.BaseBranch,
		Body:  &req.Description,
	}

	result, _, err := g.client.PullRequests.Create(ctx, req.Owner, req.Repo, pr)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	return &PullRequestResponse{
		URL:    result.GetHTMLURL(),
		Number: result.GetNumber(),
	}, nil
}

func ParseGitHubURL(remoteURL string) (owner, repo string, err error) {
	sshRegex := regexp.MustCompile(`git@github\.com:(.+)/(.+)\.git`)
	if matches := sshRegex.FindStringSubmatch(remoteURL); len(matches) == 3 {
		return matches[1], matches[2], nil
	}

	u, err := url.Parse(remoteURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse URL: %w", err)
	}

	if u.Host != "github.com" {
		return "", "", fmt.Errorf("not a GitHub URL: %s", remoteURL)
	}

	pathParts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(pathParts) != 2 {
		return "", "", fmt.Errorf("invalid GitHub URL format: %s", remoteURL)
	}

	repo = strings.TrimSuffix(pathParts[1], ".git")
	return pathParts[0], repo, nil
}
