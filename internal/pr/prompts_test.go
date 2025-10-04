package pr

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrompter_YesNo_yes(t *testing.T) {
	input := strings.NewReader("y\n")
	output := &bytes.Buffer{}
	p := NewPrompter(output, input)
	if !p.YesNo("Proceed?") {
		t.Errorf("Expected YesNo to return true for 'y' input")
	}
	if !strings.Contains(output.String(), "Proceed?") {
		t.Errorf("Prompt not written to output")
	}
}

func TestPrompter_YesNo_no(t *testing.T) {
	input := strings.NewReader("n\n")
	output := &bytes.Buffer{}
	p := NewPrompter(output, input)
	if p.YesNo("Continue?") {
		t.Errorf("Expected YesNo to return false for 'n' input")
	}
}

func TestPrompter_Input(t *testing.T) {
	input := strings.NewReader("hello world\n")
	output := &bytes.Buffer{}
	p := NewPrompter(output, input)
	val, err := p.Input("Enter value")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if val != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", val)
	}
	if !strings.Contains(output.String(), "Enter value") {
		t.Errorf("Prompt not written to output")
	}
}

func TestPrompter_InputWithDefault_usesDefault(t *testing.T) {
	input := strings.NewReader("\n")
	output := &bytes.Buffer{}
	p := NewPrompter(output, input)
	val, err := p.InputWithDefault("Name", "default")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if val != "default" {
		t.Errorf("Expected default value, got '%s'", val)
	}
}

func TestPrompter_InputWithDefault_userInput(t *testing.T) {
	input := strings.NewReader("custom\n")
	output := &bytes.Buffer{}
	p := NewPrompter(output, input)
	val, err := p.InputWithDefault("Name", "default")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if val != "custom" {
		t.Errorf("Expected user input, got '%s'", val)
	}
}
