package ui

import (
	"fmt"
	"strings"

	"github.com/mohan-s-gopal/redis-valkey-tui/internal/config"
	"github.com/mohan-s-gopal/redis-valkey-tui/internal/logger"
	"github.com/mohan-s-gopal/redis-valkey-tui/internal/redis"

	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// KeysView represents the keys view
type KeysView struct {
	redis  *redis.Client
	config *config.Config

	// Components
	flex          *tview.Flex
	table         *tview.Table
	keyDetail     *tview.TextView
	filter        *tview.InputField
	commandInput  *tview.InputField
	commandOutput *tview.TextView

	// State
	keys         []*redis.KeyInfo
	filteredKeys []*redis.KeyInfo
	selectedKey  string
	filterText   string
	focusIndex   int // 0=table, 1=filter, 2=command

	// Callbacks
	onFocusChange func(component tview.Primitive)

	// Layout
	filterVisible bool
}

// NewKeysView creates a new keys view
func NewKeysView(redisClient *redis.Client, cfg *config.Config) *KeysView {
	logger.Logger.Println("[KeysView] Initializing NewKeysView...")
	view := &KeysView{
		redis:      redisClient,
		config:     cfg,
		focusIndex: 0, // Start with table focused
	}

	view.setupUI()
	logger.Logger.Println("[KeysView] Scheduling async loadKeys()...")
	// Show loading message
	view.table.Clear()
	view.table.SetCell(0, 0, tview.NewTableCell("Loading keys...").SetTextColor(tcell.ColorYellow))
	// Load keys asynchronously
	go func() {
		view.loadKeys()
	}()
	logger.Logger.Println("[KeysView] Async loadKeys() scheduled.")

	return view
} // setupUI initializes the UI components
func (v *KeysView) setupUI() {
	// Create table for keys
	v.table = tview.NewTable().
		SetBorders(false).
		SetFixed(1, 0).
		SetSelectable(true, false)

	v.table.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseLeftClick {
			// Get the clicked position
			_, y := event.Position()
			// Convert screen coordinates to table coordinates
			_, tableY, _, _ := v.table.GetInnerRect()
			relY := y - tableY

			// Calculate which row was clicked (accounting for header row)
			clickedRow := relY + 1 // +1 because header is row 0

			if clickedRow > 0 && clickedRow <= len(v.getDisplayKeys()) {
				// Set the selection to the clicked row
				v.table.Select(clickedRow, 0)

				keyInfo := v.getDisplayKeys()[clickedRow-1]
				if keyInfo != nil {
					v.selectedKey = keyInfo.Name
					v.showKeyDetails(keyInfo.Name)
				}
			}
		}
		return action, event
	})

	// Set up headers
	headers := []string{"Type", "Key", "TTL", "Size", "Encoding"}
	for i, header := range headers {
		v.table.SetCell(0, i,
			tview.NewTableCell(header).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignLeft).
				SetSelectable(false))
	}

	// Handle selection changes (navigation with arrow keys)
	v.table.SetSelectionChangedFunc(func(row, col int) {
		if row > 0 && row <= len(v.getDisplayKeys()) {
			keyInfo := v.getDisplayKeys()[row-1]
			if keyInfo != nil {
				v.selectedKey = keyInfo.Name
				v.showKeyDetails(keyInfo.Name)
			}
		}
	})

	// Handle enter key on selection
	v.table.SetSelectedFunc(func(row, col int) {
		if row > 0 && row <= len(v.getDisplayKeys()) {
			keyInfo := v.getDisplayKeys()[row-1]
			if keyInfo != nil {
				v.selectedKey = keyInfo.Name
				v.showKeyDetails(keyInfo.Name)
			}
		}
	})

	// Key detail view
	v.keyDetail = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true)
	v.keyDetail.SetBorder(true).
		SetTitle("Key Details")

	// Filter input
	v.filter = tview.NewInputField().
		SetLabel("Filter: ").
		SetFieldWidth(0).
		SetChangedFunc(func(text string) {
			// Apply filter as user types
			v.applyFilter(text)
		}).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEscape {
				v.setFocus(0) // Return to table on escape
			} else {
				v.applyFilter(v.filter.GetText())
			}
		})

	v.filter.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Handle key events in filter input
		switch event.Key() {
		case tcell.KeyCtrlL:
			// Clear filter input with Ctrl+L
			v.filter.SetText("")
			v.applyFilter("")
			return nil
		case tcell.KeyCtrlC:
			// Clear filter and return to table with Ctrl+C
			v.filter.SetText("")
			v.applyFilter("")
			v.setFocus(0)
			return nil
		}
		return event
	})

	// Main layout
	v.flex = tview.NewFlex().
		SetDirection(tview.FlexRow)

	// Set up view-specific key handling
	v.flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		logger.Tracef("[KeysView] Input capture received: Key=%v, Rune=%c, FocusIndex=%d",
			event.Key(), event.Rune(), v.focusIndex)

		// Handle view-specific keys only when table has focus (focusIndex == 0)
		if v.focusIndex == 0 {
			logger.Tracef("[KeysView] Handling view-specific keys (table has focus)")
			switch event.Key() {
			case tcell.KeyRune:
				switch event.Rune() {
				case '/':
					logger.Debug("[KeysView] '/' key pressed, focusing filter")
					// Focus on filter (like vim search)
					v.setFocus(1)
					return nil
				case 'r', 'R':
					logger.Debug("[KeysView] 'r' key pressed, reloading keys")
					// Reload/refresh keys
					go v.loadKeys()
					return nil
				// Let all other runes pass through to global handler (numbers, ?, etc.)
				default:
					logger.Tracef("[KeysView] Rune '%c' passed through to global handler", event.Rune())
				}
			case tcell.KeyTab:
				logger.Debug("[KeysView] Tab key pressed, cycling focus")
				// Cycle through focusable elements
				v.cycleFocus()
				return nil
			case tcell.KeyEscape:
				logger.Debug("[KeysView] Escape key pressed, ensuring table focus")
				// Ensure table has focus
				v.setFocus(0)
				return nil
			default:
				logger.Tracef("[KeysView] Key %v passed through to global handler", event.Key())
			}
		} else {
			logger.Tracef("[KeysView] Non-table focus (index=%d), passing key through", v.focusIndex)
		}

		// Let all other events pass through to global handler
		logger.Tracef("[KeysView] Event passed to global handler: Key=%v, Rune=%c", event.Key(), event.Rune())
		return event
	})

	v.refreshLayout()
}

// refreshLayout updates the view layout
func (v *KeysView) refreshLayout() {
	v.flex.Clear()

	// Main content area - left: keys table, right: value details
	mainContent := tview.NewFlex().SetDirection(tview.FlexColumn)

	// Left side - filter and keys table
	leftSide := tview.NewFlex().SetDirection(tview.FlexRow)

	// Filter
	leftSide.AddItem(v.filter, 1, 0, false)

	// Keys table wrapper with border
	keysWrapper := tview.NewFlex().SetDirection(tview.FlexRow)
	keysWrapper.SetBorder(true).SetTitle("Keys")

	// Keys table (no border to avoid double border)
	v.table.SetBorder(false).SetTitle("")
	keysWrapper.AddItem(v.table, 0, 1, true)
	leftSide.AddItem(keysWrapper, 0, 1, true)

	// Right side - value details
	v.keyDetail.SetBorder(true).SetTitle("Value")

	// Add left and right sides to main content
	mainContent.AddItem(leftSide, 0, 1, true)     // Left side gets equal space
	mainContent.AddItem(v.keyDetail, 0, 1, false) // Right side gets equal space

	// Add main content to the main flex (no status bar)
	v.flex.AddItem(mainContent, 0, 1, true)
}

// GetComponent returns the view's main component
func (v *KeysView) GetComponent() tview.Primitive {
	return v.flex
}

// GetKeyCount returns the total number of keys
func (v *KeysView) GetKeyCount() int {
	return len(v.keys)
}

// GetFilter returns the current filter text
func (v *KeysView) GetFilter() string {
	return v.filterText
}

// loadKeys loads and displays Redis keys
func (v *KeysView) loadKeys() {
	logger.Logger.Println("[KeysView] Starting to load keys...")
	
	keys, err := v.redis.GetKeys("*")
	if err != nil {
		logger.Logger.Printf("[KeysView] Error getting keys: %v", err)
		// Show error in the table
		v.showError(fmt.Sprintf("Error loading keys: %v", err))
		return
	}

	logger.Logger.Printf("[KeysView] Found %d keys, loading details...", len(keys))
	
	// Limit the number of keys to process to avoid long waits
	maxKeys := 1000
	if len(keys) > maxKeys {
		logger.Logger.Printf("[KeysView] Too many keys (%d), limiting to first %d keys", len(keys), maxKeys)
		keys = keys[:maxKeys]
	}

	v.keys = make([]*redis.KeyInfo, 0, len(keys))
	
	// Process keys in batches to show progress
	batchSize := 50
	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}
		
		for j := i; j < end; j++ {
			key := keys[j]
			info, err := v.redis.GetKeyInfo(key)
			if err != nil {
				logger.Logger.Printf("[KeysView] Error getting info for key %s: %v", key, err)
				// Create a basic key info even if we can't get all details
				info = &redis.KeyInfo{
					Key:  key,
					Name: key,
					Type: "unknown",
				}
			}
			v.keys = append(v.keys, info)
		}
		
		// Update UI after each batch for progress feedback
		if i%100 == 0 || end == len(keys) {
			logger.Logger.Printf("[KeysView] Processed %d/%d keys", end, len(keys))
			// Update UI synchronously to show progress
			v.refreshKeys()
		}
	}

	logger.Logger.Printf("[KeysView] Finished loading %d keys", len(v.keys))
	v.refreshKeys()
}

// refreshKeys updates the table with current keys
func (v *KeysView) refreshKeys() {
	// Clear existing rows
	v.table.Clear()

	// Set headers (removed Encoding column)
	headers := []string{"Type", "Key", "TTL", "Size"}
	for i, header := range headers {
		v.table.SetCell(0, i,
			tview.NewTableCell(header).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignLeft).
				SetSelectable(false))
	}

	// Add key rows
	displayKeys := v.getDisplayKeys()
	
	if len(displayKeys) == 0 {
		// Show a message when no keys are found
		v.table.SetCell(1, 0, tview.NewTableCell("No keys found"))
		v.table.SetCell(1, 1, tview.NewTableCell(""))
		v.table.SetCell(1, 2, tview.NewTableCell(""))
		v.table.SetCell(1, 3, tview.NewTableCell(""))
		return
	}

	for i, key := range displayKeys {
		row := i + 1

		// Type column
		v.table.SetCell(row, 0, tview.NewTableCell(key.Type))

		// Key name
		v.table.SetCell(row, 1, tview.NewTableCell(key.Name))

		// TTL
		ttl := "-"
		if key.TTL > 0 {
			ttl = fmt.Sprintf("%ds", int64(key.TTL.Seconds()))
		} else if key.TTL == -1 {
			ttl = "persistent"
		}
		v.table.SetCell(row, 2, tview.NewTableCell(ttl))

		// Size
		sizeStr := "-"
		if key.MemoryUsage > 0 {
			sizeStr = humanize.Bytes(uint64(key.MemoryUsage))
		} else if key.Size > 0 {
			sizeStr = humanize.Bytes(uint64(key.Size))
		}
		v.table.SetCell(row, 3, tview.NewTableCell(sizeStr))
	}

	// Select first row if we have data
	if len(displayKeys) > 0 {
		v.table.Select(1, 0) // Select first data row (row 1, column 0)
		v.selectedKey = displayKeys[0].Name
		v.showKeyDetails(displayKeys[0].Name)
	} else {
		v.keyDetail.SetText("No keys found")
	}
}

// getTypeIcon returns an icon for the Redis key type
func (v *KeysView) getTypeIcon(keyType string) string {
	switch strings.ToLower(keyType) {
	case "string":
		return "📄"
	case "hash":
		return "📑"
	case "list":
		return "📝"
	case "set":
		return "📦"
	case "zset":
		return "📊"
	case "stream":
		return "📈"
	default:
		return "❓"
	}
}

// showKeyDetails displays detailed information about a key
func (v *KeysView) showKeyDetails(key string) {
	if key == "" {
		v.keyDetail.SetText("Select a key to view details")
		return
	}

	info, err := v.redis.GetKeyInfo(key)
	if err != nil {
		v.keyDetail.SetText(fmt.Sprintf("Error getting key details: %v", err))
		return
	}

	// Get key value based on type
	value, err := v.redis.GetValue(key)
	if err != nil {
		v.keyDetail.SetText(fmt.Sprintf("Error getting key value: %v", err))
		return
	}

	// Format TTL string
	ttlStr := "never expires"
	if info.TTL > 0 {
		ttlStr = fmt.Sprintf("%v", info.TTL)
	} else if info.TTL < 0 {
		ttlStr = "no expiration"
	}

	details := fmt.Sprintf(`[yellow]Key:[white] %s
[yellow]Type:[white] %s %s
[yellow]TTL:[white] %s
[yellow]Size:[white] %s
[yellow]Encoding:[white] %s

[yellow]Value:[white]
%s`,
		info.Name,
		v.getTypeIcon(info.Type),
		info.Type,
		ttlStr,
		humanize.Bytes(uint64(info.MemoryUsage)),
		info.Encoding,
		value,
	)

	v.keyDetail.SetText(details)
}

// applyFilter filters the keys based on the given pattern
func (v *KeysView) applyFilter(pattern string) {
	v.filterText = pattern
	if pattern == "" {
		v.filteredKeys = nil
		v.refreshKeys()
		return
	}

	filtered := make([]*redis.KeyInfo, 0)
	for _, key := range v.keys {
		if strings.Contains(strings.ToLower(key.Name), strings.ToLower(pattern)) {
			filtered = append(filtered, key)
		}
	}
	v.filteredKeys = filtered
	v.refreshKeys()
}

// executeCommand executes a Redis command and displays the result
func (v *KeysView) executeCommand(command string) {
	if command == "" {
		return
	}

	// Clear the input
	v.commandInput.SetText("")

	// Parse command and arguments
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}

	cmd := parts[0]
	args := make([]interface{}, len(parts)-1)
	for i, arg := range parts[1:] {
		args[i] = arg
	}

	// Execute the command
	result, err := v.redis.ExecuteCommand(cmd, args...)
	if err != nil {
		v.commandOutput.SetText(fmt.Sprintf("[red]Error:[white] %v", err))
	} else {
		// Format the result nicely
		formattedResult := v.formatCommandResult(result)
		v.commandOutput.SetText(fmt.Sprintf("[green]Result:[white] %s", formattedResult))
	}

	// If the command might have changed keys, reload them
	commandLower := strings.ToLower(strings.TrimSpace(cmd))
	if v.shouldReloadKeys(commandLower) {
		go v.loadKeys()
	}
}

// formatCommandResult formats the command result for display
func (v *KeysView) formatCommandResult(result interface{}) string {
	if result == nil {
		return "(nil)"
	}

	switch r := result.(type) {
	case string:
		// Check if it looks like JSON
		if (strings.HasPrefix(r, "{") && strings.HasSuffix(r, "}")) ||
			(strings.HasPrefix(r, "[") && strings.HasSuffix(r, "]")) {
			return r // Return JSON as-is for now
		}
		return fmt.Sprintf(`"%s"`, r)
	case int64:
		return fmt.Sprintf("%d", r)
	case float64:
		return fmt.Sprintf("%.2f", r)
	case bool:
		if r {
			return "true"
		}
		return "false"
	case []interface{}:
		if len(r) == 0 {
			return "(empty list or set)"
		}
		var parts []string
		for i, item := range r {
			if i >= 5 { // Limit to first 5 items
				parts = append(parts, fmt.Sprintf("... (%d more)", len(r)-5))
				break
			}
			if str, ok := item.(string); ok {
				parts = append(parts, fmt.Sprintf(`"%s"`, str))
			} else {
				parts = append(parts, fmt.Sprintf("%v", item))
			}
		}
		return fmt.Sprintf("[\n  %s\n]", strings.Join(parts, ",\n  "))
	case map[string]interface{}:
		if len(r) == 0 {
			return "(empty hash)"
		}
		var parts []string
		count := 0
		for k, v := range r {
			if count >= 5 { // Limit to first 5 pairs
				parts = append(parts, fmt.Sprintf("  ... (%d more)", len(r)-5))
				break
			}
			if str, ok := v.(string); ok {
				parts = append(parts, fmt.Sprintf(`  "%s": "%s"`, k, str))
			} else {
				parts = append(parts, fmt.Sprintf(`  "%s": %v`, k, v))
			}
			count++
		}
		return fmt.Sprintf("{\n%s\n}", strings.Join(parts, ",\n"))
	default:
		return fmt.Sprintf("%v", result)
	}
}

// shouldReloadKeys determines if keys should be reloaded after a command
func (v *KeysView) shouldReloadKeys(command string) bool {
	reloadCommands := []string{
		"set", "del", "rename", "expire", "persist", "flushdb", "flushall",
		"hset", "hdel", "lpush", "lpop", "rpush", "rpop", "sadd", "srem",
		"zadd", "zrem", "xadd", "xdel",
	}

	for _, cmd := range reloadCommands {
		if strings.HasPrefix(command, cmd) {
			return true
		}
	}
	return false
}

// setFocus sets focus to a specific component
func (v *KeysView) setFocus(index int) {
	logger.Debugf("[KeysView] Setting focus from index %d to %d", v.focusIndex, index)
	v.focusIndex = index

	var focusComponent tview.Primitive
	var componentName string
	switch index {
	case 0:
		// Focus on table
		componentName = "table"
		v.table.SetSelectable(true, false)
		focusComponent = v.table
	case 1:
		// Focus on filter
		componentName = "filter"
		v.table.SetSelectable(false, false)
		focusComponent = v.filter
	default:
		componentName = "table (default)"
		focusComponent = v.table
	}

	logger.Debugf("[KeysView] Focus set to: %s (index %d)", componentName, index)

	// Call the focus callback if set
	if v.onFocusChange != nil && focusComponent != nil {
		logger.Tracef("[KeysView] Calling focus callback for component: %s", componentName)
		v.onFocusChange(focusComponent)
	} else {
		if v.onFocusChange == nil {
			logger.Trace("[KeysView] No focus callback set")
		}
		if focusComponent == nil {
			logger.Error("[KeysView] Focus component is nil")
		}
	}
}

// SetFocusCallback sets the callback function for focus changes
func (v *KeysView) SetFocusCallback(callback func(component tview.Primitive)) {
	v.onFocusChange = callback
}

// GetCurrentFocus returns the currently focused component
func (v *KeysView) GetCurrentFocus() tview.Primitive {
	switch v.focusIndex {
	case 0:
		return v.table
	case 1:
		return v.filter
	default:
		return v.table
	}
}

// cycleFocus cycles through focusable components
func (v *KeysView) cycleFocus() {
	v.focusIndex = (v.focusIndex + 1) % 2
	v.setFocus(v.focusIndex)
}

// getDisplayKeys returns the keys to display (filtered or all)
func (v *KeysView) getDisplayKeys() []*redis.KeyInfo {
	if len(v.filteredKeys) > 0 {
		return v.filteredKeys
	}
	return v.keys
}

// Refresh reloads the keys from Redis
func (v *KeysView) Refresh() {
	v.loadKeys()
}

// showError displays an error message in the table
func (v *KeysView) showError(message string) {
	v.table.Clear()
	
	// Set headers
	headers := []string{"Type", "Key", "TTL", "Size"}
	for i, header := range headers {
		v.table.SetCell(0, i,
			tview.NewTableCell(header).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignLeft).
				SetSelectable(false))
	}
	
	// Show error message
	v.table.SetCell(1, 0, tview.NewTableCell("ERROR").SetTextColor(tcell.ColorRed))
	v.table.SetCell(1, 1, tview.NewTableCell(message).SetTextColor(tcell.ColorRed))
	v.table.SetCell(1, 2, tview.NewTableCell(""))
	v.table.SetCell(1, 3, tview.NewTableCell(""))
	
	// Show error in key details as well
	v.keyDetail.SetText(fmt.Sprintf("[red]Error:[white] %s", message))
}
