package scanner

import (
	"fmt"
	"os"
)

// BuildSentinelRules returns the built-in sentinel rules merged with any custom rules.
func BuildSentinelRules(custom []SentinelRule) []SentinelRule {
	rules := make([]SentinelRule, len(SentinelRules), len(SentinelRules)+len(custom))
	copy(rules, SentinelRules)
	return append(rules, custom...)
}

var SentinelRules = []SentinelRule{
	{Directory: "node_modules", Sentinels: []string{"package.json"}, Ecosystem: "Node.js"},
	{Directory: ".parcel-cache", Sentinels: []string{"package.json"}, Ecosystem: "Node.js (Parcel)"},
	{Directory: "vendor", Sentinels: []string{"composer.json"}, Ecosystem: "PHP (Composer)"},
	{Directory: "vendor", Sentinels: []string{"Gemfile"}, Ecosystem: "Ruby (Bundler)"},
	{Directory: "vendor", Sentinels: []string{"go.mod"}, Ecosystem: "Go"},
	{Directory: "target", Sentinels: []string{"Cargo.toml"}, Ecosystem: "Rust (Cargo)"},
	{Directory: "target", Sentinels: []string{"pom.xml"}, Ecosystem: "Java (Maven)"},
	{Directory: "target", Sentinels: []string{"build.sbt"}, Ecosystem: "Scala (sbt)"},
	{Directory: ".gradle", Sentinels: []string{"build.gradle", "build.gradle.kts"}, Ecosystem: "Gradle"},
	{Directory: "build", Sentinels: []string{"build.gradle", "build.gradle.kts"}, Ecosystem: "Gradle"},
	{Directory: ".build", Sentinels: []string{"Package.swift"}, Ecosystem: "Swift (SPM)"},
	{Directory: ".build", Sentinels: []string{"mix.exs"}, Ecosystem: "Elixir (Mix)"},
	{Directory: "deps", Sentinels: []string{"mix.exs"}, Ecosystem: "Elixir (Mix)"},
	{Directory: ".dart_tool", Sentinels: []string{"pubspec.yaml"}, Ecosystem: "Dart/Flutter"},
	{Directory: ".packages", Sentinels: []string{"pubspec.yaml"}, Ecosystem: "Dart (Pub)"},
	{Directory: ".stack-work", Sentinels: []string{"stack.yaml"}, Ecosystem: "Haskell (Stack)"},
	{Directory: ".tox", Sentinels: []string{"tox.ini"}, Ecosystem: "Python (Tox)"},
	{Directory: ".nox", Sentinels: []string{"noxfile.py"}, Ecosystem: "Python (Nox)"},
	{Directory: ".venv", Sentinels: []string{"requirements.txt", "pyproject.toml"}, Ecosystem: "Python"},
	{Directory: "venv", Sentinels: []string{"requirements.txt", "pyproject.toml"}, Ecosystem: "Python"},
	{Directory: "dist", Sentinels: []string{"setup.py"}, Ecosystem: "Python"},
	{Directory: "Pods", Sentinels: []string{"Podfile"}, Ecosystem: "iOS (CocoaPods)"},
	{Directory: "Carthage", Sentinels: []string{"Cartfile"}, Ecosystem: "iOS (Carthage)"},
	{Directory: "bower_components", Sentinels: []string{"bower.json"}, Ecosystem: "JavaScript (Bower)"},
	{Directory: ".vagrant", Sentinels: []string{"Vagrantfile"}, Ecosystem: "Vagrant"},
	{Directory: ".terraform.d", Sentinels: []string{".terraformrc"}, Ecosystem: "Terraform"},
	{Directory: ".terragrunt-cache", Sentinels: []string{"terragrunt.hcl"}, Ecosystem: "Terragrunt"},
	{Directory: "cdk.out", Sentinels: []string{"cdk.json"}, Ecosystem: "AWS CDK"},
}

func FixedPathTiers() (map[int][]FixedPathRule, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolving home directory: %w", err)
	}
	return map[int][]FixedPathRule{
		2: {
			{Path: home + "/Library/Developer/Xcode/DerivedData", Ecosystem: "Xcode", Description: "Xcode derived data", Tier: 2},
			{Path: home + "/Library/Developer/Xcode/DocumentationCache", Ecosystem: "Xcode", Description: "Xcode documentation cache", Tier: 2},
			{Path: home + "/Library/Developer/Xcode/iOS DeviceSupport", Ecosystem: "Xcode", Description: "iOS device support", Tier: 2},
			{Path: home + "/Library/Developer/Xcode/tvOS DeviceSupport", Ecosystem: "Xcode", Description: "tvOS device support", Tier: 2},
			{Path: home + "/Library/Developer/Xcode/watchOS DeviceSupport", Ecosystem: "Xcode", Description: "watchOS device support", Tier: 2},
			{Path: home + "/Library/Developer/CoreSimulator", Ecosystem: "Xcode", Description: "Core Simulator", Tier: 2},
			{Path: home + "/Library/Developer/XCPGDevices", Ecosystem: "Xcode", Description: "XCPGDevices", Tier: 2},
			{Path: home + "/Library/Developer/XCTestDevices", Ecosystem: "Xcode", Description: "XCTestDevices", Tier: 2},
			{Path: home + "/Library/Logs/CoreSimulator", Ecosystem: "Xcode", Description: "Core Simulator logs", Tier: 2},
		},
		3: {
			{Path: home + "/Library/Containers/com.docker.docker/Data", Ecosystem: "Docker", Description: "Docker data", Tier: 3},
			{Path: home + "/.docker", Ecosystem: "Docker", Description: "Docker config", Tier: 3},
		},
		4: {
			{Path: home + "/Library/Parallels", Ecosystem: "VMs", Description: "Parallels VMs", Tier: 4},
			{Path: home + "/Documents/Virtual Machines", Ecosystem: "VMs", Description: "Virtual machines", Tier: 4},
		},
		5: {
			{Path: home + "/Downloads", Ecosystem: "Optional", Description: "Downloads folder", Tier: 5},
			{Path: "/Applications/Xcode.app", Ecosystem: "Optional", Description: "Xcode application", Tier: 5},
			{Path: home + "/Library/Application Support/Steam/steamapps", Ecosystem: "Optional", Description: "Steam games", Tier: 5},
		},
	}, nil
}

var PruneDirs = map[string]bool{
	"node_modules":       true,
	"vendor":             true,
	"target":             true,
	".build":             true,
	".gradle":            true,
	"Pods":               true,
	"Carthage":           true,
	"bower_components":   true,
	".dart_tool":         true,
	".packages":          true,
	".stack-work":        true,
	".tox":               true,
	".nox":               true,
	".venv":              true,
	"venv":               true,
	".vagrant":           true,
	".terraform.d":       true,
	".terragrunt-cache":  true,
	"cdk.out":            true,
	".parcel-cache":      true,
	"deps":               true,
	".git":               true,
	".svn":               true,
	".hg":                true,
	".Trash":             true,
	".Spotlight-V100":    true,
	".fseventsd":         true,
}
