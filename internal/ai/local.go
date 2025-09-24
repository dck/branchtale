package ai

import (
	"context"
)

type Local struct{}

func NewLocal() *Local {
	return &Local{}
}

func (l *Local) GeneratePRTitle(ctx context.Context, diff string) (string, error) {
	return "", nil
}

func (l *Local) GeneratePRDescription(ctx context.Context, diff string) (string, error) {
	return "", nil
}

func (l *Local) GenerateBranchName(ctx context.Context, diff string) (string, error) {
	return "", nil
}
