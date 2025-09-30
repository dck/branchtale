package pr

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/deck/branchtale/internal/config"
	"github.com/deck/branchtale/internal/git"
	"github.com/fatih/color"
)

// WorkflowHandler handles branch-specific workflows
type WorkflowHandler struct {
	config *config.Config
}

func NewWorkflowHandler(cfg *config.Config) *WorkflowHandler {
	return &WorkflowHandler{config: cfg}
}

func (w *WorkflowHandler) HandleMainBranch(ctx context.Context, gitRepo *git.Repository, generator ContentGenerator, repoInfo *git.RepoInfo, greenFunc func(...interface{}) string) error {
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
	if w.config.UseAI {
		generatedName, err := generator.GenerateBranchName(ctx, diffInfo.Diff)
		if err != nil {
			fmt.Printf("Failed to generate branch name: %v\n", err)
			branchName, err = promptForBranchName()
			if err != nil {
				return err
			}
		} else if generatedName == "" {
			branchName, err = promptForBranchName()
			if err != nil {
				return err
			}
		} else {
			branchName = generatedName
			fmt.Printf("Generated branch name: %s\n", greenFunc(branchName))
		}
	} else {
		branchName, err = promptForBranchName()
		if err != nil {
			return err
		}
	}

	fmt.Printf("Create branch '%s'? (y/N): ", branchName)
	if promptYesNo() {
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

func (w *WorkflowHandler) HandleFeatureBranch(ctx context.Context, repoInfo *git.RepoInfo, greenFunc func(...interface{}) string) error {
	fmt.Printf("You are already on a feature branch: %s\n", greenFunc(repoInfo.CurrentBranch))
	fmt.Println("\nSuggestion: Push this branch to create a pull request.")
	fmt.Printf("Run: git push -u origin %s\n", repoInfo.CurrentBranch)
	return nil
}

func promptForBranchName() (string, error) {
	fmt.Print("Enter branch name: ")
	reader := bufio.NewReader(os.Stdin)
	branchName, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read branch name: %w", err)
	}
	return strings.TrimSpace(branchName), nil
}

func promptYesNo() bool {
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
