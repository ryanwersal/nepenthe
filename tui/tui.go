package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ryanwersal/nepenthe/internal/config"
	"github.com/ryanwersal/nepenthe/internal/format"
	"github.com/ryanwersal/nepenthe/internal/scanner"
	"github.com/ryanwersal/nepenthe/internal/state"
	"github.com/ryanwersal/nepenthe/internal/tmutil"
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
	exclusions   []SystemExclusion
	excCursor    int
	spinner      spinner.Model
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

	return Model{
		phase:     phaseScanning,
		view:      viewScan,
		cfg:       cfg,
		selected:  make(map[int]bool),
		scanCh:    ch,
		groupMode: GroupByEcosystem,
		spinner:   s,
		width:     80,
		height:    24,
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
	return func() tea.Msg {
		defer close(ch)

		custom := make([]scanner.SentinelRule, 0, len(cfg.CustomSentinelRules))
		for _, cr := range cfg.CustomSentinelRules {
			custom = append(custom, scanner.SentinelRule{
				Directory: cr.Directory,
				Sentinels: cr.Sentinels,
				Ecosystem: cr.Ecosystem,
			})
		}
		rules := scanner.BuildSentinelRules(custom)

		scanner.ScanSentinelRules(scanner.WalkOptions{
			Roots:       cfg.Roots,
			Rules:       rules,
			Concurrency: cfg.Concurrency.ScanWorkers,
			OnFound: func(r scanner.ScanResult) {
				ch <- r
			},
		})

		fixedResults, err := scanner.ScanFixedPaths(cfg.EnabledTiers, nil)
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

	case ErrorMsg:
		m.err = msg.Err
		return m, tea.Quit

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, keys.Quit) {
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
				// Check if all are selected
				allSelected := true
				for _, idx := range indices {
					if !m.selected[idx] {
						allSelected = true
						break
					}
				}
				// If checking properly: see if all non-excluded are selected
				allSelected = true
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
	}
	return m, nil
}

func (m *Model) startApplyExclusions() tea.Cmd {
	results := m.results
	selected := m.selected

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

		var applied, failed, done int
		for i, r := range results {
			if !selected[i] || r.IsExcluded {
				continue
			}
			success := false
			if err := tmutil.AddExclusion(r.Path); err != nil {
				failed++
			} else {
				state.AddExclusion(&st, r.Path, r.Tier, r.Type, r.Ecosystem)
				applied++
				success = true
			}
			done++
			ch <- ApplyProgressMsg{Index: i, Success: success, Done: done, Total: total}
		}

		if err := state.Save(st); err != nil {
			ch <- ErrorMsg{Err: err}
			close(ch)
			return nil
		}
		ch <- ApplyDoneMsg{Applied: applied, Failed: failed}
		close(ch)
		return nil
	}
}

func (m *Model) startRemoveExclusions() tea.Cmd {
	results := m.results
	selected := m.selected

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

		var removed, failed, done int
		for i, r := range results {
			if !selected[i] || !r.IsExcluded {
				continue
			}
			success := false
			if err := tmutil.RemoveExclusion(r.Path); err != nil {
				failed++
			} else {
				state.RemoveExclusion(&st, r.Path)
				removed++
				success = true
			}
			done++
			ch <- RemoveProgressMsg{Index: i, Success: success, Done: done, Total: total}
		}

		if err := state.Save(st); err != nil {
			ch <- ErrorMsg{Err: err}
			close(ch)
			return nil
		}
		ch <- RemoveDoneMsg{Removed: removed, Failed: failed}
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
	concurrency := m.cfg.Concurrency.MeasureWorkers
	return func() tea.Msg {
		scanner.MeasureSizesStream(results, concurrency, func(sm scanner.SizeMeasurement) {
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

func (m Model) View() string {
	var b strings.Builder

	switch m.view {
	case viewScan:
		m.renderScanHeader(&b)
		if m.flatRows != nil {
			b.WriteString(renderTreeView(m.flatRows, m.results, m.selected, m.cursor, m.width, m.height))
		} else {
			b.WriteString(renderScanView(m.results, m.selected, m.cursor, m.width, m.height))
		}
		b.WriteByte('\n')
		m.renderScanFooter(&b)

	case viewAllExclusions:
		m.renderExclusionsHeader(&b)
		b.WriteString(renderExclusionsView(m.exclusions, m.excCursor, m.width, m.height))
		b.WriteByte('\n')
		m.renderExclusionsFooter(&b)
	}

	return b.String()
}

func (m Model) renderScanHeader(b *strings.Builder) {
	title := "Nepenthe"
	if m.flatRows != nil {
		switch m.groupMode {
		case GroupByDirectory:
			title += "  " + dimStyle.Render("[grouped by directory]")
		case GroupByEcosystem:
			title += "  " + dimStyle.Render("[grouped by ecosystem]")
		}
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteByte('\n')

	// Status line
	excludedCount := 0
	for _, r := range m.results {
		if r.IsExcluded {
			excludedCount++
		}
	}

	if m.phase == phaseScanning {
		b.WriteString(m.spinner.View())
		b.WriteString(fmt.Sprintf(" Scanning... %d found  ", m.scanCount))
		b.WriteString(renderIndeterminateBar(20, m.tick))
		b.WriteByte('\n')
	} else {
		b.WriteString(statusBarStyle.Render(
			fmt.Sprintf("%d/%d excluded", excludedCount, len(m.results)),
		))
		if m.message != "" {
			b.WriteString(dimStyle.Render(" — " + m.message))
		}
		b.WriteByte('\n')
		if m.applying && m.applyTotal > 0 {
			b.WriteString(m.spinner.View())
			b.WriteString(fmt.Sprintf(" %s  ", m.applyMsg))
			b.WriteString(renderProgressBar(30, m.applyDone, m.applyTotal))
			b.WriteString(dimStyle.Render(fmt.Sprintf("  %d/%d", m.applyDone, m.applyTotal)))
			b.WriteByte('\n')
		}
		if m.measuring {
			b.WriteString(m.spinner.View())
			b.WriteString(fmt.Sprintf(" Measuring sizes  "))
			b.WriteString(renderProgressBar(30, m.measureCount, len(m.results)))
			b.WriteString(dimStyle.Render(fmt.Sprintf("  %d/%d", m.measureCount, len(m.results))))
			b.WriteByte('\n')
		}
	}
	b.WriteString(dividerStyle.Render(strings.Repeat("─", m.width)))
	b.WriteByte('\n')
}

func (m Model) renderScanFooter(b *strings.Builder) {
	b.WriteString(dividerStyle.Render(strings.Repeat("─", m.width)))
	b.WriteByte('\n')
	b.WriteString(helpStyle.Render("j/k navigate  space toggle  a all  n none  h/l collapse/expand"))
	b.WriteByte('\n')
	b.WriteString(helpStyle.Render("enter apply   r remove      e all exclusions  g group  q quit"))
}

func (m Model) renderExclusionsHeader(b *strings.Builder) {
	b.WriteString(titleStyle.Render("Nepenthe — All System Exclusions"))
	b.WriteByte('\n')

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
	b.WriteByte('\n')
	b.WriteString(dividerStyle.Render(strings.Repeat("─", m.width)))
	b.WriteByte('\n')
}

func (m Model) renderExclusionsFooter(b *strings.Builder) {
	b.WriteString(dividerStyle.Render(strings.Repeat("─", m.width)))
	b.WriteByte('\n')
	b.WriteString(helpStyle.Render("j/k navigate  e back to scan  q quit"))
}
