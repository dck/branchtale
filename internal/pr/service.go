package pr

import (
	"bufio"
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
	config    *config.Config
	greenFunc func(...interface{}) string
}

func NewService(cfg *config.Config) *Service {
	return &Service{config: cfg, greenFunc: color.New(color.FgGreen, color.Bold).SprintFunc()}
}

func (s *Service) Run(ctx context.Context) error {
	gitRepo, generator, _, err := s.initializeServices()
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

	if repoInfo.IsOnMain {
		return s.handleMainBranch(ctx, gitRepo, generator, repoInfo)
	} else {
		return s.handleFeatureBranch(ctx, repoInfo)
	}
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

func (s *Service) handleMainBranch(ctx context.Context, gitRepo *git.Repository, generator ContentGenerator, repoInfo *git.RepoInfo) error {
	fmt.Println("You are on the main branch.")

	diffInfo, err := gitRepo.GetLocalCommitsAheadOfOrigin(ctx, repoInfo.MainBranch)
	if err != nil {
		return fmt.Errorf("failed to get local commits ahead of origin: %w", err)
	}

	if len(diffInfo.Commits) == 0 {
		fmt.Println("No local commits found ahead of origin. Your branch is up to date.")
		return nil
	}

	fmt.Printf("\nFound %d local commit(s) ahead of origin:\n", len(diffInfo.Commits))
	for i, commit := range diffInfo.Commits {
		fmt.Printf("  %d. %s - %s\n", i+1, commit.Hash.String()[:8], strings.TrimSpace(commit.Message))
	}

	fmt.Println("\nSuggesting to create a new branch for these changes.")

	var branchName string
	if s.config.UseAI {
		generatedName, err := generator.GenerateBranchName(ctx, diffInfo.Diff)
		if err != nil {
			fmt.Printf("Failed to generate branch name: %v\n", err)
			branchName, err = s.promptForBranchName()
			if err != nil {
				return err
			}
		} else if generatedName == "" {
			branchName, err = s.promptForBranchName()
			if err != nil {
				return err
			}
		} else {
			branchName = generatedName
			fmt.Printf("Generated branch name: %s\n", s.greenFunc(branchName))
		}
	} else {
		branchName, err = s.promptForBranchName()
		if err != nil {
			return err
		}
	}

	fmt.Printf("Create branch '%s'? (y/N): ", branchName)
	if s.promptYesNo() {
		err = gitRepo.CreateBranch(ctx, branchName)
		if err != nil {
			return fmt.Errorf("failed to create branch: %w", err)
		}

		err = gitRepo.CheckoutBranch(ctx, branchName)
		if err != nil {
			return fmt.Errorf("failed to checkout branch: %w", err)
		}

		color.Green("✓ Branch '%s' created and checked out successfully", branchName)
	}

	return nil
}

func (s *Service) handleFeatureBranch(ctx context.Context, repoInfo *git.RepoInfo) error {
	fmt.Printf("You are already on a feature branch: %s\n", s.greenFunc(repoInfo.CurrentBranch))
	fmt.Println("\nSuggestion: Push this branch to create a pull request.")
	fmt.Printf("Run: git push -u origin %s\n", repoInfo.CurrentBranch)
	return nil
}

func (s *Service) promptForBranchName() (string, error) {
	fmt.Print("Enter branch name: ")
	reader := bufio.NewReader(os.Stdin)
	branchName, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read branch name: %w", err)
	}
	return strings.TrimSpace(branchName), nil
}

func (s *Service) promptYesNo() bool {
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
