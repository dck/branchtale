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
)

type Service struct {
	config *config.Config
	colors *ColorPrinter
}

func NewService(cfg *config.Config) *Service {
	return &Service{
		config: cfg,
		colors: NewColorPrinter(),
	}
}

func (s *Service) Run(ctx context.Context) error {
	gitRepo, generator, vcsProvider, err := s.initializeServices()
	if err != nil {
		return err
	}

	if s.config.Verbose {
		s.colors.Success("Services initialized successfully")
	}

	repoInfo, err := gitRepo.GetInfo()
	if err != nil {
		return err
	}

	fmt.Printf("%s: %s\n", s.colors.Bold("Current branch"), s.colors.Green(repoInfo.CurrentBranch))

	prompter := NewPrompter(s.colors)
	preparation := NewPreparationPhase(s.colors, prompter, gitRepo, generator)
	contentPhase := NewContentPhase(s.colors, prompter, gitRepo, generator)
	creator := NewCreator(s.colors, prompter, gitRepo, vcsProvider)

	if repoInfo.IsOnMain {
		branchCreated, err := preparation.PrepareMainBranch(ctx, repoInfo)
		if err != nil {
			return err
		}
		if branchCreated {
			fmt.Println("\n" + s.colors.Green("Branch created! Run this command again to create a pull request."))
		}
		return nil
	}

	ready, err := preparation.PrepareFeatureBranch(ctx, repoInfo)
	if err != nil {
		return err
	}
	if !ready {
		return nil
	}

	content, err := contentPhase.GenerateContent(ctx, repoInfo)
	if err != nil {
		return err
	}
	if content == nil {
		return nil
	}

	return creator.CreatePR(ctx, repoInfo, content)
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
