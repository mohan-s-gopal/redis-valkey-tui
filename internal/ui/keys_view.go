package ui

import (
	"fmt"
	"strings"

	"redis-cli-dashboard/internal/config"
	"redis-cli-dashboard/internal/redis"

	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// KeysView represents the keys view
type KeysView struct {
	redis  *redis.Client
	config *config.Config

	// Components
	flex      *tview.Flex
	table     *tview.Table
	keyDetail *tview.TextView
	filter    *tview.InputField

	// State
	keys         []*redis.KeyInfo
	filteredKeys []*redis.KeyInfo
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
}	// setupUI initializes the UI components
func (v *KeysView) setupUI() {
	// Create table for keys
	v.table = tview.NewTable().
		SetBorders(false).
		SetFixed(1, 0).
		SetSelectable(true, false)

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
		if row > 0 && row <= len(v.filteredKeys) {
			displayKeys := v.filteredKeys
			if len(displayKeys) == 0 {
				displayKeys = v.keys
			}
			keyInfo := displayKeys[row-1]
			if keyInfo != nil {
				v.selectedKey = keyInfo.Name
				v.showKeyDetails(keyInfo.Name)
			}
		}
	})

	// Handle enter key on selection
	v.table.SetSelectedFunc(func(row, col int) {
		if row > 0 && row <= len(v.filteredKeys) {
			displayKeys := v.filteredKeys
			if len(displayKeys) == 0 {
				displayKeys = v.keys
			}
			keyInfo := displayKeys[row-1]
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
		SetFieldWidth(30).
		SetDoneFunc(func(key tcell.Key) {
			v.applyFilter(v.filter.GetText())
		})

	// Main layout
	v.flex = tview.NewFlex().
		SetDirection(tview.FlexRow)
	
	v.refreshLayout()
}

// refreshLayout updates the view layout
func (v *KeysView) refreshLayout() {
	v.flex.Clear()

	// Create main horizontal flex for splitting table and details
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn)

	// Left side with filter and table
	leftFlex := tview.NewFlex().
		SetDirection(tview.FlexRow)

	// Add filter if visible
	if v.filterVisible {
		filterBox := tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(v.filter, 30, 0, true).
			AddItem(nil, 0, 1, false)
		leftFlex.AddItem(filterBox, 1, 0, true)
	}

	// Create table box with border
	tableBox := tview.NewBox().
		SetBorder(true).
		SetTitle("Redis Keys")

	// Add table with border
	leftFlex.AddItem(tableBox, 0, 1, true)

	// Make sure table fills the box
	tableBox.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		innerX, innerY, innerWidth, innerHeight := tableBox.GetInnerRect()
		v.table.SetRect(innerX, innerY, innerWidth, innerHeight)
		v.table.Draw(screen)
		return x, y, width, height
	})

	// Right side with key details
	v.keyDetail.SetBorder(true).
		SetTitle("Key Details")

	// Set up the split view with good proportions
	mainFlex.AddItem(leftFlex, 0, 2, true).        // Left side gets 2/3
		AddItem(v.keyDetail, 0, 1, false)  // Right side gets 1/3

	// Add the split view to the main flex
	v.flex.AddItem(mainFlex, 0, 1, true)

	// We've already added contentFlex to v.flex in the code above
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
	keys, err := v.redis.GetKeys("*")
	if err != nil {
		return
	}

	v.keys = make([]*redis.KeyInfo, 0, len(keys))
	for _, key := range keys {
		info, err := v.redis.GetKeyInfo(key)
		if err != nil {
			continue
		}
		v.keys = append(v.keys, info)
	}

	v.refreshKeys()
}

// refreshKeys updates the table with current keys
func (v *KeysView) refreshKeys() {
	// Clear existing rows
	v.table.Clear()

	// Set headers
	headers := []string{"Type", "Key", "TTL", "Size", "Encoding"}
	for i, header := range headers {
		v.table.SetCell(0, i,
			tview.NewTableCell(header).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignLeft).
				SetSelectable(false))
	}

	// Add key rows
	displayKeys := v.filteredKeys
	if len(displayKeys) == 0 {
		displayKeys = v.keys
	}

	for i, key := range displayKeys {
		row := i + 1
		
		// Type column with icon
		typeIcon := v.getTypeIcon(key.Type)
		v.table.SetCell(row, 0, tview.NewTableCell(typeIcon+" "+key.Type))
		
		// Key name
		v.table.SetCell(row, 1, tview.NewTableCell(key.Name))
		
		// TTL
		ttl := "-"
		if key.TTL > 0 {
			ttl = fmt.Sprintf("%ds", int64(key.TTL.Seconds()))
		}
		v.table.SetCell(row, 2, tview.NewTableCell(ttl))
		
		// Size
		v.table.SetCell(row, 3, tview.NewTableCell(humanize.Bytes(uint64(key.MemoryUsage))))
		
		// Encoding
		v.table.SetCell(row, 4, tview.NewTableCell(key.Encoding))
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

// Refresh reloads the keys from Redis
func (v *KeysView) Refresh() {
	v.loadKeys()
}


