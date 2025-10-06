package pr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/deck/branchtale/internal/ai"
	"github.com/deck/branchtale/internal/config"
	"github.com/deck/branchtale/internal/git"
	"github.com/fatih/color"
)

type Service struct {
	config *config.Config
}

func NewService(cfg *config.Config) *Service {
	return &Service{
		config: cfg,
	}
}

func (s *Service) Run(ctx context.Context) error {
	gitRepo, generator, err := s.initializeServices()
	if err != nil {
		return err
	}

	if s.config.Verbose {
		fmt.Println("Services initialized successfully")
	}

	repoInfo, err := gitRepo.GetInfo()
	if err != nil {
		return err
	}

	fmt.Printf("Current branch: %s\n", color.GreenString(repoInfo.CurrentBranch))
	r := &Requirements{
		BaseBranch: repoInfo.MainBranch,
	}

	var diffInfo *git.DiffInfo
	if repoInfo.IsOnMain {
		diffInfo, err = gitRepo.GetDiffBetweenBranches(ctx, "origin", repoInfo.MainBranch, repoInfo.MainBranch)
		if err != nil {
			return fmt.Errorf("failed to get local commits ahead of origin: %w", err)
		}

		if len(diffInfo.Commits) == 0 {
			color.Blue("Your branch is up to date with origin/%s. Nothing to do.\n", repoInfo.MainBranch)
			return nil
		}

		if s.config.Verbose {
			fmt.Printf("Found %s local commit(s) ahead of origin:\n", color.New(color.Bold).Sprintf("%d", len(diffInfo.Commits)))
			for i, commit := range diffInfo.Commits {
				fmt.Printf("  %d. %s - %s\n", i+1, color.YellowString(commit.Hash.String()[:8]), strings.TrimSpace(commit.Message))
			}
		}

		fmt.Println("Generating a feature branch name for these changes...")
		branchName, err := generator.GenerateBranchName(ctx, diffInfo.Diff)
		if err != nil {
			return fmt.Errorf("failed to generate branch name: %w", err)
		}
		if s.config.BranchPrefix != "" {
			branchName = s.config.BranchPrefix + branchName
		}
		fmt.Printf("Suggested branch: %s\n", color.GreenString(branchName))
		r.CreateBranch = true
		r.BranchName = branchName
	} else {
		if s.config.Verbose {
			fmt.Println("You are already on a feature branch.")
		}

		branchOnRemote, err := gitRepo.BranchExistsOnRemote(ctx, repoInfo.CurrentBranch, "origin")
		if err != nil {
			return fmt.Errorf("failed to check remote branch: %w", err)
		}
		if !branchOnRemote {
			r.PushBranch = true
			fmt.Printf("Branch %s does not exist on remote. It will be pushed.\n", color.YellowString(repoInfo.CurrentBranch))
		}

		diffInfo, err = gitRepo.GetDiffBetweenBranches(ctx, "origin", repoInfo.MainBranch, repoInfo.CurrentBranch)
		if err != nil {
			return fmt.Errorf("failed to get diff from origin/master: %w", err)
		}
	}

	title, err := generator.GeneratePRTitle(ctx, diffInfo.Diff)
	if err != nil {
		return fmt.Errorf("failed to generate PR title: %w", err)
	}
	r.PullRequestTitle = title

	description, err := generator.GeneratePRDescription(ctx, diffInfo.Diff)
	if err != nil {
		return fmt.Errorf("failed to generate PR description: %w", err)
	}
	r.PullRequestDescription = description
	return Execute(ctx, r, gitRepo, s.config)
}

func (s *Service) initializeServices() (*git.Repository, ContentGenerator, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	repoPath, err := findGitRepo(cwd)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find git repository: %w", err)
	}

	gitRepo, err := git.NewRepository(repoPath, s.config.DryRun)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize git repository: %w", err)
	}

	var generator ContentGenerator
	if s.config.UseAI {
		generator = ai.NewYandexGPT(s.config.YandexGPTAPIKey, s.config.YandexFolderID)
	} else {
		generator = ai.NewLocal()
	}

	return gitRepo, generator, nil
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
