package pr

import "github.com/fatih/color"

type ColorPrinter struct {
	Green  func(...interface{}) string
	Blue   func(...interface{}) string
	Yellow func(...interface{}) string
	Red    func(...interface{}) string
	Bold   func(...interface{}) string
	Cyan   func(...interface{}) string
}

func NewColorPrinter() *ColorPrinter {
	return &ColorPrinter{
		Green:  color.New(color.FgGreen, color.Bold).SprintFunc(),
		Blue:   color.New(color.FgBlue, color.Bold).SprintFunc(),
		Yellow: color.New(color.FgYellow).SprintFunc(),
		Red:    color.New(color.FgRed, color.Bold).SprintFunc(),
		Bold:   color.New(color.Bold).SprintFunc(),
		Cyan:   color.New(color.FgCyan).SprintFunc(),
	}
}

func (c *ColorPrinter) Success(format string, args ...interface{}) {
	color.Green("✓ "+format, args...)
}

func (c *ColorPrinter) Info(format string, args ...interface{}) {
	color.Cyan(format, args...)
}

func (c *ColorPrinter) Warning(format string, args ...interface{}) {
	color.Yellow("⚠ "+format, args...)
}

func (c *ColorPrinter) Error(format string, args ...interface{}) {
	color.Red("✗ "+format, args...)
}
