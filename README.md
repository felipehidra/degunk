# degunk 🧹

**De-gunk your Mac.** Scan 113 known junk spots — Xcode leftovers, package-manager caches, browser blobs, chat-app media, VM disks — see what's actually eating your disk, and move what you pick to the **Trash**. Nothing is ever deleted outright.

```
$ degunk
🧹 degunk — macOS disk cleanup
==============================

Scanning common junk locations...

  [User Caches] total: 6.03 GB  (/Users/you/Library/Caches)
  [Xcode DerivedData] total: 14.21 GB  (/Users/you/Library/Developer/Xcode/DerivedData)
  [Go module cache] total: 8.77 GB  (/Users/you/go/pkg/mod)
  [npm cache] total: 1.92 GB  (/Users/you/.npm)
  ...

Top candidates (>= 10.00 MB), largest first:
--------------------------------------------------------------------------------
#    SIZE       CATEGORY                  PATH
--------------------------------------------------------------------------------
1    9.34 GB    Simulator runtimes        /Library/Developer/CoreSimulator/Profiles/Runtimes/iOS 17.5.simruntime
2    4.87 GB    Xcode DerivedData         /Users/you/Library/Developer/Xcode/DerivedData/MyApp-gtkeb
3    2.10 GB    Go module cache           /Users/you/go/pkg/mod/cache
...
--------------------------------------------------------------------------------
Total reclaimable (top 30 shown): 41.52 GB
```

Then an interactive checkbox picker lets you toggle items, and after a typed `y` confirmation the selected paths move to the Trash. Empty the Trash yourself when you're sure.

## Install

```bash
go install github.com/felipehidra/degunk@latest
```

Or clone and build — it's a single `main.go`, read it first if you like:

```bash
git clone https://github.com/felipehidra/degunk
cd degunk
go build -o degunk .
./degunk
```

Requires Go (`brew install go`). macOS only, by design.

## Usage

1. Run `degunk`. It prints a per-category summary while scanning, then the top 30 candidates ≥ 10 MB, largest first.
2. Pick items in the checkbox list (space to toggle, Enter to confirm).
3. Type `y` at the final confirmation.
4. Selected items are moved to the **Trash** — restore anything from there if you change your mind, and empty it to actually reclaim the space.

Flags:

- `--scan-home` — also sweep your home directory for files/dirs ≥ 1 GB that live outside the curated targets (slower). `node_modules` directories under `~` are found in the normal scan already.

## What it scans

- **Xcode & simulators** — DerivedData, device support symbols, simulator runtimes and devices, SwiftUI previews, archives, test devices
- **Package managers & build caches** — npm, pnpm, Yarn, Bun, pip, uv, Go, Cargo, Gradle, Maven, CocoaPods, Composer, Homebrew, ccache and friends
- **Browsers** — Chrome, Safari, Firefox, Edge, Brave, Arc caches and service-worker storage
- **Chat & media apps** — Slack, Discord, Teams, Zoom, WhatsApp, Telegram, Spotify, Podcasts, Apple Books/TV downloads
- **VMs & containers** — Docker, OrbStack, colima, Lima, Podman, Vagrant, Android emulator images
- **ML & models** — Hugging Face, PyTorch hub, Ollama
- **System odds and ends** — logs, diagnostic reports, Quick Look thumbnails, Trash itself

<details>
<summary>Full list of all 113 targets</summary>

| Target | Path | Wipe? |
|---|---|---|
| User Caches | `~/Library/Caches` | safe to wipe |
| User Logs | `~/Library/Logs` | safe to wipe |
| System Caches | `/Library/Caches` | safe to wipe |
| Xcode DerivedData | `~/Library/Developer/Xcode/DerivedData` | safe to wipe |
| Xcode Archives | `~/Library/Developer/Xcode/Archives` | ⚠ review first |
| Xcode iOS DeviceSupport | `~/Library/Developer/Xcode/iOS DeviceSupport` | safe to wipe |
| Xcode Simulators Caches | `~/Library/Developer/CoreSimulator/Caches` | safe to wipe |
| iOS Device Backups | `~/Library/Application Support/MobileSync/Backup` | ⚠ review first |
| Trash | `~/.Trash` | safe to wipe |
| Downloads | `~/Downloads` | ⚠ review first |
| npm cache | `~/.npm` | safe to wipe |
| Yarn cache | `~/.yarn/cache` | safe to wipe |
| pip cache | `~/Library/Caches/pip` | safe to wipe |
| Homebrew cache | `~/Library/Caches/Homebrew` | safe to wipe |
| Go build cache | `~/Library/Caches/go-build` | safe to wipe |
| Docker | `~/Library/Containers/com.docker.docker/Data/vms` | ⚠ review first |
| Slack cache | `~/Library/Application Support/Slack/Cache` | safe to wipe |
| Chrome cache | `~/Library/Caches/Google/Chrome` | safe to wipe |
| Safari cache | `~/Library/Caches/com.apple.Safari` | safe to wipe |
| Firefox cache | `~/Library/Caches/Firefox` | safe to wipe |
| Edge cache | `~/Library/Caches/com.microsoft.edgemac` | safe to wipe |
| Brave cache | `~/Library/Caches/BraveSoftware` | safe to wipe |
| Arc cache | `~/Library/Caches/company.thebrowser.Browser` | safe to wipe |
| Discord cache | `~/Library/Application Support/discord/Cache` | safe to wipe |
| Teams cache | `~/Library/Containers/com.microsoft.teams2/Data/Library/Caches` | safe to wipe |
| Zoom cache | `~/Library/Caches/us.zoom.xos` | safe to wipe |
| WhatsApp cache | `~/Library/Caches/net.whatsapp.WhatsApp` | safe to wipe |
| Telegram cache | `~/Library/Caches/ru.keepcoder.Telegram` | safe to wipe |
| Spotify cache | `~/Library/Caches/com.spotify.client` | safe to wipe |
| Cursor cache | `~/Library/Application Support/Cursor/Cache` | safe to wipe |
| JetBrains caches | `~/Library/Caches/JetBrains` | safe to wipe |
| Gradle cache | `~/.gradle/caches` | safe to wipe |
| Cargo registry | `~/.cargo/registry` | safe to wipe |
| Cargo git cache | `~/.cargo/git` | safe to wipe |
| pnpm store | `~/Library/pnpm/store` | safe to wipe |
| CocoaPods cache | `~/Library/Caches/CocoaPods` | safe to wipe |
| Flutter pub cache | `~/.pub-cache` | safe to wipe |
| iPhone Software Updates | `~/Library/iTunes/iPhone Software Updates` | safe to wipe |
| iPad Software Updates | `~/Library/iTunes/iPad Software Updates` | safe to wipe |
| Xcode Simulator Devices | `~/Library/Developer/CoreSimulator/Devices` | ⚠ review first |
| Xcode SwiftUI Previews | `~/Library/Developer/Xcode/UserData/Previews` | safe to wipe |
| Xcode IDE cache | `~/Library/Caches/com.apple.dt.Xcode` | safe to wipe |
| Simulator dyld caches | `/Library/Developer/CoreSimulator/Caches/dyld` | safe to wipe |
| Simulator runtimes | `/Library/Developer/CoreSimulator/Profiles/Runtimes` | ⚠ review first |
| Simulator device logs | `~/Library/Logs/CoreSimulator` | safe to wipe |
| Swift Package Manager cache | `~/Library/Caches/org.swift.swiftpm` | safe to wipe |
| Xcode DocSets | `~/Library/Developer/Shared/Documentation/DocSets` | ⚠ review first |
| Diagnostic reports | `~/Library/Logs/DiagnosticReports` | safe to wipe |
| Quick Look cache | `~/Library/Caches/com.apple.QuickLook` | safe to wipe |
| Adobe media cache | `~/Library/Application Support/Adobe/Common/Media Cache` | safe to wipe |
| Android Studio caches | `~/Library/Caches/Google` | safe to wipe |
| Android AVD images | `~/.android/avd` | ⚠ review first |
| rbenv versions | `~/.rbenv/versions` | ⚠ review first |
| Zoom recordings | `~/Documents/Zoom` | ⚠ review first |
| VS Code cache | `~/Library/Application Support/Code/Cache` | safe to wipe |
| VS Code CachedData | `~/Library/Application Support/Code/CachedData` | safe to wipe |
| Cursor CachedData | `~/Library/Application Support/Cursor/CachedData` | safe to wipe |
| Chrome Service Worker | `~/Library/Application Support/Google/Chrome/Default/Service Worker` | safe to wipe |
| Spotify offline cache | `~/Library/Application Support/Spotify/PersistentCache` | safe to wipe |
| Gradle wrapper dists | `~/.gradle/wrapper/dists` | safe to wipe |
| Maven local repo | `~/.m2/repository` | safe to wipe |
| Hugging Face cache | `~/.cache/huggingface` | safe to wipe |
| PyTorch hub cache | `~/.cache/torch/hub` | safe to wipe |
| Yarn Berry cache | `~/.yarn/berry/cache` | safe to wipe |
| Bun install cache | `~/.bun/install/cache` | safe to wipe |
| Deno cache | `~/Library/Caches/deno` | safe to wipe |
| Composer cache | `~/.composer/cache` | safe to wipe |
| Bundler cache | `~/.bundle/cache` | safe to wipe |
| Arc Service Worker | `~/Library/Application Support/Arc/User Data/Default/Service Worker` | safe to wipe |
| Brave Service Worker | `~/Library/Application Support/BraveSoftware/Brave-Browser/Default/Service Worker` | safe to wipe |
| Edge Service Worker | `~/Library/Application Support/Microsoft Edge/Default/Service Worker` | safe to wipe |
| pyenv versions | `~/.pyenv/versions` | ⚠ review first |
| nvm Node versions | `~/.nvm/versions/node` | ⚠ review first |
| rustup toolchains | `~/.rustup/toolchains` | ⚠ review first |
| Miniconda packages | `~/miniconda3/pkgs` | ⚠ review first |
| Anaconda packages | `~/anaconda3/pkgs` | ⚠ review first |
| OrbStack data | `~/.orbstack/data` | ⚠ review first |
| colima config | `~/.colima` | ⚠ review first |
| Lima VM disks | `~/.lima` | ⚠ review first |
| Podman containers | `~/.local/share/containers` | ⚠ review first |
| Vagrant boxes | `~/.vagrant.d/boxes` | ⚠ review first |
| Mail Downloads | `~/Library/Containers/com.apple.mail/Data/Library/Mail Downloads` | ⚠ review first |
| Ollama models | `~/.ollama/models` | ⚠ review first |
| ccache (C/C++) | `~/.ccache` | safe to wipe |
| Bazelisk cache | `~/.cache/bazelisk` | safe to wipe |
| Cabal store | `~/.cabal/store` | ⚠ review first |
| Stack data | `~/.stack` | ⚠ review first |
| GHCup cache | `~/.ghcup/cache` | safe to wipe |
| opam download cache | `~/.opam/download-cache` | safe to wipe |
| Hex packages | `~/.hex/packages` | safe to wipe |
| Mix archives | `~/.mix/archives` | safe to wipe |
| Apple Books downloads | `~/Library/Containers/com.apple.iBooksX/Data/Library/Caches/com.apple.iBooksX/Books` | ⚠ review first |
| Apple TV downloads | `~/Library/Containers/com.apple.TV/Data/Library/Application Support/com.apple.tv/private/Downloads` | ⚠ review first |
| uv cache | `~/.cache/uv` | safe to wipe |
| Playwright browsers | `~/Library/Caches/ms-playwright` | safe to wipe |
| Puppeteer browsers | `~/.cache/puppeteer` | safe to wipe |
| Electron cache | `~/Library/Caches/electron` | safe to wipe |
| electron-builder cache | `~/Library/Caches/electron-builder` | safe to wipe |
| Go module cache | `~/go/pkg/mod` | safe to wipe |
| Xcode test devices | `~/Library/Developer/XCTestDevices` | safe to wipe |
| watchOS DeviceSupport | `~/Library/Developer/Xcode/watchOS DeviceSupport` | safe to wipe |
| tvOS DeviceSupport | `~/Library/Developer/Xcode/tvOS DeviceSupport` | safe to wipe |
| Unity cache | `~/Library/Unity/cache` | safe to wipe |
| Terraform plugin cache | `~/.terraform.d/plugin-cache` | safe to wipe |
| CocoaPods spec repos | `~/.cocoapods/repos` | safe to wipe |
| Android SDK system images | `~/Library/Android/sdk/system-images` | ⚠ review first |
| Steam shader cache | `~/Library/Application Support/Steam/steamapps/shadercache` | safe to wipe |
| Podcasts episodes | `~/Library/Group Containers/243LU875E5.groups.com.apple.podcasts/Library/Cache` | safe to wipe |
| Dropbox cache | `~/Dropbox/.dropbox.cache` | safe to wipe |
| Steam games | `~/Library/Application Support/Steam/steamapps/common` | ⚠ review first |
| WhatsApp media | `~/Library/Group Containers/group.net.whatsapp.WhatsApp.shared/Message/Media` | ⚠ review first |
| Google Drive cache | `~/Library/Application Support/Google/DriveFS` | ⚠ review first |
| GarageBand sound library | `/Library/Application Support/GarageBand` | ⚠ review first |

</details>

"Safe to wipe" means the OS or app fully regenerates the data (caches, build artifacts). "⚠ review first" means the data is *not* regenerable — installed games, chat photos, iOS backups, your Downloads — so look before you leap.

## Safety

- **Trash, not `rm`.** Items are moved with macOS's own `NSFileManager.trashItem` API. Until you empty the Trash, everything is restorable.
- **Nothing happens without a typed `y`.** The tool never deletes on its own.
- **No `sudo` needed.** Unreadable paths are silently skipped. For the few root-owned paths (like old simulator runtimes), it offers a `sudo mv` fallback that still lands in the Trash — or tells you the manual command and touches nothing.
- **Non-regenerable data is labeled.** Targets holding real user data are marked "review first" instead of being treated as junk.

## Contributing

A new cleanup target is **one line in `main.go`** — a name, a path, a description, and a safe-to-wipe flag. PRs welcome.

Other ideas on the list: `docker system prune` / `brew cleanup` one-shot actions, a `--dry-run` flag, JSON output.

## License

[MIT](LICENSE)
