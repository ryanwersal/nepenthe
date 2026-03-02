package tui

import (
	"github.com/ryanwersal/nepenthe/internal/config"
	"github.com/ryanwersal/nepenthe/internal/scanner"
)

type ScanResultMsg struct {
	Result scanner.ScanResult
}

type ScanDoneMsg struct{}

type ApplyProgressMsg struct {
	Index   int
	Success bool
	Done    int
	Total   int
}

type ApplyDoneMsg struct {
	Applied int
	Failed  int
}

type RemoveProgressMsg struct {
	Index   int
	Success bool
	Done    int
	Total   int
}

type RemoveDoneMsg struct {
	Removed int
	Failed  int
}

type MeasureDoneMsg struct {
	Results []scanner.ScanResult
}

type SizeMeasuredMsg struct {
	Index     int
	SizeBytes int64
	FileCount int64
}

type AllSizesDoneMsg struct{}

type AllExclusionsMsg struct {
	Paths []string
}

type ConfigUpdatedMsg struct {
	Cfg config.Config
	Err error
}

type ErrorMsg struct {
	Err error
}
