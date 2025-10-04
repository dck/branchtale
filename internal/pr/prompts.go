package pr

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
)

type Prompter struct {
	writer io.Writer
	reader *bufio.Reader
}

func NewPrompter(output io.Writer, input io.Reader) *Prompter {
	return &Prompter{
		writer: output,
		reader: bufio.NewReader(input),
	}
}

func (p *Prompter) YesNo(prompt string) bool {
	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Fprintf(p.writer, "%s (y/N): ", cyan(prompt))
	response, err := p.reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func (p *Prompter) Input(prompt string) (string, error) {
	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Fprintf(p.writer, "%s: ", cyan(prompt))
	input, err := p.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}
	return strings.TrimSpace(input), nil
}

func (p *Prompter) InputWithDefault(prompt, defaultValue string) (string, error) {
	cyan := color.New(color.FgCyan).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()
	if defaultValue != "" {
		fmt.Fprintf(p.writer, "%s [%s]: ", cyan(prompt), bold(defaultValue))
	} else {
		fmt.Fprintf(p.writer, "%s: ", cyan(prompt))
	}

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "" && defaultValue != "" {
		return defaultValue, nil
	}

	return input, nil
}
