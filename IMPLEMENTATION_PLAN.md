# Ancestry Downloader - Implementation Plan

## Phase 1: Foundation & Setup

### Commit 1: Project Initialization
- [ ] Initialize Go module (`go mod init`)
- [ ] Create basic project structure (directories only)
- [ ] Add `.gitignore` for Go projects
- [ ] Create empty placeholder files for main packages

**Files:**
- `go.mod`
- `.gitignore`
- `main.go` (empty main function)
- `pkg/ancestry/`, `commands/`, `pkg/storage/` directories

### Commit 2: Makefile and CI Setup
- [ ] Create `Makefile` with targets: build, test, lint, clean, install
- [ ] Add golangci-lint configuration (`.golangci.yml`)
- [ ] Create GitHub Actions workflow (`.github/workflows/ci.yml`)
- [ ] Configure workflow to run tests and linting on push/PR
- [ ] Test that `make build` works

**Files:**
- `Makefile`
- `.golangci.yml`
- `.github/workflows/ci.yml`

### Commit 3: CLI Framework Setup
- [ ] Install urfave/cli dependency
- [ ] Implement basic CLI app structure in `main.go`
- [ ] Add command stubs (login, logout, list-trees, export)
- [ ] Test that `ancestry-dl --help` works

**Files:**
- `main.go` (complete with command structure)
- `go.mod`, `go.sum` (with urfave/cli)

### Commit 4: Keyring Integration
- [ ] Add go-keyring dependency
- [ ] Create `pkg/config/credentials.go`
- [ ] Implement functions: `SaveCredentials()`, `GetCredentials()`, `DeleteCredentials()`
- [ ] Add basic error handling

**Files:**
- `pkg/config/credentials.go`
- `go.mod`, `go.sum` (with keyring dependency)

### Commit 5: Login Command Implementation
- [ ] Implement `commands/login.go`
- [ ] Accept username/password flags
- [ ] Validate inputs
- [ ] Store in keyring
- [ ] Print success/error messages

**Files:**
- `commands/login.go`
- Update `main.go` to wire up login command

### Commit 6: Logout Command Implementation
- [ ] Implement `commands/logout.go`
- [ ] Remove credentials from keyring
- [ ] Handle case where no credentials exist
- [ ] Print confirmation messages

**Files:**
- `commands/logout.go`
- Update `main.go` to wire up logout command

## Phase 2: Browser Automation

### Commit 7: Browser Automation Setup
- [ ] Add go-rod dependency
- [ ] Create `pkg/ancestry/client.go` with basic browser setup
- [ ] Implement `NewClient()` function
- [ ] Add function to launch browser and navigate to Ancestry.com
- [ ] Test basic navigation (no authentication yet)

**Files:**
- `pkg/ancestry/client.go`
- `go.mod`, `go.sum` (with go-rod)

### Commit 8: Authentication Flow
- [ ] Create `pkg/ancestry/auth.go`
- [ ] Implement `Login(username, password)` function
- [ ] Handle login form submission
- [ ] Detect successful/failed login
- [ ] Add error handling for common issues

**Files:**
- `pkg/ancestry/auth.go`
- Update `pkg/ancestry/client.go` if needed

### Commit 9: List Trees Command - Part 1
- [ ] Implement `commands/list.go`
- [ ] Retrieve credentials from keyring
- [ ] Initialize browser client
- [ ] Authenticate to Ancestry.com

**Files:**
- `commands/list.go`
- Update `main.go` to wire up command

### Commit 10: List Trees Command - Part 2
- [ ] Navigate to trees page after authentication
- [ ] Scrape tree list (ID, name, person count)
- [ ] Format and display results to user
- [ ] Handle case where user has no trees

**Files:**
- Update `commands/list.go`
- Update `pkg/ancestry/client.go` with tree listing methods

## Phase 3: GEDCOM Export

### Commit 11: Export Command Structure
- [ ] Implement `commands/export.go` skeleton
- [ ] Add all flags (tree-id, output, media-types, slow, resume, dry-run, delay)
- [ ] Validate inputs
- [ ] Create output directory structure

**Files:**
- `commands/export.go`
- Update `main.go` to wire up command

### Commit 12: GEDCOM Parser Integration
- [ ] Add GEDCOM parser library dependency
- [ ] Create `pkg/gedcom/parser.go`
- [ ] Implement `Parse(filename)` function
- [ ] Extract individual IDs from parsed GEDCOM

**Files:**
- `pkg/gedcom/parser.go`
- `go.mod`, `go.sum` (with GEDCOM library)

### Commit 13: GEDCOM Export Automation
- [ ] Create `pkg/ancestry/gedcom.go`
- [ ] Implement browser automation to navigate to export page
- [ ] Click through export wizard
- [ ] Download GEDCOM file to output directory

**Files:**
- `pkg/ancestry/gedcom.go`
- Update `commands/export.go` to use GEDCOM export

### Commit 14: Progress Bar Integration
- [ ] Add progressbar dependency
- [ ] Create progress bar for GEDCOM export
- [ ] Add progress bar for person iteration (preparation for media download)

**Files:**
- Update `commands/export.go` with progress bars
- `go.mod`, `go.sum` (with progressbar)

## Phase 4: Media Download

### Commit 15: Media Scraping Foundation
- [ ] Create `pkg/ancestry/media.go`
- [ ] Implement function to navigate to person profile
- [ ] Identify media elements on page
- [ ] Extract media URLs and metadata

**Files:**
- `pkg/ancestry/media.go`

### Commit 16: Media Download Implementation
- [ ] Implement media download function
- [ ] Add systematic naming: `{personID}_{mediaType}_{index}.{ext}`
- [ ] Save files to appropriate media subdirectories
- [ ] Handle download errors/retries

**Files:**
- Update `pkg/ancestry/media.go`
- Create `pkg/storage/organizer.go` for file organization

### Commit 17: Metadata Generation
- [ ] Create metadata JSON for each downloaded media item
- [ ] Include: person ID, media type, source URL, download timestamp
- [ ] Save per-person metadata files

**Files:**
- Update `pkg/ancestry/media.go`
- Update `pkg/storage/organizer.go`

### Commit 18: Manifest Generation
- [ ] Create `manifest.json` generation
- [ ] Map people to their media files
- [ ] Include GEDCOM data references
- [ ] Save manifest to output directory

**Files:**
- Update `pkg/storage/organizer.go`
- Update `commands/export.go` to generate manifest

## Phase 5: Polish & Features

### Commit 19: Rate Limiting
- [ ] Implement configurable delays between requests
- [ ] Add `--slow` flag support (5+ second delays)
- [ ] Add `--delay` flag for custom timing
- [ ] Respect rate limit responses

**Files:**
- Update `pkg/ancestry/client.go` with rate limiting
- Update `commands/export.go` to pass delay settings

### Commit 20: Resume Capability
- [ ] Create progress state file (`.ancestry-dl-progress.json`)
- [ ] Track downloaded files
- [ ] Skip already-downloaded items on resume
- [ ] Clean up state file on completion

**Files:**
- Create `pkg/storage/state.go`
- Update `commands/export.go` with resume logic

### Commit 21: Dry Run Mode
- [ ] Implement `--dry-run` flag
- [ ] Show what would be downloaded without actually downloading
- [ ] Display estimated file counts and sizes if possible

**Files:**
- Update `commands/export.go` with dry-run logic

### Commit 22: Error Handling & Logging
- [ ] Add comprehensive error handling throughout
- [ ] Create `logs/` directory and `download.log`
- [ ] Log all operations with timestamps
- [ ] Add retry logic for failed downloads

**Files:**
- Create logging utility
- Update all command files with proper error handling

### Commit 23: Configuration File Support
- [ ] Create `~/.ancestry-dl/config.yaml` support
- [ ] Load default settings from config
- [ ] Allow CLI flags to override config
- [ ] Document config options

**Files:**
- Create `pkg/config/config.go`
- Update commands to read from config

### Commit 24: Documentation
- [ ] Create comprehensive README.md
- [ ] Add usage examples
- [ ] Document all commands and flags
- [ ] Add responsible use guidelines
- [ ] Include setup instructions

**Files:**
- `README.md`
- `LICENSE` (MIT)

### Commit 25: User Agent & Best Practices
- [ ] Set proper User-Agent header
- [ ] Add request headers for transparency
- [ ] Final polish on automation behavior

**Files:**
- Update `pkg/ancestry/client.go`

---

## Total: 25 Commits

Each commit should be buildable and testable independently where possible.
