package pr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/deck/branchtale/internal/ai"
	"github.com/deck/branchtale/internal/config"
	"github.com/deck/branchtale/internal/git"
	"github.com/deck/branchtale/internal/vcs"
	"github.com/fatih/color"
)

type ContentGenerator interface {
	GeneratePRTitle(ctx context.Context, diff string) (string, error)
	GeneratePRDescription(ctx context.Context, diff string) (string, error)
	GenerateBranchName(ctx context.Context, diff string) (string, error)
}

type VCSProvider interface {
	CreatePullRequest(ctx context.Context, req *vcs.PullRequestRequest) (*vcs.PullRequestResponse, error)
}

type Service struct {
	config    *config.Config
	greenFunc func(...interface{}) string
}

func NewService(cfg *config.Config) *Service {
	return &Service{config: cfg, greenFunc: color.New(color.FgGreen, color.Bold).SprintFunc()}
}

func (s *Service) Run(ctx context.Context) error {
	gitRepo, _, _, err := s.initializeServices()
	if err != nil {
		return err
	}

	if s.config.Verbose {
		color.Green("✓ Services initialized successfully")
	}
	repoInfo, err := gitRepo.GetInfo()
	if err != nil {
		return err
	}
	fmt.Printf("Current branch: %s\n", s.greenFunc(repoInfo.CurrentBranch))

	return nil
}

func (s *Service) initializeServices() (*git.Repository, ContentGenerator, VCSProvider, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	repoPath, err := findGitRepo(cwd)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to find git repository: %w", err)
	}

	gitRepo, err := git.NewRepository(repoPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize git repository: %w", err)
	}

	var generator ContentGenerator
	if s.config.UseAI {
		generator = ai.NewYandexGPT(s.config.YandexGPTAPIKey, s.config.YandexFolderID)
	} else {
		if s.config.Interactive {
			generator = ai.NewLocal()
		} else {
			return nil, nil, nil, fmt.Errorf("non-GPT mode requires interactive mode")
		}
	}

	vcsProvider := vcs.NewGitHubProvider(s.config.GitHubToken)

	return gitRepo, generator, vcsProvider, nil
}

func findGitRepo(startPath string) (string, error) {
	currentPath := startPath
	for {
		gitPath := filepath.Join(currentPath, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return currentPath, nil
		}

		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			break
		}
		currentPath = parent
	}

	return "", fmt.Errorf("not a git repository (or any of the parent directories)")
}
