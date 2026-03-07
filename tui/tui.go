package tui

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ryanwersal/nepenthe/internal/config"
	"github.com/ryanwersal/nepenthe/internal/format"
	"github.com/ryanwersal/nepenthe/internal/scanner"
	"github.com/ryanwersal/nepenthe/internal/state"
	"github.com/ryanwersal/nepenthe/internal/tmutil"
	"golang.org/x/sync/errgroup"
)

type phase int

const (
	phaseScanning phase = iota
	phaseReady
	phaseLoadingAll
)

type view int

const (
	viewScan view = iota
	viewAllExclusions
	viewSettings
)

type Model struct {
	phase        phase
	view         view
	cfg          config.Config
	results      []scanner.ScanResult
	selected     map[int]bool
	cursor       int
	message      string
	scanCount    int
	scanCh       chan scanner.ScanResult
	measuring    bool
	measureCh    chan SizeMeasuredMsg
	measureCount int
	applying     bool
	applyMsg     string
	applyCh      chan any
	applyDone    int
	applyTotal   int
	groupMode    GroupMode
	treeRoot     *TreeNode
	flatRows     []FlatRow
	resultToNode map[int]*TreeNode
	exclusions     []SystemExclusion
	excCursor      int
	settingsCursor int
	settingsItems  []settingsItem
	settingsInput  textinput.Model
	settingsField  settingsEditField
	settingsTmpVal string // holds first value in two-step flows (custom path)
	ctx            context.Context
	cancelScan     context.CancelFunc
	spinner        spinner.Model
	help           help.Model
	tick         int
	width        int
	height       int
	err          error
}

func (m Model) Err() error {
	return m.err
}

func New() (Model, error) {
	cfg, err := config.Load()
	if err != nil {
		return Model{}, err
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	ch := make(chan scanner.ScanResult, 256)
	ctx, cancel := context.WithCancel(context.Background())

	h := help.New()
	h.Styles.ShortKey = helpKeyStyle
	h.Styles.ShortDesc = helpDescStyle
	h.Styles.ShortSeparator = dimStyle
	h.Styles.FullKey = helpKeyStyle
	h.Styles.FullDesc = helpDescStyle
	h.Styles.FullSeparator = dimStyle

	return Model{
		phase:      phaseScanning,
		view:       viewScan,
		cfg:        cfg,
		selected:   make(map[int]bool),
		scanCh:     ch,
		ctx:        ctx,
		cancelScan: cancel,
		groupMode:  GroupByEcosystem,
		spinner:    s,
		help:       h,
		width:      80,
		height:     24,
	}, nil
}

func (m *Model) rebuildTree() {
	if len(m.results) == 0 {
		m.treeRoot = nil
		m.flatRows = nil
		m.resultToNode = nil
		return
	}

	// Save current node for cursor stability
	var currentNode *TreeNode
	if m.flatRows != nil && m.cursor < len(m.flatRows) {
		currentNode = m.flatRows[m.cursor].Node
	}

	switch m.groupMode {
	case GroupByDirectory:
		m.treeRoot, m.resultToNode = buildDirectoryTree(m.results)
	case GroupByEcosystem:
		m.treeRoot, m.resultToNode = buildEcosystemTree(m.results)
	}

	m.flatRows = flattenTree(m.treeRoot)

	// Restore cursor position
	if currentNode != nil {
		for i, row := range m.flatRows {
			if row.Node == currentNode {
				m.cursor = i
				return
			}
		}
		// Node collapsed away - try parent
		for p := currentNode.Parent; p != nil; p = p.Parent {
			for i, row := range m.flatRows {
				if row.Node == p {
					m.cursor = i
					return
				}
			}
		}
	}

	m.cursor = 0
}

func (m *Model) reflattenTree() {
	// Save current node for cursor stability
	var currentNode *TreeNode
	if m.flatRows != nil && m.cursor < len(m.flatRows) {
		currentNode = m.flatRows[m.cursor].Node
	}

	m.flatRows = flattenTree(m.treeRoot)

	// Restore cursor position
	if currentNode != nil {
		for i, row := range m.flatRows {
			if row.Node == currentNode {
				m.cursor = i
				return
			}
		}
		// Node collapsed away - try parent
		for p := currentNode.Parent; p != nil; p = p.Parent {
			for i, row := range m.flatRows {
				if row.Node == p {
					m.cursor = i
					return
				}
			}
		}
	}
	if m.cursor >= len(m.flatRows) {
		m.cursor = max(0, len(m.flatRows)-1)
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.startScan(),
		m.waitForScanResult(),
	)
}

func (m Model) startScan() tea.Cmd {
	ch := m.scanCh
	cfg := m.cfg
	ctx := m.ctx
	return func() tea.Msg {
		defer close(ch)

		rules := scanner.BuildSentinelRules()

		scanner.ScanSentinelRules(ctx, scanner.WalkOptions{
			Roots: cfg.Roots,
			Rules: rules,
			OnFound: func(r scanner.ScanResult) {
				ch <- r
			},
		})

		var customFixed []scanner.FixedPathRule
		for _, cf := range cfg.CustomFixedPaths {
			customFixed = append(customFixed, scanner.FixedPathRule{
				Path:      cf.Path,
				Ecosystem: cf.Ecosystem,
				Category:  scanner.CategoryCustom,
			})
		}
		fixedResults, err := scanner.ScanFixedPaths(ctx, scanner.AllCategories, customFixed)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		for _, r := range fixedResults {
			ch <- r
		}

		// ScanDoneMsg will be sent by waitForScanResult when channel closes
		return nil
	}
}

func (m Model) waitForScanResult() tea.Cmd {
	ch := m.scanCh
	return func() tea.Msg {
		r, ok := <-ch
		if !ok {
			return ScanDoneMsg{}
		}
		return ScanResultMsg{Result: r}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		return m, nil

	case spinner.TickMsg:
		m.tick++
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case ScanResultMsg:
		idx := len(m.results)
		m.results = append(m.results, msg.Result)
		m.scanCount++
		if !msg.Result.IsExcluded {
			m.selected[idx] = true
		}
		m.rebuildTree()
		return m, m.waitForScanResult()

	case ScanDoneMsg:
		m.phase = phaseReady
		m.message = fmt.Sprintf("Scan complete — %d directories found", len(m.results))
		m.rebuildTree()
		if len(m.results) > 0 {
			m.measuring = true
			m.measureCh = make(chan SizeMeasuredMsg, 256)
			m.measureCount = 0
			return m, tea.Batch(m.startMeasure(), m.waitForSizeMeasured())
		}
		return m, nil

	case ApplyProgressMsg:
		m.applyDone = msg.Done
		if msg.Success && msg.Index < len(m.results) {
			m.results[msg.Index].IsExcluded = true
			delete(m.selected, msg.Index)
		}
		return m, m.waitForApplyProgress()

	case ApplyDoneMsg:
		m.applying = false
		m.applyCh = nil
		m.message = fmt.Sprintf("Excluded %d directories", msg.Applied)
		if msg.Failed > 0 {
			m.message += fmt.Sprintf(" (%d failed)", msg.Failed)
		}
		return m, nil

	case RemoveProgressMsg:
		m.applyDone = msg.Done
		if msg.Success && msg.Index < len(m.results) {
			m.results[msg.Index].IsExcluded = false
			delete(m.selected, msg.Index)
		}
		return m, m.waitForApplyProgress()

	case RemoveDoneMsg:
		m.applying = false
		m.applyCh = nil
		m.message = fmt.Sprintf("Removed %d exclusions", msg.Removed)
		if msg.Failed > 0 {
			m.message += fmt.Sprintf(" (%d failed)", msg.Failed)
		}
		return m, nil

	case SizeMeasuredMsg:
		if msg.Index < len(m.results) {
			m.results[msg.Index].SizeBytes = msg.SizeBytes
			m.results[msg.Index].FileCount = msg.FileCount
			m.measureCount++
			if m.resultToNode != nil {
				updateNodeSize(m.resultToNode, msg.Index, msg.SizeBytes, msg.FileCount)
			}
		}
		return m, m.waitForSizeMeasured()

	case AllSizesDoneMsg:
		m.measuring = false
		var totalSize int64
		for _, r := range m.results {
			totalSize += r.SizeBytes
		}
		m.message = fmt.Sprintf("Sizes measured — %s total", format.Bytes(totalSize))
		return m, nil

	case MeasureDoneMsg:
		m.phase = phaseReady
		m.results = msg.Results
		m.message = "Sizes measured"
		return m, nil

	case AllExclusionsMsg:
		m.phase = phaseReady
		st, _ := state.Load()
		m.exclusions = categorizeExclusions(msg.Paths, &st)
		m.view = viewAllExclusions
		m.excCursor = 0
		return m, nil

	case ConfigUpdatedMsg:
		if msg.Err != nil {
			m.message = "Config save error: " + msg.Err.Error()
		} else {
			m.cfg = msg.Cfg
			m.settingsItems = buildSettingsItems(m.cfg)
		}
		return m, nil

	case ErrorMsg:
		m.err = msg.Err
		return m, tea.Quit

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Settings view handles its own quit/escape logic
	if m.view == viewSettings {
		return m.handleSettingsKey(msg)
	}

	if key.Matches(msg, keys.Quit) {
		if m.cancelScan != nil {
			m.cancelScan()
		}
		return m, tea.Quit
	}

	if m.phase != phaseReady {
		return m, nil
	}

	if m.view == viewAllExclusions {
		return m.handleExclusionsKey(msg)
	}

	return m.handleScanKey(msg)
}

func (m Model) handleScanKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// During scanning phase, only show flat streaming list
	if m.phase == phaseScanning {
		return m, nil
	}

	// Tree-based navigation (post-scan)
	if m.flatRows != nil {
		return m.handleTreeKey(msg)
	}

	return m, nil
}

func (m Model) handleTreeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	rowCount := len(m.flatRows)

	switch {
	case key.Matches(msg, keys.Down):
		if m.cursor < rowCount-1 {
			m.cursor++
		}
	case key.Matches(msg, keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, keys.Expand):
		if m.cursor < rowCount {
			row := m.flatRows[m.cursor]
			if !row.IsLeaf && !row.Node.Expanded {
				row.Node.Expanded = true
				m.reflattenTree()
			}
		}
	case key.Matches(msg, keys.Collapse):
		if m.cursor < rowCount {
			row := m.flatRows[m.cursor]
			if !row.IsLeaf && row.Node.Expanded {
				// Collapse this internal node
				row.Node.Expanded = false
				m.reflattenTree()
			} else if row.Node.Parent != nil && row.Node.Parent.Parent != nil {
				// Jump to parent (but not the root)
				for i, r := range m.flatRows {
					if r.Node == row.Node.Parent {
						m.cursor = i
						break
					}
				}
			}
		}
	case key.Matches(msg, keys.Toggle):
		if m.cursor < rowCount {
			row := m.flatRows[m.cursor]
			if row.IsLeaf && row.Node.ResultIdx >= 0 {
				// Toggle single leaf
				idx := row.Node.ResultIdx
				if m.selected[idx] {
					delete(m.selected, idx)
				} else {
					m.selected[idx] = true
				}
			} else {
				// Toggle all descendant leaves
				indices := leafIndices(row.Node)
				allSelected := true
				for _, idx := range indices {
					if !m.results[idx].IsExcluded && !m.selected[idx] {
						allSelected = false
						break
					}
				}
				if allSelected {
					// Deselect all
					for _, idx := range indices {
						delete(m.selected, idx)
					}
				} else {
					// Select all non-excluded
					for _, idx := range indices {
						if !m.results[idx].IsExcluded {
							m.selected[idx] = true
						}
					}
				}
			}
		}
	case key.Matches(msg, keys.Group):
		if m.groupMode == GroupByDirectory {
			m.groupMode = GroupByEcosystem
		} else {
			m.groupMode = GroupByDirectory
		}
		m.rebuildTree()
	case key.Matches(msg, keys.All):
		for i, r := range m.results {
			if !r.IsExcluded {
				m.selected[i] = true
			}
		}
	case key.Matches(msg, keys.None):
		m.selected = make(map[int]bool)
	case key.Matches(msg, keys.Apply):
		if m.applying {
			break
		}
		m.applying = true
		m.applyMsg = "Applying exclusions..."
		cmd := m.startApplyExclusions()
		return m, tea.Batch(cmd, m.waitForApplyProgress())
	case key.Matches(msg, keys.Remove):
		if m.applying {
			break
		}
		m.applying = true
		m.applyMsg = "Removing exclusions..."
		cmd := m.startRemoveExclusions()
		return m, tea.Batch(cmd, m.waitForApplyProgress())
	case key.Matches(msg, keys.SwitchV):
		m.phase = phaseLoadingAll
		m.message = "Loading all exclusions..."
		return m, m.loadAllExclusions()
	case key.Matches(msg, keys.Settings):
		m.settingsItems = buildSettingsItems(m.cfg)
		m.settingsCursor = firstSelectableIndex(m.settingsItems)
		m.settingsField = editNone
		m.view = viewSettings
	}
	return m, nil
}

func (m Model) handleExclusionsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Down):
		if m.excCursor < len(m.exclusions)-1 {
			m.excCursor++
		}
	case key.Matches(msg, keys.Up):
		if m.excCursor > 0 {
			m.excCursor--
		}
	case key.Matches(msg, keys.SwitchV):
		m.view = viewScan
	case key.Matches(msg, keys.Settings):
		m.settingsItems = buildSettingsItems(m.cfg)
		m.settingsCursor = firstSelectableIndex(m.settingsItems)
		m.settingsField = editNone
		m.view = viewSettings
	}
	return m, nil
}

func (m Model) handleSettingsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Editing mode: forward keys to text input
	if m.settingsField != editNone {
		return m.handleSettingsEditKey(msg)
	}

	// Browsing mode
	if key.Matches(msg, keys.Quit) {
		if m.cancelScan != nil {
			m.cancelScan()
		}
		return m, tea.Quit
	}

	switch {
	case key.Matches(msg, keys.Down):
		m.settingsCursor = m.nextSelectableCursor(1)
	case key.Matches(msg, keys.Up):
		m.settingsCursor = m.nextSelectableCursor(-1)

	case key.Matches(msg, keys.Toggle):
		// Toggle category
		if m.settingsCursor < len(m.settingsItems) {
			item := m.settingsItems[m.settingsCursor]
			if item.Kind == settingsToggle {
				return m, m.toggleCategory(item.Value)
			}
		}

	case key.Matches(msg, keys.Apply):
		// Enter on add button or value field -> start editing
		if m.settingsCursor < len(m.settingsItems) {
			item := m.settingsItems[m.settingsCursor]
			switch item.Kind {
			case settingsAddButton:
				m.startSettingsEdit(item.EditField, "")
			case settingsValue:
				m.startSettingsEdit(item.EditField, item.Value)
			}
		}

	case key.Matches(msg, keys.Delete):
		// Delete root or custom path
		if m.settingsCursor < len(m.settingsItems) {
			item := m.settingsItems[m.settingsCursor]
			if item.Kind == settingsPath && item.Deletable {
				return m, m.deleteSettingsItem(item)
			}
		}

	case key.Matches(msg, keys.Settings):
		// 's' returns to scan view
		m.view = viewScan
	}

	// Handle esc separately since it's not in the keymap
	if msg.Type == tea.KeyEsc {
		m.view = viewScan
	}

	return m, nil
}

func (m Model) handleSettingsEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Cancel editing
		m.settingsField = editNone
		return m, nil
	case tea.KeyEnter:
		val := strings.TrimSpace(m.settingsInput.Value())
		if val == "" {
			m.settingsField = editNone
			return m, nil
		}
		return m.confirmSettingsEdit(val)
	}

	// Forward to text input
	var cmd tea.Cmd
	m.settingsInput, cmd = m.settingsInput.Update(msg)
	return m, cmd
}

func (m *Model) startSettingsEdit(field settingsEditField, initial string) {
	m.settingsField = field
	ti := textinput.New()
	switch field {
	case editRoot:
		ti.Placeholder = "~/path/to/scan"
		ti.Prompt = "Root: "
	case editCustomPath:
		ti.Placeholder = "~/path/to/exclude"
		ti.Prompt = "Path: "
	case editCustomEcosystem:
		ti.Placeholder = "e.g. Node.js"
		ti.Prompt = "Ecosystem: "
	}
	ti.SetValue(initial)
	ti.Focus()
	m.settingsInput = ti
}

func (m Model) confirmSettingsEdit(val string) (tea.Model, tea.Cmd) {
	switch m.settingsField {
	case editRoot:
		m.settingsField = editNone
		return m, m.addRoot(val)
	case editCustomPath:
		// Two-step: save path, prompt for ecosystem
		m.settingsTmpVal = val
		m.startSettingsEdit(editCustomEcosystem, "")
		return m, nil
	case editCustomEcosystem:
		path := m.settingsTmpVal
		m.settingsField = editNone
		m.settingsTmpVal = ""
		return m, m.addCustomFixedPath(path, val)
	}
	m.settingsField = editNone
	return m, nil
}

func (m Model) nextSelectableCursor(dir int) int {
	c := m.settingsCursor
	for {
		c += dir
		if c < 0 || c >= len(m.settingsItems) {
			return m.settingsCursor // don't wrap
		}
		if m.settingsItems[c].Kind != settingsHeader {
			return c
		}
	}
}

// Config mutation commands that run asynchronously and return ConfigUpdatedMsg.

func (m Model) toggleCategory(category string) tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.ToggleCategory(category)
		return ConfigUpdatedMsg{Cfg: cfg, Err: err}
	}
}

func (m Model) addRoot(root string) tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.AddRoot(root)
		return ConfigUpdatedMsg{Cfg: cfg, Err: err}
	}
}

func (m Model) deleteSettingsItem(item settingsItem) tea.Cmd {
	// Determine which section we're in by looking at preceding headers
	section := ""
	for i := m.settingsCursor; i >= 0; i-- {
		if m.settingsItems[i].Kind == settingsHeader {
			section = m.settingsItems[i].Label
			break
		}
	}
	return func() tea.Msg {
		var cfg config.Config
		var err error
		switch section {
		case "Scan Roots":
			cfg, err = config.RemoveRoot(item.Value)
		case "Custom Paths":
			cfg, err = config.RemoveCustomFixedPath(item.Value)
		}
		return ConfigUpdatedMsg{Cfg: cfg, Err: err}
	}
}

func (m Model) addCustomFixedPath(path, ecosystem string) tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.AddCustomFixedPath(path, ecosystem)
		return ConfigUpdatedMsg{Cfg: cfg, Err: err}
	}
}

func (m *Model) startApplyExclusions() tea.Cmd {
	results := make([]scanner.ScanResult, len(m.results))
	copy(results, m.results)
	selected := make(map[int]bool, len(m.selected))
	for k, v := range m.selected {
		selected[k] = v
	}

	// Count how many we'll process
	total := 0
	for i, r := range results {
		if selected[i] && !r.IsExcluded {
			total++
		}
	}
	m.applyTotal = total
	m.applyDone = 0
	m.applyCh = make(chan any, 256)

	ch := m.applyCh
	return func() tea.Msg {
		st, err := state.Load()
		if err != nil {
			ch <- ErrorMsg{Err: err}
			close(ch)
			return nil
		}

		type successInfo struct {
			path, category, typ, ecosystem string
		}

		var (
			mu        sync.Mutex
			done      int
			failed    int
			successes []successInfo
		)

		g := new(errgroup.Group)
		g.SetLimit(16)

		for i, r := range results {
			if !selected[i] || r.IsExcluded {
				continue
			}
			idx := i
			g.Go(func() error {
				success := false
				if err := tmutil.AddExclusion(r.Path); err != nil {
					mu.Lock()
					failed++
					done++
					d := done
					mu.Unlock()
					ch <- ApplyProgressMsg{Index: idx, Success: false, Done: d, Total: total}
				} else {
					mu.Lock()
					successes = append(successes, successInfo{r.Path, string(r.Category), r.Type, r.Ecosystem})
					done++
					d := done
					mu.Unlock()
					success = true
					ch <- ApplyProgressMsg{Index: idx, Success: success, Done: d, Total: total}
				}
				return nil
			})
		}
		_ = g.Wait()

		for _, s := range successes {
			state.AddExclusion(&st, s.path, s.category, s.typ, s.ecosystem)
		}

		if err := state.Save(st); err != nil {
			ch <- ErrorMsg{Err: err}
			close(ch)
			return nil
		}
		ch <- ApplyDoneMsg{Applied: len(successes), Failed: failed}
		close(ch)
		return nil
	}
}

func (m *Model) startRemoveExclusions() tea.Cmd {
	results := make([]scanner.ScanResult, len(m.results))
	copy(results, m.results)
	selected := make(map[int]bool, len(m.selected))
	for k, v := range m.selected {
		selected[k] = v
	}

	total := 0
	for i, r := range results {
		if selected[i] && r.IsExcluded {
			total++
		}
	}
	m.applyTotal = total
	m.applyDone = 0
	m.applyCh = make(chan any, 256)

	ch := m.applyCh
	return func() tea.Msg {
		st, err := state.Load()
		if err != nil {
			ch <- ErrorMsg{Err: err}
			close(ch)
			return nil
		}

		var (
			mu           sync.Mutex
			done         int
			failed       int
			removedPaths []string
		)

		g := new(errgroup.Group)
		g.SetLimit(16)

		for i, r := range results {
			if !selected[i] || !r.IsExcluded {
				continue
			}
			idx := i
			g.Go(func() error {
				success := false
				if err := tmutil.RemoveExclusion(r.Path); err != nil {
					mu.Lock()
					failed++
					done++
					d := done
					mu.Unlock()
					ch <- RemoveProgressMsg{Index: idx, Success: false, Done: d, Total: total}
				} else {
					mu.Lock()
					removedPaths = append(removedPaths, r.Path)
					done++
					d := done
					mu.Unlock()
					success = true
					ch <- RemoveProgressMsg{Index: idx, Success: success, Done: d, Total: total}
				}
				return nil
			})
		}
		_ = g.Wait()

		for _, p := range removedPaths {
			state.RemoveExclusion(&st, p)
		}

		if err := state.Save(st); err != nil {
			ch <- ErrorMsg{Err: err}
			close(ch)
			return nil
		}
		ch <- RemoveDoneMsg{Removed: len(removedPaths), Failed: failed}
		close(ch)
		return nil
	}
}

func (m Model) waitForApplyProgress() tea.Cmd {
	ch := m.applyCh
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

func (m Model) startMeasure() tea.Cmd {
	results := make([]scanner.ScanResult, len(m.results))
	copy(results, m.results)
	ch := m.measureCh
	ctx := m.ctx
	return func() tea.Msg {
		scanner.MeasureSizesStream(ctx, results, func(sm scanner.SizeMeasurement) {
			ch <- SizeMeasuredMsg{
				Index:     sm.Index,
				SizeBytes: sm.SizeBytes,
				FileCount: sm.FileCount,
			}
		})
		close(ch)
		return nil
	}
}

func (m Model) waitForSizeMeasured() tea.Cmd {
	ch := m.measureCh
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return AllSizesDoneMsg{}
		}
		return msg
	}
}

func (m Model) loadAllExclusions() tea.Cmd {
	return func() tea.Msg {
		paths, err := tmutil.ListAllExclusions()
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return AllExclusionsMsg{Paths: paths}
	}
}

// contentWidth returns the usable width inside the content border.
func (m Model) contentWidth() int {
	// Rounded border uses 1 char on each side + 1 padding on each side = 4
	w := m.width - 4
	if w < 20 {
		w = 20
	}
	return w
}

// contentHeight returns the usable height for list content.
func (m Model) contentHeight() int {
	// header bar(1) + status(~2) + border top/bottom(2) + footer(~2)
	h := m.height - 7
	if h < 1 {
		h = 1
	}
	return h
}

func (m Model) View() string {
	// 1. Header bar with tabs
	header := m.renderHeader()

	// 2. Status line (view-specific)
	status := m.renderStatus()

	// 3. Content (view-specific)
	cw := m.contentWidth()
	ch := m.contentHeight()
	var content string
	switch m.view {
	case viewScan:
		if m.flatRows != nil {
			content = renderTreeView(m.flatRows, m.results, m.selected, m.cursor, cw, ch)
		} else {
			content = renderScanView(m.results, m.selected, m.cursor, cw, ch)
		}
	case viewAllExclusions:
		content = renderExclusionsView(m.exclusions, m.excCursor, cw, ch)
	case viewSettings:
		content = renderSettingsView(m.settingsItems, m.settingsCursor, cw, ch)
		if m.settingsField != editNone {
			content += "\n" + m.settingsInput.View()
		}
	}

	// 4. Wrap content in border
	bordered := contentBorderStyle.Width(m.width - 2).Render(content)

	// 5. Footer with help keys
	footer := m.renderFooter()

	// 6. Join vertically
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		status,
		bordered,
		footer,
	)
}

// renderHeader builds the full-width header bar with tab pills.
func (m Model) renderHeader() string {
	tabs := []struct {
		label string
		v     view
	}{
		{"Scan", viewScan},
		{"Exclusions", viewAllExclusions},
		{"Settings", viewSettings},
	}

	var tabParts []string
	for _, t := range tabs {
		if t.v == m.view {
			tabParts = append(tabParts, activeTabStyle.Render(t.label))
		} else {
			tabParts = append(tabParts, inactiveTabStyle.Render(t.label))
		}
	}

	title := headerBarStyle.Render("Nepenthe")
	tabStr := strings.Join(tabParts, " ")

	// Fill the rest of the bar width
	barContent := title + "  " + tabStr
	barWidth := lipgloss.Width(barContent)
	padding := m.width - barWidth
	if padding < 0 {
		padding = 0
	}

	return headerBarStyle.Width(m.width).Render(
		"Nepenthe  " + tabStr + strings.Repeat(" ", padding),
	)
}

// renderStatus builds the view-specific status line.
func (m Model) renderStatus() string {
	var b strings.Builder

	switch m.view {
	case viewScan:
		excludedCount := 0
		for _, r := range m.results {
			if r.IsExcluded {
				excludedCount++
			}
		}

		if m.phase == phaseScanning {
			b.WriteString(m.spinner.View())
			fmt.Fprintf(&b, " Scanning... %d found  ", m.scanCount)
			b.WriteString(renderIndeterminateBar(20, m.tick))
		} else {
			b.WriteString(statusBarStyle.Render(
				fmt.Sprintf("%d/%d excluded", excludedCount, len(m.results)),
			))
			if m.message != "" {
				b.WriteString(dimStyle.Render(" — " + m.message))
			}
			if m.flatRows != nil {
				switch m.groupMode {
				case GroupByDirectory:
					b.WriteString(dimStyle.Render("  [directory]"))
				case GroupByEcosystem:
					b.WriteString(dimStyle.Render("  [ecosystem]"))
				}
			}
			if m.applying && m.applyTotal > 0 {
				b.WriteByte('\n')
				b.WriteString(m.spinner.View())
				fmt.Fprintf(&b, " %s  ", m.applyMsg)
				b.WriteString(renderProgressBar(30, m.applyDone, m.applyTotal))
				b.WriteString(dimStyle.Render(fmt.Sprintf("  %d/%d", m.applyDone, m.applyTotal)))
			}
			if m.measuring {
				b.WriteByte('\n')
				b.WriteString(m.spinner.View())
				b.WriteString(" Measuring sizes  ")
				b.WriteString(renderProgressBar(30, m.measureCount, len(m.results)))
				b.WriteString(dimStyle.Render(fmt.Sprintf("  %d/%d", m.measureCount, len(m.results))))
			}
		}

	case viewAllExclusions:
		nepentheCount := 0
		for _, e := range m.exclusions {
			if e.Source == "nepenthe" {
				nepentheCount++
			}
		}
		otherCount := len(m.exclusions) - nepentheCount
		b.WriteString(statusBarStyle.Render(
			fmt.Sprintf("%d total (%d by Nepenthe, %d by other tools)",
				len(m.exclusions), nepentheCount, otherCount),
		))

	case viewSettings:
		if m.message != "" {
			b.WriteString(dimStyle.Render(m.message))
		}
	}

	return b.String()
}

// renderFooter builds the help key footer using bubbles/help.
func (m Model) renderFooter() string {
	switch m.view {
	case viewScan:
		return m.help.View(scanKeyMap{keys})
	case viewAllExclusions:
		return m.help.View(exclusionsKeyMap{keys})
	case viewSettings:
		if m.settingsField != editNone {
			return m.help.View(settingsEditKeyMap{})
		}
		return m.help.View(settingsKeyMap{keys})
	}
	return ""
}
