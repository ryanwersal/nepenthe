package scanner

type ScanResult struct {
	Path       string
	Ecosystem  string
	Tier       int
	Type       string // "sticky" or "fixed"
	IsExcluded bool
	SizeBytes  int64
	FileCount  int64
}

type SentinelRule struct {
	Directory string
	Sentinels []string
	Ecosystem string
}

type FixedPathRule struct {
	Path        string
	Ecosystem   string
	Description string
	Tier        int
}

type WalkOptions struct {
	Roots       []string
	Rules       []SentinelRule
	Concurrency int
	OnFound     func(ScanResult)
}
