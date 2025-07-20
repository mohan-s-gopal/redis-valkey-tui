package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// HelpView represents the help view
type HelpView struct {
	flex *tview.Flex
	text *tview.TextView
}

// NewHelpView creates a new help view
func NewHelpView() *HelpView {
	view := &HelpView{}
	view.setupUI()
	return view
}

// setupUI initializes the UI components
func (v *HelpView) setupUI() {
	v.text = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false).
		SetTextAlign(tview.AlignLeft)

	v.text.SetBorder(true).
		SetTitle("Help")

	// Set up input capture to pass through global navigation keys
	v.text.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Let all keys pass through to global handler
		return event
	})

	helpText := `[yellow]Navigation[white]
  [yellow]↑/↓[white]............Navigate keys
  [yellow]Enter[white]..........Select key
  [yellow]Esc[white]............Back/Close
  [yellow]Tab[white]............Next view
  [yellow]Shift+Tab[white]......Previous view

[yellow]Views[white]
  [yellow]1[white]..............Keys view
  [yellow]2[white]..............Monitor view
  [yellow]3[white]..............Info view
  [yellow]4[white]..............CLI view
  [yellow]5[white]..............Config view

[yellow]Key Actions[white]
  [yellow]a[white]..............Add key
  [yellow]d[white]..............Delete key
  [yellow]e[white]..............Edit key
  [yellow]r[white]..............Refresh
  [yellow]f[white]..............Filter keys
  [yellow]/[white]..............Search
  [yellow]?[white]..............Show/hide help

[yellow]Global[white]
  [yellow]Ctrl+C[white].........Quit
  [yellow]Ctrl+R[white].........Refresh all
  [yellow]Ctrl+F[white].........Filter mode`

	v.text.SetText(helpText)

	v.flex = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(v.text, 70, 1, true).
		AddItem(nil, 0, 1, false)
}

// GetComponent returns the view's main component
func (v *HelpView) GetComponent() tview.Primitive {
	return v.flex
}
