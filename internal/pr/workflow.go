package pr

import (
	"context"
	"fmt"

	"github.com/deck/branchtale/internal/config"
	"github.com/deck/branchtale/internal/git"
	"github.com/deck/branchtale/internal/vcs"
	"github.com/fatih/color"
)

type Requirements struct {
	CreateBranch           bool
	PushBranch             bool
	BranchName             string
	BaseBranch             string
	CreatePullRequest      bool
	PullRequestTitle       string
	PullRequestDescription string
	PullRequestTags        []string
}

func Execute(ctx context.Context, reqs *Requirements, gitRepo *git.Repository, cfg *config.Config) error {
	vcsProvider := vcs.NewGitHubProvider(cfg.GitHubToken)

	if reqs.CreateBranch {
		if err := gitRepo.CreateBranch(ctx, reqs.BranchName); err != nil {
			return err
		}
		fmt.Printf("Created branch: %s\n", color.GreenString(reqs.BranchName))

		if err := gitRepo.CheckoutBranch(ctx, reqs.BranchName); err != nil {
			return err
		}
		fmt.Printf("Checked out branch: %s\n", color.GreenString(reqs.BranchName))
	}

	if reqs.PushBranch {
		if err := gitRepo.PushBranch(ctx, reqs.BranchName, "origin"); err != nil {
			return err
		}
		fmt.Printf("Pushed branch: %s\n", color.GreenString(reqs.BranchName))
	}

	remoteUrl, err := gitRepo.GetRemoteUrl(ctx, "origin")
	if err != nil {
		return err
	}

	owner, repo, err := vcs.ParseGitHubURL(remoteUrl)
	if err != nil {
		return err
	}

	if reqs.CreatePullRequest {
		pr := &vcs.PullRequestRequest{
			Owner:       owner,
			Repo:        repo,
			Title:       reqs.PullRequestTitle,
			Description: reqs.PullRequestDescription,
			HeadBranch:  reqs.BranchName,
			BaseBranch:  reqs.BaseBranch,
		}
		if cfg.Verbose {
			fmt.Printf("Creating Pull Request with this payload: %+v\n", pr)
		}

		response, err := vcsProvider.CreatePullRequest(ctx, pr)
		if err != nil {
			return err
		}
		fmt.Printf("Pull Request created: %s\n", color.GreenString(response.URL))
	}

	return nil
}
