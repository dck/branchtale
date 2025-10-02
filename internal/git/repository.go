package git

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Repository struct {
	repo *git.Repository
}

type RepoInfo struct {
	CurrentBranch string
	MainBranch    string
	IsOnMain      bool
}

type DiffInfo struct {
	Diff    string
	Commits []*object.Commit
}

func NewRepository(repoPath string) (*Repository, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	return &Repository{repo: repo}, nil
}

func (s *Repository) GetInfo() (*RepoInfo, error) {
	head, err := s.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	currentBranch := head.Name().Short()

	detectedMain := "master"

	branches, err := s.repo.Branches()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}
	defer branches.Close()

	for {
		ref, err := branches.Next()
		if err != nil {
			break
		}
		name := ref.Name().Short()
		if name == "main" || name == "master" {
			detectedMain = name
			break
		}
	}

	isOnMain := currentBranch == detectedMain

	return &RepoInfo{
		CurrentBranch: currentBranch,
		MainBranch:    detectedMain,
		IsOnMain:      isOnMain,
	}, nil
}

func (s *Repository) GetDiffFromMain(ctx context.Context, mainBranch string) (*DiffInfo, error) {
	head, err := s.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	mainRef, err := s.repo.Reference(plumbing.NewBranchReferenceName(mainBranch), true)
	if err != nil {
		return nil, fmt.Errorf("failed to get main branch reference: %w", err)
	}

	headCommit, err := s.repo.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get head commit: %w", err)
	}

	mainCommit, err := s.repo.CommitObject(mainRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get main commit: %w", err)
	}

	headTree, err := headCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get head tree: %w", err)
	}

	mainTree, err := mainCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get main tree: %w", err)
	}

	patch, err := mainTree.Patch(headTree)
	if err != nil {
		return nil, fmt.Errorf("failed to generate patch: %w", err)
	}

	commits, err := s.getCommitsBetween(mainCommit, headCommit)
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	return &DiffInfo{
		Diff:    patch.String(),
		Commits: commits,
	}, nil
}

func (s *Repository) GetLocalCommitsAheadOfOrigin(ctx context.Context, mainBranch string) (*DiffInfo, error) {
	localRef, err := s.repo.Reference(plumbing.NewBranchReferenceName(mainBranch), true)
	if err != nil {
		return nil, fmt.Errorf("failed to get local main branch reference: %w", err)
	}

	originRef, err := s.repo.Reference(plumbing.NewRemoteReferenceName("origin", mainBranch), true)
	if err != nil {
		return nil, fmt.Errorf("failed to get origin main branch reference: %w", err)
	}

	localCommit, err := s.repo.CommitObject(localRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get local commit: %w", err)
	}

	originCommit, err := s.repo.CommitObject(originRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get origin commit: %w", err)
	}

	localTree, err := localCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get local tree: %w", err)
	}

	originTree, err := originCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get origin tree: %w", err)
	}

	patch, err := originTree.Patch(localTree)
	if err != nil {
		return nil, fmt.Errorf("failed to generate patch: %w", err)
	}

	commits, err := s.getCommitsBetween(originCommit, localCommit)
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	return &DiffInfo{
		Diff:    patch.String(),
		Commits: commits,
	}, nil
}

func (s *Repository) CreateBranch(ctx context.Context, branchName string) error {
	head, err := s.repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	refName := plumbing.NewBranchReferenceName(branchName)
	ref := plumbing.NewHashReference(refName, head.Hash())

	err = s.repo.Storer.SetReference(ref)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	return nil
}

func (s *Repository) CheckoutBranch(ctx context.Context, branchName string) error {
	worktree, err := s.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	return nil
}

func (s *Repository) GetRemoteURL(remoteName string) (string, error) {
	remote, err := s.repo.Remote(remoteName)
	if err != nil {
		return "", fmt.Errorf("failed to get remote: %w", err)
	}

	config := remote.Config()
	if len(config.URLs) == 0 {
		return "", fmt.Errorf("no URLs found for remote %s", remoteName)
	}

	return config.URLs[0], nil
}

func (s *Repository) BranchExistsOnRemote(ctx context.Context, branchName, remoteName string) (bool, error) {
	remote, err := s.repo.Remote(remoteName)
	if err != nil {
		return false, fmt.Errorf("failed to get remote: %w", err)
	}

	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to list remote refs: %w", err)
	}

	branchRef := plumbing.NewBranchReferenceName(branchName)
	for _, ref := range refs {
		if ref.Name().Short() == branchRef.Short() {
			return true, nil
		}
	}

	return false, nil
}

func (s *Repository) PushBranch(ctx context.Context, branchName, remoteName string) error {
	remote, err := s.repo.Remote(remoteName)
	if err != nil {
		return fmt.Errorf("failed to get remote: %w", err)
	}

	refSpec := config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branchName, branchName))

	err = remote.PushContext(ctx, &git.PushOptions{
		RefSpecs: []config.RefSpec{refSpec},
	})
	if err != nil {
		return fmt.Errorf("failed to push branch: %w", err)
	}

	return nil
}

func (s *Repository) getCommitsBetween(from, to *object.Commit) ([]*object.Commit, error) {
	var commits []*object.Commit

	iter, err := s.repo.Log(&git.LogOptions{
		From: to.Hash,
	})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	err = iter.ForEach(func(commit *object.Commit) error {
		if commit.Hash == from.Hash {
			return fmt.Errorf("reached base commit")
		}
		commits = append(commits, commit)
		return nil
	})

	if err != nil && !strings.Contains(err.Error(), "reached base commit") {
		return nil, err
	}

	return commits, nil
}
