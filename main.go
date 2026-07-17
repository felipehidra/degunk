package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CleanTarget represents a location on disk that can be scanned and cleaned
type CleanTarget struct {
	Name        string
	Path        string // supports ~ expansion
	Description string
	SafeToWipe  bool // true = contents are fully regenerable (caches); false = review before delete
}

// FileEntry holds info about a file/folder candidate for removal
type FileEntry struct {
	Path    string
	Size    int64
	IsDir   bool
	Target  string // which CleanTarget it came from
}

// targets lists well-known macOS locations that commonly accumulate junk.
// SafeToWipe=true means the OS/apps will recreate the data on next run.
var targets = []CleanTarget{
	{"User Caches", "~/Library/Caches", "App caches – safe to clear, will be regenerated", true},
	{"User Logs", "~/Library/Logs", "Application logs – safe to clear", true},
	{"System Caches", "/Library/Caches", "System-wide caches – safe to clear", true},
	{"Xcode DerivedData", "~/Library/Developer/Xcode/DerivedData", "Xcode build artifacts – rebuilt automatically", true},
	{"Xcode Archives", "~/Library/Developer/Xcode/Archives", "Old app archives – review before deleting", false},
	{"Xcode iOS DeviceSupport", "~/Library/Developer/Xcode/iOS DeviceSupport", "Symbol files for old iOS versions", true},
	{"Xcode Simulators Caches", "~/Library/Developer/CoreSimulator/Caches", "Simulator caches", true},
	{"iOS Device Backups", "~/Library/Application Support/MobileSync/Backup", "iPhone/iPad backups – REVIEW before deleting", false},
	{"Trash", "~/.Trash", "Files in Trash", true},
	{"Downloads", "~/Downloads", "Downloaded files – REVIEW before deleting", false},
	{"npm cache", "~/.npm", "Node.js npm cache", true},
	{"Yarn cache", "~/.yarn/cache", "Yarn package cache", true},
	{"pip cache", "~/Library/Caches/pip", "Python pip cache", true},
	{"Homebrew cache", "~/Library/Caches/Homebrew", "Homebrew downloads", true},
	{"Go build cache", "~/Library/Caches/go-build", "Go compiler build cache", true},
	{"Docker", "~/Library/Containers/com.docker.docker/Data/vms", "Docker VM data – use 'docker system prune' instead", false},
	{"Slack cache", "~/Library/Application Support/Slack/Cache", "Slack app cache", true},
	{"Chrome cache", "~/Library/Caches/Google/Chrome", "Chrome browser cache", true},
	{"Safari cache", "~/Library/Caches/com.apple.Safari", "Safari browser cache", true},
	{"Firefox cache", "~/Library/Caches/Firefox", "Firefox browser cache", true},
	{"Edge cache", "~/Library/Caches/com.microsoft.edgemac", "Microsoft Edge browser cache", true},
	{"Brave cache", "~/Library/Caches/BraveSoftware", "Brave browser cache", true},
	{"Arc cache", "~/Library/Caches/company.thebrowser.Browser", "Arc browser cache", true},
	{"Discord cache", "~/Library/Application Support/discord/Cache", "Discord app cache (Cache subdir only; parent holds login state)", true},
	{"Teams cache", "~/Library/Containers/com.microsoft.teams2/Data/Library/Caches", "Microsoft Teams cache", true},
	{"Zoom cache", "~/Library/Caches/us.zoom.xos", "Zoom cache (recordings live elsewhere and are not touched)", true},
	{"WhatsApp cache", "~/Library/Caches/net.whatsapp.WhatsApp", "WhatsApp cache", true},
	{"Telegram cache", "~/Library/Caches/ru.keepcoder.Telegram", "Telegram cache", true},
	{"Spotify cache", "~/Library/Caches/com.spotify.client", "Spotify offline cache", true},
	{"Cursor cache", "~/Library/Application Support/Cursor/Cache", "Cursor editor cache", true},
	{"JetBrains caches", "~/Library/Caches/JetBrains", "JetBrains IDE caches (IntelliJ / GoLand / PyCharm / etc.)", true},
	{"Gradle cache", "~/.gradle/caches", "Gradle dependency and build cache", true},
	{"Cargo registry", "~/.cargo/registry", "Rust crate registry cache (NOT ~/.cargo – bin/ has installed binaries)", true},
	{"Cargo git cache", "~/.cargo/git", "Rust git-dependency cache", true},
	{"pnpm store", "~/Library/pnpm/store", "pnpm global package store", true},
	{"CocoaPods cache", "~/Library/Caches/CocoaPods", "CocoaPods download cache", true},
	{"Flutter pub cache", "~/.pub-cache", "Flutter/Dart pub package cache", true},
	{"iPhone Software Updates", "~/Library/iTunes/iPhone Software Updates", "Downloaded .ipsw files for iPhone", true},
	{"iPad Software Updates", "~/Library/iTunes/iPad Software Updates", "Downloaded .ipsw files for iPad", true},
	{"Xcode Simulator Devices", "~/Library/Developer/CoreSimulator/Devices", "Installed simulator devices – REVIEW before deleting (each subdir is a full simulator)", false},
	{"Xcode SwiftUI Previews", "~/Library/Developer/Xcode/UserData/Previews", "SwiftUI preview snapshots – rebuilt by Xcode", true},
	{"Xcode IDE cache", "~/Library/Caches/com.apple.dt.Xcode", "Xcode IDE caches – rebuilt automatically", true},
	{"Simulator dyld caches", "/Library/Developer/CoreSimulator/Caches/dyld", "Simulator runtime dyld shared caches – rebuilt by Xcode on next simulator boot", true},
	{"Simulator runtimes", "/Library/Developer/CoreSimulator/Profiles/Runtimes", "Simulator OS runtimes – REVIEW (each is 5-15 GB; deleting removes that OS from simulators)", false},
	{"Simulator device logs", "~/Library/Logs/CoreSimulator", "Log files from simulator runs – safe to clear", true},
	{"Swift Package Manager cache", "~/Library/Caches/org.swift.swiftpm", "SPM dependency cache – rebuilt on next build", true},
	{"Xcode DocSets", "~/Library/Developer/Shared/Documentation/DocSets", "Downloaded Xcode offline documentation sets – REVIEW (must re-download)", false},
	{"Diagnostic reports", "~/Library/Logs/DiagnosticReports", "App crash/hang reports – safe to clear", true},
	{"Quick Look cache", "~/Library/Caches/com.apple.QuickLook", "Quick Look thumbnail cache – rebuilt on demand", true},
	{"Adobe media cache", "~/Library/Application Support/Adobe/Common/Media Cache", "Adobe Premiere/After Effects media cache – safe to clear, regenerated on playback", true},
	{"Android Studio caches", "~/Library/Caches/Google", "Android Studio and other Google app caches", true},
	{"Android AVD images", "~/.android/avd", "Android emulator disk images – REVIEW (re-creating an AVD takes time)", false},
	{"rbenv versions", "~/.rbenv/versions", "Ruby versions installed via rbenv – REVIEW (removes those Ruby installs)", false},
	{"Zoom recordings", "~/Documents/Zoom", "Zoom meeting recordings – REVIEW", false},
	{"VS Code cache", "~/Library/Application Support/Code/Cache", "VS Code disk cache", true},
	{"VS Code CachedData", "~/Library/Application Support/Code/CachedData", "VS Code extension/index cache", true},
	{"Cursor CachedData", "~/Library/Application Support/Cursor/CachedData", "Cursor extension/index cache", true},
	{"Chrome Service Worker", "~/Library/Application Support/Google/Chrome/Default/Service Worker", "Chrome service-worker / IndexedDB cache – sites refetch", true},
	{"Spotify offline cache", "~/Library/Application Support/Spotify/PersistentCache", "Spotify Premium offline-song cache", true},
	{"Gradle wrapper dists", "~/.gradle/wrapper/dists", "Gradle wrapper distributions (separate from caches/)", true},
	{"Maven local repo", "~/.m2/repository", "Maven local artifact repository", true},
	{"Hugging Face cache", "~/.cache/huggingface", "Hugging Face model/dataset cache", true},
	{"PyTorch hub cache", "~/.cache/torch/hub", "PyTorch hub model cache", true},
	{"Yarn Berry cache", "~/.yarn/berry/cache", "Yarn 2+ (Berry) package cache", true},
	{"Bun install cache", "~/.bun/install/cache", "Bun JS runtime install cache", true},
	{"Deno cache", "~/Library/Caches/deno", "Deno runtime cache", true},
	{"Composer cache", "~/.composer/cache", "PHP Composer package cache", true},
	{"Bundler cache", "~/.bundle/cache", "Ruby Bundler gem cache", true},
	{"Arc Service Worker", "~/Library/Application Support/Arc/User Data/Default/Service Worker", "Arc service-worker / IndexedDB cache – sites refetch", true},
	{"Brave Service Worker", "~/Library/Application Support/BraveSoftware/Brave-Browser/Default/Service Worker", "Brave service-worker / IndexedDB cache – sites refetch", true},
	{"Edge Service Worker", "~/Library/Application Support/Microsoft Edge/Default/Service Worker", "Edge service-worker / IndexedDB cache – sites refetch", true},
	{"pyenv versions", "~/.pyenv/versions", "Installed Python versions – REVIEW (clearing uninstalls active Python)", false},
	{"nvm Node versions", "~/.nvm/versions/node", "Installed Node versions via nvm – REVIEW", false},
	{"rustup toolchains", "~/.rustup/toolchains", "Installed Rust toolchains – REVIEW", false},
	{"Miniconda packages", "~/miniconda3/pkgs", "Conda package cache – REVIEW (some files hard-linked into envs)", false},
	{"Anaconda packages", "~/anaconda3/pkgs", "Conda package cache – REVIEW (some files hard-linked into envs)", false},
	{"OrbStack data", "~/.orbstack/data", "OrbStack VM/container data – REVIEW", false},
	{"colima config", "~/.colima", "colima config and runtime – REVIEW", false},
	{"Lima VM disks", "~/.lima", "Lima VM disk images (used by colima/etc.) – REVIEW", false},
	{"Podman containers", "~/.local/share/containers", "Podman container/image storage – REVIEW", false},
	{"Vagrant boxes", "~/.vagrant.d/boxes", "Vagrant downloaded box images – REVIEW", false},
	{"Mail Downloads", "~/Library/Containers/com.apple.mail/Data/Library/Mail Downloads", "Local email attachments – REVIEW (lose access if message expires)", false},
	{"Ollama models", "~/.ollama/models", "Local LLM model files – REVIEW (re-pull is slow)", false},
	{"ccache (C/C++)", "~/.ccache", "ccache C/C++ compiler cache", true},
	{"Bazelisk cache", "~/.cache/bazelisk", "Bazelisk version-managed Bazel cache", true},
	{"Cabal store", "~/.cabal/store", "Haskell Cabal compiled-library store – REVIEW", false},
	{"Stack data", "~/.stack", "Haskell Stack toolchains and snapshots – REVIEW", false},
	{"GHCup cache", "~/.ghcup/cache", "GHCup (Haskell installer) cache", true},
	{"opam download cache", "~/.opam/download-cache", "OCaml opam package download cache", true},
	{"Hex packages", "~/.hex/packages", "Erlang/Elixir Hex package cache", true},
	{"Mix archives", "~/.mix/archives", "Elixir Mix archives cache", true},
	{"Apple Books downloads", "~/Library/Containers/com.apple.iBooksX/Data/Library/Caches/com.apple.iBooksX/Books", "Apple Books downloaded books – REVIEW", false},
	{"Apple TV downloads", "~/Library/Containers/com.apple.TV/Data/Library/Application Support/com.apple.tv/private/Downloads", "Apple TV downloaded shows – REVIEW", false},
	{"uv cache", "~/.cache/uv", "Python uv package cache", true},
	{"Playwright browsers", "~/Library/Caches/ms-playwright", "Playwright downloaded browser builds – re-downloaded on next install", true},
	{"Puppeteer browsers", "~/.cache/puppeteer", "Puppeteer downloaded Chromium builds", true},
	{"Electron cache", "~/Library/Caches/electron", "Electron framework download cache", true},
	{"electron-builder cache", "~/Library/Caches/electron-builder", "electron-builder packaging cache", true},
	{"Go module cache", "~/go/pkg/mod", "Go module download cache – re-downloaded on next build", true},
	{"Xcode test devices", "~/Library/Developer/XCTestDevices", "Simulator clones used by XCTest – recreated on next test run", true},
	{"watchOS DeviceSupport", "~/Library/Developer/Xcode/watchOS DeviceSupport", "Symbol files for old watchOS versions", true},
	{"tvOS DeviceSupport", "~/Library/Developer/Xcode/tvOS DeviceSupport", "Symbol files for old tvOS versions", true},
	{"Unity cache", "~/Library/Unity/cache", "Unity Asset Store and package cache", true},
	{"Terraform plugin cache", "~/.terraform.d/plugin-cache", "Terraform provider plugin cache", true},
	{"CocoaPods spec repos", "~/.cocoapods/repos", "CocoaPods podspec repos – re-cloned on next pod install", true},
	{"Android SDK system images", "~/Library/Android/sdk/system-images", "Android emulator OS images – REVIEW (re-download per image is slow)", false},
	{"Steam shader cache", "~/Library/Application Support/Steam/steamapps/shadercache", "Steam per-game shader caches – rebuilt while playing", true},
	{"Podcasts episodes", "~/Library/Group Containers/243LU875E5.groups.com.apple.podcasts/Library/Cache", "Apple Podcasts downloaded episodes – re-downloadable", true},
	{"Dropbox cache", "~/Dropbox/.dropbox.cache", "Dropbox sync cache", true},
	{"Steam games", "~/Library/Application Support/Steam/steamapps/common", "Installed Steam games – REVIEW (each subdir is a full game install)", false},
	{"WhatsApp media", "~/Library/Group Containers/group.net.whatsapp.WhatsApp.shared/Message/Media", "WhatsApp chat photos/videos – REVIEW (not regenerable)", false},
	{"Google Drive cache", "~/Library/Application Support/Google/DriveFS", "Google Drive local content cache – REVIEW (offline files live here)", false},
	{"GarageBand sound library", "/Library/Application Support/GarageBand", "GarageBand instruments/loops – REVIEW (re-download via GarageBand)", false},
}

// expandHome replaces a leading ~ with the user's home directory.
func expandHome(p string) string {
	if strings.HasPrefix(p, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		return filepath.Join(home, p[1:])
	}
	return p
}

// dirSize walks a directory and returns its total size in bytes.
// Permission errors are skipped silently so one unreadable subtree
// doesn't abort the whole scan.
func dirSize(path string) (int64, error) {
	var total int64
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total, err
}

// humanSize formats bytes into a human-readable string (KB, MB, GB, ...).
func humanSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// truncLeft shortens s to at most maxLen runes, keeping the rightmost portion
// and prepending a leading ellipsis. Returns s unchanged if it already fits.
func truncLeft(s string, maxLen int) string {
	if maxLen <= 1 {
		return "…"
	}
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return "…" + string(r[len(r)-(maxLen-1):])
}

// scanTargets walks each known target location and returns the largest
// top-level entries across all of them, sorted descending by size.
// minSize filters out tiny entries (in bytes).
func scanTargets(minSize int64) ([]FileEntry, map[string]bool) {
	var entries []FileEntry

	// Build a set of expanded target paths so we can avoid double-listing a
	// child that is itself a dedicated target (e.g. `~/Library/Caches/Homebrew`
	// shouldn't appear under User Caches because the Homebrew cache target
	// already scans it for granular children).
	targetPaths := make(map[string]bool, len(targets))
	for _, t := range targets {
		targetPaths[filepath.Clean(expandHome(t.Path))] = true
	}

	for _, t := range targets {
		path := expandHome(t.Path)
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			continue
		}

		children, err := os.ReadDir(path)
		if err != nil {
			continue
		}

		// If target has few children, also include the target itself as a summary
		var targetTotal int64
		for _, c := range children {
			childPath := filepath.Join(path, c.Name())
			var size int64
			if c.IsDir() {
				size, _ = dirSize(childPath)
			} else {
				fi, err := c.Info()
				if err == nil {
					size = fi.Size()
				}
			}
			targetTotal += size

			// Skip listing this child if another target covers it directly —
			// its dedicated target will scan its grandchildren. The size still
			// contributes to this target's total, so the per-target summary
			// remains accurate.
			if targetPaths[filepath.Clean(childPath)] {
				continue
			}

			if size >= minSize {
				entries = append(entries, FileEntry{
					Path:   childPath,
					Size:   size,
					IsDir:  c.IsDir(),
					Target: t.Name,
				})
			}
		}

		// Print a per-target summary line so the user sees totals even if
		// individual children are below the threshold.
		if targetTotal > 0 {
			fmt.Printf("  [%s] total: %s  (%s)\n",
				t.Name, humanSize(targetTotal), path)
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Size > entries[j].Size
	})
	return entries, targetPaths
}

// containsTarget reports whether path is itself a curated target OR is an
// ancestor of one (so its subtree is already partially handled by curated
// scanning). Used by scanHome to avoid surfacing parents of curated paths.
func containsTarget(path string, skip map[string]bool) bool {
	cleanPath := filepath.Clean(path)
	if skip[cleanPath] {
		return true
	}
	prefix := cleanPath + string(os.PathSeparator)
	for k := range skip {
		if strings.HasPrefix(k, prefix) {
			return true
		}
	}
	return false
}

// scanHome walks the user's home directory for large files/dirs not already
// covered by the curated targets. Skips Library and .Trash (handled by
// targets) and any path that IS a curated target or contains one. For
// top-level dirs above threshold, descends one level so the user gets
// actionable grand-children rather than a single coarse-grained
// "~/Documents" entry.
func scanHome(threshold int64, skip map[string]bool) []FileEntry {
	var entries []FileEntry

	home, err := os.UserHomeDir()
	if err != nil {
		return entries
	}

	children, err := os.ReadDir(home)
	if err != nil {
		return entries
	}

	for _, c := range children {
		name := c.Name()
		if name == "Library" || name == ".Trash" {
			continue
		}
		childPath := filepath.Join(home, name)
		if containsTarget(childPath, skip) {
			continue
		}

		var size int64
		if c.IsDir() {
			size, _ = dirSize(childPath)
		} else {
			fi, err := c.Info()
			if err == nil {
				size = fi.Size()
			}
		}
		if size < threshold {
			continue
		}

		// For dirs above threshold, surface granular grand-children if any
		// of them are themselves >= threshold/2. Otherwise emit the dir.
		if c.IsDir() {
			if sub := scanHomeOneLevel(childPath, threshold/2, skip); len(sub) > 0 {
				entries = append(entries, sub...)
				fmt.Printf("  [Home scan] descended into %s (%s, %d sub-entries)\n",
					childPath, humanSize(size), len(sub))
				continue
			}
		}

		entries = append(entries, FileEntry{
			Path:   childPath,
			Size:   size,
			IsDir:  c.IsDir(),
			Target: "Home (scan)",
		})
		fmt.Printf("  [Home scan] found %s (%s)\n", childPath, humanSize(size))
	}

	return entries
}

// scanHomeOneLevel returns direct children of `parent` whose size is >= threshold.
func scanHomeOneLevel(parent string, threshold int64, skip map[string]bool) []FileEntry {
	var entries []FileEntry
	children, err := os.ReadDir(parent)
	if err != nil {
		return entries
	}
	for _, c := range children {
		childPath := filepath.Join(parent, c.Name())
		if containsTarget(childPath, skip) {
			continue
		}
		var size int64
		if c.IsDir() {
			size, _ = dirSize(childPath)
		} else {
			fi, err := c.Info()
			if err == nil {
				size = fi.Size()
			}
		}
		if size >= threshold {
			entries = append(entries, FileEntry{
				Path:   childPath,
				Size:   size,
				IsDir:  c.IsDir(),
				Target: "Home (scan)",
			})
		}
	}
	return entries
}

// scanNodeModules walks ~ recursively and returns every node_modules directory
// >= minSize. It prunes the walk at each node_modules it finds (no nested
// scanning) and skips Library, .Trash, VCS dirs, and curated target paths.
func scanNodeModules(minSize int64, skip map[string]bool) []FileEntry {
	var entries []FileEntry

	home, err := os.UserHomeDir()
	if err != nil {
		return entries
	}

	filepath.WalkDir(home, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}

		name := d.Name()

		// Skip dirs that never contain useful node_modules.
		if path != home {
			switch name {
			case "Library", ".Trash", ".git", ".hg", ".svn":
				return filepath.SkipDir
			}
		}

		// Skip curated target paths (already covered by scanTargets).
		if path != home && containsTarget(filepath.Clean(path), skip) {
			return filepath.SkipDir
		}

		if name == "node_modules" {
			size, _ := dirSize(path)
			if size >= minSize {
				entries = append(entries, FileEntry{
					Path:   path,
					Size:   size,
					IsDir:  true,
					Target: "node_modules",
				})
			}
			return filepath.SkipDir // don't descend into node_modules
		}

		return nil
	})

	return entries
}

// confirm prompts the user with a y/N question and returns true on 'y'/'yes'.
func confirm(prompt string) bool {
	fmt.Print(prompt + " [y/N]: ")
	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	if err != nil {
		return false
	}
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}

// moveToTrash uses NSFileManager.trashItem (via swift -e) to move a path to
// the Trash. For system paths that return permission-denied (Code=513), it
// falls back to sudo mv so the user is prompted for their admin password once.
func moveToTrash(path string) error {
	escaped := strings.ReplaceAll(path, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	script := fmt.Sprintf(
		`import Foundation; try FileManager.default.trashItem(at: URL(fileURLWithPath: "%s"), resultingItemURL: nil)`,
		escaped,
	)
	cmd := exec.Command("swift", "-e", script)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	errStr := string(out)
	if strings.Contains(errStr, "Code=513") || strings.Contains(errStr, "afpAccessDenied") {
		return moveToTrashWithSudo(path)
	}
	msg := strings.TrimSpace(errStr)
	if msg != "" {
		return fmt.Errorf("%s", msg)
	}
	return err
}

// moveToTrashWithSudo moves a system-owned path to ~/.Trash via sudo mv.
// If mv fails (e.g. APFS volume boundary or daemon lock), it prints the
// manual command and returns an error rather than attempting a permanent delete.
func moveToTrashWithSudo(path string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	name := filepath.Base(path)
	dest := filepath.Join(home, ".Trash", name)
	if _, err := os.Stat(dest); err == nil {
		ext := filepath.Ext(name)
		base := strings.TrimSuffix(name, ext)
		dest = filepath.Join(home, ".Trash", fmt.Sprintf("%s_%d%s", base, time.Now().UnixMilli(), ext))
	}
	fmt.Printf("  (system path — enter admin password if prompted)\n")
	cmd := exec.Command("sudo", "mv", path, dest)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("  → could not move to Trash automatically.\n")
		fmt.Printf("    To delete manually: sudo rm -rf '%s'\n", path)
		return err
	}
	return nil
}

// --- interactive checkbox picker ---

var (
	unsafeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214")) // orange-yellow
	cursorStyle = lipgloss.NewStyle().Bold(true)
	footerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")) // dim
)

type pickerModel struct {
	entries   []FileEntry
	checked   []bool
	cursor    int
	offset    int
	width     int
	height    int
	safeMap   map[string]bool // CleanTarget.Name -> SafeToWipe
	cancelled bool
}

func newPickerModel(entries []FileEntry) pickerModel {
	safeMap := make(map[string]bool, len(targets)+1)
	for _, t := range targets {
		safeMap[t.Name] = t.SafeToWipe
	}
	safeMap["node_modules"] = true
	return pickerModel{
		entries: entries,
		checked: make([]bool, len(entries)),
		safeMap: safeMap,
	}
}

func (m pickerModel) Init() tea.Cmd { return nil }

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.clampOffset()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.cancelled = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.clampOffset()
			}
		case "down", "j":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
				m.clampOffset()
			}
		case " ":
			if len(m.checked) > 0 {
				m.checked[m.cursor] = !m.checked[m.cursor]
			}
		case "enter":
			return m, tea.Quit
		}
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			if m.offset > 0 {
				m.offset--
			}
		case tea.MouseButtonWheelDown:
			v := m.visibleRows()
			maxOff := len(m.entries) - v
			if maxOff < 0 {
				maxOff = 0
			}
			if m.offset < maxOff {
				m.offset++
			}
		case tea.MouseButtonLeft:
			if msg.Action != tea.MouseActionPress {
				break
			}
			// Title is row 0, blank is row 1, list starts at row 2.
			const listStartY = 2
			yRow := msg.Y - listStartY
			if yRow < 0 || yRow >= m.visibleRows() {
				break
			}
			idx := m.offset + yRow
			if idx < 0 || idx >= len(m.entries) {
				break
			}
			m.cursor = idx
			m.checked[idx] = !m.checked[idx]
		}
	}
	return m, nil
}

func (m *pickerModel) visibleRows() int {
	v := m.height - 3 // title, blank, footer
	if v < 1 {
		v = 1
	}
	return v
}

func (m *pickerModel) clampOffset() {
	v := m.visibleRows()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+v {
		m.offset = m.cursor - v + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
	maxOff := len(m.entries) - v
	if maxOff < 0 {
		maxOff = 0
	}
	if m.offset > maxOff {
		m.offset = maxOff
	}
}

func (m *pickerModel) selectedTotal() (count int, size int64) {
	for i, c := range m.checked {
		if c {
			count++
			size += m.entries[i].Size
		}
	}
	return
}

func (m pickerModel) View() string {
	var b strings.Builder
	b.WriteString("Pick items to move to Trash — ↑/↓ move, space to toggle, enter to confirm, q to quit\n\n")

	v := m.visibleRows()
	end := m.offset + v
	if end > len(m.entries) {
		end = len(m.entries)
	}
	for i := m.offset; i < end; i++ {
		e := m.entries[i]
		marker := "  "
		if i == m.cursor {
			marker = "▌ "
		}
		box := "[ ]"
		if m.checked[i] {
			box = "[x]"
		}
		warn := "  "
		unsafe := !m.safeMap[e.Target] // missing entry treated as safe (default false)
		if unsafe {
			warn = "⚠ "
		}
		// Width budget = total width minus the fixed prefix:
		//   marker(2) + box(3) + space(1) + warn(2) + size(10) + space(1) + category(25) + space(1) = 45
		const fixedPrefix = 45
		path := e.Path
		if m.width > fixedPrefix+4 {
			budget := m.width - fixedPrefix
			path = truncLeft(path, budget)
		}
		row := fmt.Sprintf("%s%s %s%-10s %-25s %s",
			marker, box, warn, humanSize(e.Size), e.Target, path)
		if unsafe {
			row = unsafeStyle.Render(row)
		}
		if i == m.cursor {
			row = cursorStyle.Render(row)
		}
		b.WriteString(row + "\n")
	}
	count, size := m.selectedTotal()
	footer := fmt.Sprintf("%d selected · %s", count, humanSize(size))
	b.WriteString("\n" + footerStyle.Render(footer) + "\n")
	return b.String()
}

func runPicker(entries []FileEntry) (picks []FileEntry, cancelled bool, err error) {
	m := newPickerModel(entries)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	finalModel, err := p.Run()
	if err != nil {
		return nil, false, err
	}
	final := finalModel.(pickerModel)
	if final.cancelled {
		return nil, true, nil
	}
	for i, c := range final.checked {
		if c {
			picks = append(picks, final.entries[i])
		}
	}
	return picks, false, nil
}

func main() {
	scanHomeFlag := flag.Bool("scan-home", false, "Also scan ~ for files/dirs >= 1 GB not in the curated targets (slower)")
	flag.Parse()

	fmt.Println("🧹 degunk — macOS disk cleanup")
	fmt.Println("==============================")
	fmt.Println()
	fmt.Println("Scanning common junk locations...")
	fmt.Println()

	// Only surface entries at least 10 MB so the list stays useful.
	const minSize = 10 * 1024 * 1024 // 10 MB
	entries, targetPaths := scanTargets(minSize)

	fmt.Println()
	fmt.Println("Scanning for node_modules under ~ (may take a moment)...")
	fmt.Println()
	nodeEntries := scanNodeModules(minSize, targetPaths)
	for _, e := range nodeEntries {
		fmt.Printf("  [node_modules] found %s (%s)\n", e.Path, humanSize(e.Size))
	}
	entries = append(entries, nodeEntries...)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Size > entries[j].Size
	})

	if *scanHomeFlag {
		const homeThreshold = 1024 * 1024 * 1024 // 1 GB
		fmt.Println()
		fmt.Printf("Scanning home directory for entries >= %s (this may take a minute)...\n", humanSize(homeThreshold))
		fmt.Println()
		homeEntries := scanHome(homeThreshold, targetPaths)
		entries = append(entries, homeEntries...)
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Size > entries[j].Size
		})
	}

	fmt.Println()
	if len(entries) == 0 {
		fmt.Println("No candidate files over 10 MB found. Nothing to clean.")
		return
	}

	fmt.Printf("Top candidates (>= %s), largest first:\n", humanSize(minSize))
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("%-4s %-10s %-25s %s\n", "#", "SIZE", "CATEGORY", "PATH")
	fmt.Println(strings.Repeat("-", 80))

	// Cap the displayed list so the terminal doesn't get flooded.
	max := len(entries)
	if max > 30 {
		max = 30
	}
	for i := 0; i < max; i++ {
		e := entries[i]
		fmt.Printf("%-4d %-10s %-25s %s\n",
			i+1, humanSize(e.Size), e.Target, e.Path)
	}
	fmt.Println(strings.Repeat("-", 80))

	var grandTotal int64
	for _, e := range entries[:max] {
		grandTotal += e.Size
	}
	fmt.Printf("Total reclaimable (top %d shown): %s\n\n", max, humanSize(grandTotal))

	picks, cancelled, err := runPicker(entries[:max])
	if err != nil {
		fmt.Printf("interactive picker requires a TTY: %v\n", err)
		os.Exit(1)
	}
	if cancelled {
		fmt.Println("Cancelled. Nothing was deleted.")
		return
	}

	if len(picks) == 0 {
		fmt.Println("No valid selections. Exiting.")
		return
	}

	fmt.Println("\nYou selected:")
	var selTotal int64
	for _, p := range picks {
		fmt.Printf("  %s  %s\n", humanSize(p.Size), p.Path)
		selTotal += p.Size
	}
	fmt.Printf("Total to move to Trash: %s\n\n", humanSize(selTotal))

	if !confirm("Move these to the Trash? (you can still restore from Trash)") {
		fmt.Println("Cancelled. Nothing was deleted.")
		return
	}

	for _, p := range picks {
		if err := moveToTrash(p.Path); err != nil {
			fmt.Printf("  ❌ failed: %s → %v\n", p.Path, err)
		} else {
			fmt.Printf("  ✅ trashed: %s\n", p.Path)
		}
	}

	fmt.Println("\nDone. Remember to empty the Trash to actually reclaim space.")
}
