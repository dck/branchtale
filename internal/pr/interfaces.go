package pr

import (
	"context"

	"github.com/deck/branchtale/internal/vcs"
)

type ContentGenerator interface {
	GeneratePRTitle(ctx context.Context, diff string) (string, error)
	GeneratePRDescription(ctx context.Context, diff string) (string, error)
	GenerateBranchName(ctx context.Context, diff string) (string, error)
}

type VCSProvider interface {
	CreatePullRequest(ctx context.Context, req *vcs.CreatePullRequestRequest) (*vcs.CreatePullRequestResponse, error)
	MergePullRequest(ctx context.Context, req *vcs.MergePullRequestRequest) (*vcs.MergePullRequestResponse, error)
}
