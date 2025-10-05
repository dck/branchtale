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
	"github.com/deck/branchtale/internal/vcs"
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
	gitRepo, generator, _, err := s.initializeServices()
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
	var diffInfo *git.DiffInfo

	if repoInfo.IsOnMain {
		diffInfo, err = gitRepo.GetDiffBetweenBranches(ctx, "origin", repoInfo.MainBranch, repoInfo.MainBranch)
		if err != nil {
			return fmt.Errorf("failed to get local commits ahead of origin: %w", err)
		}

		if len(diffInfo.Commits) == 0 {
			return fmt.Errorf("no local commits found ahead of origin")
		}
		fmt.Printf("Found %s local commit(s) ahead of origin:\n", color.New(color.Bold).Sprintf("%d", len(diffInfo.Commits)))
		for i, commit := range diffInfo.Commits {
			fmt.Printf("  %d. %s - %s\n", i+1, color.YellowString(commit.Hash.String()[:8]), strings.TrimSpace(commit.Message))
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

		if err := gitRepo.CreateBranch(ctx, branchName); err != nil {
			return fmt.Errorf("failed to create branch: %w", err)
		}

		if err := gitRepo.CheckoutBranch(ctx, branchName); err != nil {
			return fmt.Errorf("failed to checkout branch: %w", err)
		}
		fmt.Printf("Switched to new branch: %s\n", color.GreenString(branchName))
	} else {
		fmt.Println("You are already on a feature branch.")
		branchOnRemote, err := gitRepo.BranchExistsOnRemote(ctx, repoInfo.CurrentBranch, "origin")
		if err != nil {
			return fmt.Errorf("failed to check remote branch: %w", err)
		}

		if !branchOnRemote {
			color.New(color.Bold).Printf("Branch '%s' is not on remote\n", repoInfo.CurrentBranch)
			fmt.Printf("Pushing branch '%s' to origin...\n", repoInfo.CurrentBranch)
			if err := gitRepo.PushBranch(ctx, repoInfo.CurrentBranch, "origin"); err != nil {
				return fmt.Errorf("failed to push branch: %w", err)
			}
			fmt.Println(color.GreenString("Branch pushed to origin successfully"))
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

	description, err := generator.GeneratePRDescription(ctx, diffInfo.Diff)
	if err != nil {
		return fmt.Errorf("failed to generate PR description: %w", err)
	}

	fmt.Printf("%s: %s\n", color.New(color.Bold).Sprintf("Generated PR Title"), color.GreenString(title))
	fmt.Printf("%s:\n%s\n", color.New(color.Bold).Sprintf("Generated PR Description"), description)
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

	gitRepo, err := git.NewRepository(repoPath, s.config.DryRun)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize git repository: %w", err)
	}

	var generator ContentGenerator
	if s.config.UseAI {
		generator = ai.NewYandexGPT(s.config.YandexGPTAPIKey, s.config.YandexFolderID)
	} else {
		generator = ai.NewLocal()
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
