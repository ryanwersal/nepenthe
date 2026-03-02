package scanner

import (
	"fmt"
	"os"
)

// BuildSentinelRules returns the built-in sentinel rules.
func BuildSentinelRules() []SentinelRule {
	rules := make([]SentinelRule, len(sentinelRules))
	copy(rules, sentinelRules)
	return rules
}

var sentinelRules = []SentinelRule{
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

func FixedPathCategories() (map[Category][]FixedPathRule, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolving home directory: %w", err)
	}
	return map[Category][]FixedPathRule{
		CategoryDevCaches: {
			{Path: home + "/Library/Developer/Xcode/DerivedData", Ecosystem: "Xcode", Description: "Xcode derived data", Category: CategoryDevCaches},
			{Path: home + "/Library/Developer/Xcode/DocumentationCache", Ecosystem: "Xcode", Description: "Xcode documentation cache", Category: CategoryDevCaches},
			{Path: home + "/Library/Developer/Xcode/iOS DeviceSupport", Ecosystem: "Xcode", Description: "iOS device support", Category: CategoryDevCaches},
			{Path: home + "/Library/Developer/Xcode/tvOS DeviceSupport", Ecosystem: "Xcode", Description: "tvOS device support", Category: CategoryDevCaches},
			{Path: home + "/Library/Developer/Xcode/watchOS DeviceSupport", Ecosystem: "Xcode", Description: "watchOS device support", Category: CategoryDevCaches},
			{Path: home + "/Library/Developer/CoreSimulator", Ecosystem: "Xcode", Description: "Core Simulator", Category: CategoryDevCaches},
			{Path: home + "/Library/Developer/XCPGDevices", Ecosystem: "Xcode", Description: "XCPGDevices", Category: CategoryDevCaches},
			{Path: home + "/Library/Developer/XCTestDevices", Ecosystem: "Xcode", Description: "XCTestDevices", Category: CategoryDevCaches},
			{Path: home + "/Library/Logs/CoreSimulator", Ecosystem: "Xcode", Description: "Core Simulator logs", Category: CategoryDevCaches},
		},
		CategoryContainers: {
			{Path: home + "/Library/Containers/com.docker.docker/Data", Ecosystem: "Docker", Description: "Docker data", Category: CategoryContainers},
			{Path: home + "/.docker", Ecosystem: "Docker", Description: "Docker config", Category: CategoryContainers},
		},
		CategoryVMs: {
			{Path: home + "/Library/Parallels", Ecosystem: "VMs", Description: "Parallels VMs", Category: CategoryVMs},
			{Path: home + "/Documents/Virtual Machines", Ecosystem: "VMs", Description: "Virtual machines", Category: CategoryVMs},
		},
		CategoryOptional: {
			{Path: home + "/Downloads", Ecosystem: "Optional", Description: "Downloads folder", Category: CategoryOptional},
			{Path: "/Applications/Xcode.app", Ecosystem: "Optional", Description: "Xcode application", Category: CategoryOptional},
			{Path: home + "/Library/Application Support/Steam/steamapps", Ecosystem: "Optional", Description: "Steam games", Category: CategoryOptional},
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
