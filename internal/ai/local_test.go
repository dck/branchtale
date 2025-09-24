package ai

import (
	"context"
	"testing"
)

func TestLocal_GeneratePRTitle(t *testing.T) {
	l := NewLocal()
	result, err := l.GeneratePRTitle(context.Background(), "some diff")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestLocal_GeneratePRDescription(t *testing.T) {
	l := NewLocal()
	result, err := l.GeneratePRDescription(context.Background(), "some diff")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestLocal_GenerateBranchName(t *testing.T) {
	l := NewLocal()
	result, err := l.GenerateBranchName(context.Background(), "some diff")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}
