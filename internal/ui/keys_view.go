package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"valkys/internal/config"
	"valkys/internal/redis"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// KeysView represents the keys view
type KeysView struct {
	redis  *redis.Client
	config *config.Config

	// Components
	flex      *tview.Flex
	keyList   *tview.List
	keyDetail *tview.TextView
	filter    *tview.InputField

	// State
	keys         []string
	filteredKeys []string
	selectedKey  string
	filterText   string

	// Layout
	filterVisible bool
}

// NewKeysView creates a new keys view
func NewKeysView(redisClient *redis.Client, cfg *config.Config) *KeysView {
	view := &KeysView{
		redis:  redisClient,
		config: cfg,
	}

	view.setupUI()
	view.loadKeys()

	return view
}

// setupUI initializes the UI components
func (v *KeysView) setupUI() {
	// Create key list
	v.keyList = tview.NewList().
		SetHighlightFullLine(true).
		SetSelectedFunc(v.selectKey)

	v.keyList.SetInputCapture(v.handleKeyInput)
	v.keyList.SetBorder(true).
		SetTitle("Keys").
		SetBorderPadding(0, 0, 1, 1)

	// Create key detail view
	v.keyDetail = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)

	v.keyDetail.SetBorder(true).
		SetTitle("Key Details").
		SetBorderPadding(0, 0, 1, 1)

	// Create filter input
	v.filter = tview.NewInputField().
		SetLabel("Filter: ").
		SetFieldWidth(30).
		SetChangedFunc(v.filterChanged).
		SetDoneFunc(v.filterDone)

	v.filter.SetBorder(true).
		SetTitle("Filter").
		SetBorderPadding(0, 0, 1, 1)

	// Create main layout
	v.flex = tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(v.keyList, 0, 1, true).
		AddItem(v.keyDetail, 0, 1, false)

	v.updateLayout()
}

// updateLayout updates the layout based on filter visibility
func (v *KeysView) updateLayout() {
	v.flex.Clear()

	if v.filterVisible {
		leftPanel := tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(v.filter, 3, 0, false).
			AddItem(v.keyList, 0, 1, true)

		v.flex.AddItem(leftPanel, 0, 1, true).
			AddItem(v.keyDetail, 0, 1, false)
	} else {
		v.flex.AddItem(v.keyList, 0, 1, true).
			AddItem(v.keyDetail, 0, 1, false)
	}
}

// GetComponent returns the main component
func (v *KeysView) GetComponent() tview.Primitive {
	return v.flex
}

// loadKeys loads keys from Redis
func (v *KeysView) loadKeys() {
	keys, err := v.redis.GetKeys("*")
	if err != nil {
		v.keyDetail.SetText(fmt.Sprintf("[red]Error loading keys: %s", err))
		return
	}

	sort.Strings(keys)
	v.keys = keys
	v.applyFilter()
}

// applyFilter applies the current filter to the keys
func (v *KeysView) applyFilter() {
	if v.filterText == "" {
		v.filteredKeys = v.keys
	} else {
		v.filteredKeys = []string{}
		for _, key := range v.keys {
			if strings.Contains(strings.ToLower(key), strings.ToLower(v.filterText)) {
				v.filteredKeys = append(v.filteredKeys, key)
			}
		}
	}

	v.updateKeyList()
}

// updateKeyList updates the key list display
func (v *KeysView) updateKeyList() {
	v.keyList.Clear()

	for i, key := range v.filteredKeys {
		// Get key info for display
		info, err := v.redis.GetKeyInfo(key)
		if err != nil {
			v.keyList.AddItem(fmt.Sprintf("%s [red](error)", key), "", rune('0'+i%10), nil)
			continue
		}

		// Format TTL
		ttlStr := ""
		if info.TTL > 0 {
			ttlStr = fmt.Sprintf(" [yellow]TTL:%s", info.TTL.Truncate(time.Second))
		} else if info.TTL == -1 {
			ttlStr = " [green]∞"
		}

		// Format memory usage
		memStr := ""
		if info.MemoryUsage > 0 {
			memStr = fmt.Sprintf(" [blue]%s", formatBytes(info.MemoryUsage))
		}

		displayText := fmt.Sprintf("%s [gray]%s%s%s", key, info.Type, ttlStr, memStr)
		v.keyList.AddItem(displayText, "", rune('0'+i%10), nil)
	}

	// Update title with count
	title := fmt.Sprintf("Keys (%d/%d)", len(v.filteredKeys), len(v.keys))
	if v.filterText != "" {
		title += fmt.Sprintf(" [filter: %s]", v.filterText)
	}
	v.keyList.SetTitle(title)
}

// selectKey handles key selection
func (v *KeysView) selectKey(index int, mainText, secondaryText string, shortcut rune) {
	if index >= len(v.filteredKeys) {
		return
	}

	v.selectedKey = v.filteredKeys[index]
	v.showKeyDetails()
}

// showKeyDetails shows details for the selected key
func (v *KeysView) showKeyDetails() {
	if v.selectedKey == "" {
		v.keyDetail.SetText("No key selected")
		return
	}

	// Get key info
	info, err := v.redis.GetKeyInfo(v.selectedKey)
	if err != nil {
		v.keyDetail.SetText(fmt.Sprintf("[red]Error getting key info: %s", err))
		return
	}

	// Get key value
	value, err := v.redis.GetValue(v.selectedKey)
	if err != nil {
		value = fmt.Sprintf("[red]Error getting value: %s", err)
	}

	// Format details
	details := fmt.Sprintf(`[white]Key:[yellow] %s
[white]Type:[yellow] %s
[white]TTL:[yellow] %s
[white]Memory:[yellow] %s

[white]Value:
[green]%s`,
		v.selectedKey,
		info.Type,
		formatTTL(info.TTL),
		formatBytes(info.MemoryUsage),
		value)

	v.keyDetail.SetText(details)
}

// handleKeyInput handles key input for the keys view
func (v *KeysView) handleKeyInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case '/':
		v.showFilter()
		return nil
	case 'r':
		v.Refresh()
		return nil
	case 'd':
		v.deleteKey()
		return nil
	case 'e':
		v.editKey()
		return nil
	case 't':
		v.setTTL()
		return nil
	}

	switch event.Key() {
	case tcell.KeyEscape:
		if v.filterVisible {
			v.hideFilter()
			return nil
		}
	}

	return event
}

// showFilter shows the filter input
func (v *KeysView) showFilter() {
	v.filterVisible = true
	v.updateLayout()
}

// hideFilter hides the filter input
func (v *KeysView) hideFilter() {
	v.filterVisible = false
	v.updateLayout()
}

// filterChanged handles filter text changes
func (v *KeysView) filterChanged(text string) {
	v.filterText = text
	v.applyFilter()
}

// filterDone handles filter input completion
func (v *KeysView) filterDone(key tcell.Key) {
	if key == tcell.KeyEscape {
		v.hideFilter()
	}
}

// deleteKey deletes the selected key
func (v *KeysView) deleteKey() {
	if v.selectedKey == "" {
		return
	}

	// For now, just delete without confirmation
	// TODO: Implement confirmation dialog through parent app
	err := v.redis.DeleteKey(v.selectedKey)
	if err != nil {
		v.keyDetail.SetText(fmt.Sprintf("[red]Error deleting key: %s", err))
	} else {
		v.loadKeys()
		v.keyDetail.SetText(fmt.Sprintf("[green]Key '%s' deleted", v.selectedKey))
		v.selectedKey = ""
	}
}

// editKey edits the selected key
func (v *KeysView) editKey() {
	if v.selectedKey == "" {
		return
	}

	// Get current value
	currentValue, err := v.redis.GetValue(v.selectedKey)
	if err != nil {
		v.keyDetail.SetText(fmt.Sprintf("[red]Error getting value: %s", err))
		return
	}

	// This would need to be handled by the parent app with a modal
	// For now, just show the current value
	v.keyDetail.SetText(fmt.Sprintf("[yellow]Edit mode not implemented yet\nCurrent value: %s", currentValue))
}

// setTTL sets TTL for the selected key
func (v *KeysView) setTTL() {
	if v.selectedKey == "" {
		return
	}

	// This would need to be handled by the parent app with a modal
	// For now, just show current TTL
	info, err := v.redis.GetKeyInfo(v.selectedKey)
	if err != nil {
		v.keyDetail.SetText(fmt.Sprintf("[red]Error getting key info: %s", err))
		return
	}

	v.keyDetail.SetText(fmt.Sprintf("[yellow]TTL edit mode not implemented yet\nCurrent TTL: %s", formatTTL(info.TTL)))
}

// Refresh refreshes the keys view
func (v *KeysView) Refresh() {
	v.loadKeys()
	if v.selectedKey != "" {
		v.showKeyDetails()
	}
}

// formatBytes formats bytes in human readable format
func formatBytes(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatTTL formats TTL in human readable format
func formatTTL(ttl time.Duration) string {
	if ttl == -1 {
		return "∞ (no expiration)"
	}
	if ttl == -2 {
		return "Key does not exist"
	}
	if ttl == 0 {
		return "Expired"
	}

	if ttl < time.Minute {
		return fmt.Sprintf("%ds", int(ttl.Seconds()))
	}
	if ttl < time.Hour {
		return fmt.Sprintf("%dm %ds", int(ttl.Minutes()), int(ttl.Seconds())%60)
	}
	if ttl < 24*time.Hour {
		return fmt.Sprintf("%dh %dm", int(ttl.Hours()), int(ttl.Minutes())%60)
	}

	days := int(ttl.Hours()) / 24
	hours := int(ttl.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}
