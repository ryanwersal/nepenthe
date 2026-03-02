package scanner

type Category string

const (
	CategoryCustom       Category = "custom"
	CategoryDependencies Category = "dependencies"
	CategoryDevCaches    Category = "dev-caches"
	CategoryContainers   Category = "containers"
	CategoryVMs          Category = "vms"
	CategoryOptional     Category = "optional"
)

var AllCategories = []Category{
	CategoryCustom,
	CategoryDependencies,
	CategoryDevCaches,
	CategoryContainers,
	CategoryVMs,
	CategoryOptional,
}

var CategoryLabel = map[Category]string{
	CategoryCustom:       "Custom paths",
	CategoryDependencies: "Build dependencies",
	CategoryDevCaches:    "Developer caches",
	CategoryContainers:   "Containers",
	CategoryVMs:          "Virtual machines",
	CategoryOptional:     "Optional",
}

type ScanResult struct {
	Path       string
	Ecosystem  string
	Category   Category
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
	Category    Category
}

type WalkOptions struct {
	Roots   []string
	Rules   []SentinelRule
	OnFound func(ScanResult)
}
