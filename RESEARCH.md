# Nepenthe - Research & Design Reference

Nepenthe: a drug of forgetfulness from Greek mythology (The Odyssey, Poe's "The Raven"). This tool gives macOS Time Machine selective forgetfulness -- excluding high-churn, reproducible directories that waste backup time and space.

---

## Problem Statement

macOS Time Machine backs up everything by default, including directories that:
- Are fully reproducible from a lockfile or manifest (`node_modules`, `vendor`, `target`, etc.)
- Contain thousands of small files that slow backup I/O (`node_modules` alone can have 50k+ files)
- Are enormous and change constantly (Docker VM images at ~50 GB, Xcode DerivedData at 1+ GB per project)
- Are re-downloadable from the internet (Xcode DeviceSupport, Steam games)

Developers routinely lose hours to slow backups or blow through backup disk space because of these directories.

## How Time Machine Exclusions Work

### tmutil (Apple's built-in CLI, macOS 10.7+)

Two types of exclusions:

1. **Sticky / location-independent** (default, preferred for dependency dirs):
   ```bash
   tmutil addexclusion /path/to/dir
   ```
   - Sets extended attribute `com.apple.metadata:com_apple_backup_excludeItem` on the item
   - Exclusion follows the file/directory if moved or renamed
   - Does NOT require root/sudo
   - Does NOT appear in Time Machine preferences GUI
   - Automatically cleaned up when the directory is deleted

2. **Fixed-path**:
   ```bash
   sudo tmutil addexclusion -p /absolute/path/to/dir
   ```
   - Stored in `/Library/Preferences/com.apple.TimeMachine.plist` under `SkipPaths` array
   - Requires root/sudo
   - DOES appear in Time Machine preferences GUI
   - Excludes whatever is at that path, even if the item changes

### Checking exclusion status

```bash
tmutil isexcluded /path/to/dir
```

### Already auto-excluded by Time Machine

These are in `/System/Library/CoreServices/backupd.bundle/Contents/Resources/StdExclusions.plist` and do NOT need manual exclusion:

- `~/Library/Caches` and `/Library/Caches`
- `.Trashes` / `/.Trashes`
- `.Spotlight-V100`
- `.fseventsd`
- `.DocumentRevisions-V100`
- `/private/var/vm/sleepimage`
- `/Users/Guest`
- `/Library/Logs`

---

## Existing Tools (Competitive Landscape)

### Asimov (stevegrunwell/asimov) - Most Popular
- **Language:** Bash
- **Install:** `brew install asimov`
- **Approach:** Scans filesystem with `find` for known dependency directory names. Only excludes a directory if a corresponding manifest/lockfile exists in the parent directory (sentinel-file pattern). Uses sticky (xattr) exclusions.
- **Scheduling:** launchd daily via `sudo brew services start asimov`
- **Strengths:** Comprehensive hardcoded rule set (31+ patterns), widely adopted, idempotent
- **Weaknesses:** Hardcoded list requires upstream updates for new ecosystems; full filesystem scan can be slow

### tmignore (samuelmeuli/tmignore)
- **Language:** Swift
- **Install:** `brew install samuelmeuli/tap/tmignore`
- **Approach:** Scans for Git repos, reads `.gitignore` files, excludes everything matched by `.gitignore` from Time Machine
- **Commands:** `tmignore run`, `tmignore list`, `tmignore reset`
- **Strengths:** Leverages existing Git conventions -- no hardcoded list needed; covers project-specific patterns automatically
- **Weaknesses:** Only works for Git repos; `.gitignore` patterns may not perfectly map to what should be excluded from backups

### Asimeow (mdnmdn/asimeow)
- **Language:** Rust
- **Install:** `brew tap mdnmdn/asimeow && brew install asimeow` or `cargo install asimeow`
- **Approach:** YAML config with roots, ignore patterns, and rules (glob-based file matching + associated dirs to exclude). Multi-threaded scanning.
- **Strengths:** User-extensible YAML config, fast (Rust + multithreaded)
- **Weaknesses:** Smaller community, less comprehensive default rules

### Heptapod (tg44/heptapod)
- **Language:** Go
- **Install:** `brew tap tg44/heptapod && brew install heptapod`
- **Approach:** YAML rules in `~/.heptapod/rules` with multiple rule types (file-trigger, global, list). Supports dry-run and prune (revert all exclusions).
- **Commands:** `heptapod rules ls -a`, `heptapod run`, `heptapod run --dryrun`, `heptapod prune -a`, `heptapod tm ls`
- **Strengths:** Most flexible rule system, dry-run mode, reversibility via prune
- **Weaknesses:** Smaller community

### tm-exclude (dev01d/tm-exclude)
- **Language:** Shell + launchd plist
- **Approach:** Event-driven via launchd `WatchPaths` -- reacts to filesystem changes in real-time rather than running on a schedule
- **Strengths:** Real-time exclusion as directories are created
- **Weaknesses:** Minimal, primarily targets `node_modules`

### tmexcludes / tmbackup (Neved4/tmexcludes)
- **Language:** POSIX sh (91 lines)
- **Approach:** NOT an auto-scanner. Exports, imports, and syncs the Time Machine exclusion list across machines.
- **Commands:** `list`, `dump`, `load`, `system`
- **Strengths:** Extremely lightweight, solves the "sync exclusions across machines" problem
- **Complementary:** Could pair well with a scanner tool like Nepenthe

### osx-ignorer (mateothegreat/osx-ignorer)
- **Language:** Shell
- **Approach:** Config file with `<filename>:<directory>` patterns. Manages both Time Machine exclusions AND Spotlight indexing.
- **Dual-purpose:** Handles Spotlight too (interesting differentiator)

### Notable: Package Managers Adding Native Support
- **pnpm:** Open PR (#8522) to set `com.apple.metadata:com_apple_backup_excludeItem` xattr on `node_modules` during install
- **npm:** Open discussion (#570) requesting the same
- **Cargo (Rust):** Had a bug (#7317) where it stopped marking `target/` as excluded; now fixed
- **Docker Desktop:** Now auto-sets xattr exclusion on its VM disk image

---

## Comprehensive Exclusion Directory List

### Tier 1: Package Manager Dependencies (Sentinel-File Pattern)

These should ONLY be excluded when the corresponding manifest file exists in the parent directory, to avoid false positives.

| Directory | Sentinel File(s) | Ecosystem |
|-----------|-------------------|-----------|
| `node_modules` | `package.json` | Node.js (npm/yarn/pnpm) |
| `vendor` | `composer.json` | PHP (Composer) |
| `vendor` | `Gemfile` | Ruby (Bundler) |
| `vendor` | `go.mod` | Go Modules |
| `target` | `Cargo.toml` | Rust (Cargo) |
| `target` | `pom.xml` | Java (Maven) |
| `target` | `build.sbt` | Scala (sbt) |
| `.build` | `Package.swift` | Swift SPM |
| `.build` | `mix.exs` | Elixir Mix |
| `.gradle` | `build.gradle` | Gradle |
| `.gradle` | `build.gradle.kts` | Gradle Kotlin DSL |
| `build` | `build.gradle` | Gradle output |
| `build` | `build.gradle.kts` | Gradle Kotlin output |
| `.dart_tool` | `pubspec.yaml` | Dart/Flutter |
| `.packages` | `pubspec.yaml` | Dart (Pub) |
| `.stack-work` | `stack.yaml` | Haskell (Stack) |
| `.tox` | `tox.ini` | Python (Tox) |
| `.nox` | `noxfile.py` | Python (Nox) |
| `.venv` | `requirements.txt` | Python virtualenv |
| `.venv` | `pyproject.toml` | Python virtualenv |
| `venv` | `requirements.txt` | Python virtualenv |
| `dist` | `setup.py` | Python dist |
| `Pods` | `Podfile` | iOS (CocoaPods) |
| `Carthage` | `Cartfile` | iOS (Carthage) |
| `bower_components` | `bower.json` | JavaScript (Bower, legacy) |
| `.parcel-cache` | `package.json` | Parcel v2 |
| `deps` | `mix.exs` | Elixir Mix deps |
| `.vagrant` | `Vagrantfile` | Vagrant |
| `.terraform.d` | `.terraformrc` | Terraform |
| `.terragrunt-cache` | `terragrunt.hcl` | Terragrunt |
| `cdk.out` | `cdk.json` | AWS CDK |

### Tier 2: Xcode / Apple Developer (Fixed-Path)

These are well-known fixed locations. No sentinel detection needed.

| Path | Typical Size | Rationale |
|------|-------------|-----------|
| `~/Library/Developer/Xcode/DerivedData` | 1+ GB per project | Build artifacts, indexes, logs; regenerated on build |
| `~/Library/Developer/Xcode/DocumentationCache` | Varies | Offline docs cache; re-downloaded |
| `~/Library/Developer/Xcode/iOS DeviceSupport` | Multiple GB | Debug symbols; re-downloaded per device |
| `~/Library/Developer/Xcode/tvOS DeviceSupport` | Multiple GB | Same |
| `~/Library/Developer/Xcode/watchOS DeviceSupport` | Multiple GB | Same |
| `~/Library/Developer/CoreSimulator` | Multiple GB | Simulator runtimes and data |
| `~/Library/Developer/XCPGDevices` | Varies | Playground device data |
| `~/Library/Developer/XCTestDevices` | Varies | Test device configs |
| `~/Library/Logs/CoreSimulator` | Varies | Simulator logs |

### Tier 3: Docker & Containers (Fixed-Path)

| Path | Typical Size | Rationale |
|------|-------------|-----------|
| `~/Library/Containers/com.docker.docker/Data` | ~50 GB | Docker Desktop VM disk image; high churn |
| `~/.docker` | ~20 GB | Docker Machine data; images re-pullable |

### Tier 4: Virtual Machines (Fixed-Path)

| Path | Typical Size | Rationale |
|------|-------------|-----------|
| `~/Library/Parallels/` | 20+ GB per VM | Monolithic disk images change completely on each use |
| `~/Documents/Virtual Machines/` | 20+ GB per VM | VMware Fusion default location |

### Tier 5: General / Optional (User-Opt-In)

These are more opinionated and should probably be opt-in rather than automatic.

| Path | Rationale |
|------|-----------|
| `~/Downloads` | Temporary staging area |
| `/Applications/Xcode.app` | 12-35 GB, re-downloadable |
| `~/Library/Application Support/Steam/steamapps` | Re-downloadable games |
| Cloud sync folders (Dropbox, Google Drive, OneDrive) | Already backed up to cloud |
| Homebrew Cellar/Caskroom (`/opt/homebrew/` or `/usr/local/`) | Reproducible from Brewfile |

---

## Key Design Insights from Existing Tools

1. **Sentinel-file validation is critical.** Asimov's core insight: don't exclude a directory named `vendor` or `build` unless a corresponding manifest file (`composer.json`, `build.gradle`, etc.) exists in the parent. This prevents false positives on unrelated directories.

2. **Sticky (xattr) exclusions are preferred** for dependency directories because they auto-clean when the directory is deleted and don't require sudo. Fixed-path exclusions are better for well-known system locations (Xcode, Docker).

3. **Idempotency matters.** Running the tool multiple times should be safe. Check `tmutil isexcluded` before adding.

4. **Reversibility builds trust.** Heptapod's `prune` command and tmignore's `reset` command let users undo everything the tool has done. This is important for adoption.

5. **Dry-run mode is essential.** Users want to see what will be excluded before it happens.

6. **Scheduling via launchd** is the standard approach. Daily runs catch new projects. Event-driven (WatchPaths) is more responsive but more complex.

7. **File count matters as much as total size.** `node_modules` may only be 500 MB but contain 50,000+ files. Time Machine performance degrades with many small files because each requires separate I/O.

8. **Docker is often the single largest offender.** 50-70+ GB can dwarf all other exclusions combined.

---

## Potential Differentiators for Nepenthe

Based on gaps in the existing landscape:

- **Tiered exclusions:** Separate "definitely exclude" (dependencies with sentinels) from "probably exclude" (Xcode, Docker) from "maybe exclude" (Downloads, Steam) with different levels of user consent
- **Dry-run with impact estimation:** Show estimated size savings before applying, not just a list of paths
- **Combined Spotlight + Time Machine management** (only osx-ignorer does this currently)
- **Export/import of exclusion configs** (only tmexcludes does this; could be built-in)
- **Status dashboard:** Show current exclusion state, total space excluded, directories that have appeared since last run
- **Watch mode:** Real-time exclusion of new dependency directories as they're created (only tm-exclude does this, and minimally)
- **.nepenthe config file in project roots** for project-specific exclusion rules (similar to how .gitignore works per-repo)

---

## Sources

- [stevegrunwell/asimov](https://github.com/stevegrunwell/asimov)
- [samuelmeuli/tmignore](https://github.com/samuelmeuli/tmignore)
- [mdnmdn/asimeow](https://github.com/mdnmdn/asimeow)
- [tg44/heptapod](https://github.com/tg44/heptapod)
- [dev01d/tm-exclude](https://github.com/dev01d/tm-exclude)
- [Neved4/tmexcludes](https://github.com/Neved4/tmexcludes)
- [mateothegreat/osx-ignorer](https://github.com/mateothegreat/osx-ignorer)
- [alexwlchan - tmutil exclusions](https://alexwlchan.net/til/2024/exclude-files-from-time-machine-with-tmutil/)
- [Eclectic Light Company - Excluding folders and files](https://eclecticlight.co/2024/07/09/excluding-folders-and-files-from-time-machine-spotlight-and-icloud-drive/)
- [Eclectic Light Company - Knowing what not to back up](https://eclecticlight.co/2021/08/03/knowing-what-not-to-back-up-and-how/)
- [Steve Grunwell - Exclude Dependencies from Time Machine](https://stevegrunwell.com/blog/time-machine-exclude-dependencies/)
- [/dev/trouble - Folders to exclude](https://tredje.se/dev/trouble/post/folders-to-exclude-from-time-machine-backups)
- [tekkie.dev - Prevent large Time Machine backups by Docker](https://tekkie.dev/docker/prevent-large-time-machine-backups-by-docker)
- [How-To Geek - Save Space on Your Time Machine Drive](https://www.howtogeek.com/294600/save-space-on-your-time-machine-drive-by-excluding-these-folders-from-backups/)
- [myByways - Exclude folders to speed up Time Machine](https://mybyways.com/blog/exclude-folders-to-speed-up-time-machine-backup)
- [Essential Apple - Folders To Exclude](https://essentialapple.com/apple-mac-iphone-how-to/what-folders-to-exclude-from-time-machine-backups/)
- [ProgrammingAreHard - Time Machine for Developers](https://programmingarehard.com/2022/03/03/time-machine-for-developers.html/)
- [pnpm PR #8522](https://github.com/pnpm/pnpm/pull/8522)
- [npm feedback #570](https://github.com/npm/feedback/discussions/570)
- [Cargo issue #7317](https://github.com/rust-lang/cargo/issues/7317)
- [Docker roadmap #586](https://github.com/docker/roadmap/issues/586)
