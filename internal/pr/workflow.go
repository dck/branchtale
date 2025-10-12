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
	MergePullRequest       bool
	PullRequestTitle       string
	PullRequestDescription string
	PullRequestTags        []string
}

func Execute(ctx context.Context, reqs *Requirements, gitRepo *git.Repository, cfg *config.Config) error {
	vcsProvider := vcs.NewGitHubProvider(cfg.GitHubToken)

	if cfg.DryRun {
		return DryExecute(ctx, reqs, gitRepo, cfg)
	}

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
		pr := &vcs.CreatePullRequestRequest{
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

		if reqs.MergePullRequest {
			mergeReq := &vcs.MergePullRequestRequest{
				Owner:       owner,
				Repo:        repo,
				Number:      response.Number,
				MergeMethod: "merge",
			}
			mergeResp, err := vcsProvider.MergePullRequest(ctx, mergeReq)

			if err != nil {
				return err
			}
			if mergeResp.Merged {
				fmt.Printf("Pull Request #%d merged successfully with SHA: %s\n", response.Number, color.GreenString(mergeResp.SHA))
			} else {
				fmt.Printf("Pull Request #%d was not merged. Message: %s\n", response.Number, color.YellowString(mergeResp.Message))
			}
		}
	}

	return nil
}

func DryExecute(ctx context.Context, reqs *Requirements, gitRepo *git.Repository, cfg *config.Config) error {
	fmt.Println("Dry run mode enabled. The following actions would be performed:")
	if reqs.CreateBranch {
		fmt.Printf("- Create branch: %s\n", color.GreenString(reqs.BranchName))
		fmt.Printf("- Checkout branch: %s\n", color.GreenString(reqs.BranchName))
	}
	if reqs.PushBranch {
		fmt.Printf("- Push branch: %s to remote 'origin'\n", color.GreenString(reqs.BranchName))
	}
	if reqs.CreatePullRequest {
		remoteUrl, err := gitRepo.GetRemoteUrl(ctx, "origin")
		if err != nil {
			return err
		}

		owner, repo, err := vcs.ParseGitHubURL(remoteUrl)
		if err != nil {
			return err
		}

		fmt.Printf("- Create Pull Request in repository %s/%s\n", color.GreenString(owner), color.GreenString(repo))
		fmt.Printf("  - Title: %s\n", color.GreenString(reqs.PullRequestTitle))
		fmt.Printf("  - Description: %s\n", color.GreenString(reqs.PullRequestDescription))
		fmt.Printf("  - Head Branch: %s\n", color.GreenString(reqs.BranchName))
		fmt.Printf("  - Base Branch: %s\n", color.GreenString(reqs.BaseBranch))
	}
	return nil
}
