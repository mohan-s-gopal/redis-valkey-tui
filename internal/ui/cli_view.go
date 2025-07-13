package ui

import (
	"fmt"
	"strings"
	"valkys/internal/redis"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// CLIView represents the CLI view
type CLIView struct {
	redis *redis.Client

	// Components
	flex   *tview.Flex
	input  *tview.InputField
	output *tview.TextView

	// State
	history      []string
	historyIndex int
}

// NewCLIView creates a new CLI view
func NewCLIView(redisClient *redis.Client) *CLIView {
	view := &CLIView{
		redis:        redisClient,
		history:      []string{},
		historyIndex: 0,
	}

	view.setupUI()
	view.showWelcome()

	return view
}

// setupUI initializes the UI components
func (v *CLIView) setupUI() {
	// Create input field
	v.input = tview.NewInputField().
		SetLabel("redis> ").
		SetFieldWidth(0).
		SetDoneFunc(v.handleCommand)

	v.input.SetInputCapture(v.handleInput)

	v.input.SetBorder(true).
		SetTitle("Command Input").
		SetBorderPadding(0, 0, 1, 1)

	// Create output view
	v.output = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)

	v.output.SetBorder(true).
		SetTitle("Output").
		SetBorderPadding(0, 0, 1, 1)

	// Create main layout
	v.flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(v.output, 0, 1, false).
		AddItem(v.input, 3, 0, true)
}

// GetComponent returns the main component
func (v *CLIView) GetComponent() tview.Primitive {
	return v.flex
}

// handleInput handles input for navigation and special keys
func (v *CLIView) handleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyUp:
		v.navigateHistory(-1)
		return nil
	case tcell.KeyDown:
		v.navigateHistory(1)
		return nil
	case tcell.KeyCtrlL:
		v.clearOutput()
		return nil
	}

	return event
}

// handleCommand handles command execution
func (v *CLIView) handleCommand(key tcell.Key) {
	if key != tcell.KeyEnter {
		return
	}

	command := strings.TrimSpace(v.input.GetText())
	if command == "" {
		return
	}

	// Add to history
	v.history = append(v.history, command)
	v.historyIndex = len(v.history)

	// Clear input
	v.input.SetText("")

	// Execute command
	v.executeCommand(command)
}

// executeCommand executes a Redis command
func (v *CLIView) executeCommand(command string) {
	// Add command to output
	v.appendOutput(fmt.Sprintf("[green]redis> %s[white]", command))

	// Parse command
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}

	// Execute command
	result, err := v.redis.ExecuteCommand(parts[0], interfaceSlice(parts[1:])...)
	if err != nil {
		v.appendOutput(fmt.Sprintf("[red]Error: %s[white]", err))
		return
	}

	// Format and display result
	v.appendOutput(v.formatResult(result))
}

// formatResult formats a Redis command result
func (v *CLIView) formatResult(result interface{}) string {
	switch v := result.(type) {
	case string:
		return fmt.Sprintf("[yellow]%s[white]", v)
	case int64:
		return fmt.Sprintf("[cyan]%d[white]", v)
	case []interface{}:
		if len(v) == 0 {
			return "[gray](empty list)[white]"
		}

		formatted := "[cyan]"
		for i, item := range v {
			if i > 0 {
				formatted += "\n"
			}
			formatted += fmt.Sprintf("%d) %v", i+1, item)
		}
		formatted += "[white]"
		return formatted
	case nil:
		return "[gray](nil)[white]"
	default:
		return fmt.Sprintf("[yellow]%v[white]", v)
	}
}

// appendOutput appends text to the output
func (v *CLIView) appendOutput(text string) {
	currentText := v.output.GetText(false)
	if currentText != "" {
		currentText += "\n"
	}
	currentText += text

	v.output.SetText(currentText)
	v.output.ScrollToEnd()
}

// navigateHistory navigates through command history
func (v *CLIView) navigateHistory(direction int) {
	if len(v.history) == 0 {
		return
	}

	v.historyIndex += direction

	if v.historyIndex < 0 {
		v.historyIndex = 0
	} else if v.historyIndex >= len(v.history) {
		v.historyIndex = len(v.history)
		v.input.SetText("")
		return
	}

	v.input.SetText(v.history[v.historyIndex])
}

// clearOutput clears the output
func (v *CLIView) clearOutput() {
	v.output.SetText("")
	v.showWelcome()
}

// showWelcome shows the welcome message
func (v *CLIView) showWelcome() {
	welcome := `[yellow]Redis/Valkey CLI Interface[white]

Welcome to the Redis CLI! You can execute any Redis command here.

Examples:
  SET mykey "hello world"
  GET mykey
  KEYS *
  INFO
  PING

Navigation:
  ↑/↓       Navigate command history
  Ctrl+L    Clear screen
  Enter     Execute command

Type your commands below:
`
	v.output.SetText(welcome)
}

// Refresh refreshes the CLI view
func (v *CLIView) Refresh() {
	// Nothing to refresh in CLI view
}

// interfaceSlice converts string slice to interface slice
func interfaceSlice(slice []string) []interface{} {
	result := make([]interface{}, len(slice))
	for i, v := range slice {
		result[i] = v
	}
	return result
}
