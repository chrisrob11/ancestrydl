# ancestrydl

A command-line tool to download and preserve family tree data from Ancestry.com, **specifically designed for trees you can view but don't own** (where GEDCOM export is unavailable).

## 📋 Overview

`ancestrydl` extracts genealogical data directly from Ancestry.com's internal APIs, bypassing the limitations of GEDCOM exports. This is particularly useful for:

- **Shared family trees** where you have view access but aren't the owner
- **Trees shared with you** by other family members
- **Backup and archival** of family research you can access but don't control
- **Offline access** to genealogical data for analysis and research

The tool downloads:
- ✅ **People** - Names, birth/death dates, relationships
- ✅ **Events** - Life events with dates and places
- ✅ **Relationships** - Parent-child, spouse, and sibling connections
- ✅ **Media** - Photos, documents, and associated metadata
- ✅ **Interactive HTML Viewer** - Self-contained web page to browse your data offline

**Note:** This tool does **not** export GEDCOM files, as that functionality is only available to tree owners through Ancestry's official interface.

## ⚠️ Legal Disclaimer

**USE AT YOUR OWN RISK**

This tool is intended for **personal use only** to preserve your own genealogical data. By using this software, you acknowledge that:

- This tool accesses Ancestry.com through browser automation and reverse-engineered API endpoints
- Using automated tools may violate [Ancestry.com's Terms and Conditions](https://www.ancestry.com/cs/legal/termsandconditions)
- This software is **not affiliated with, endorsed by, or supported by** Ancestry.com
- You are solely responsible for ensuring your use complies with applicable laws and terms of service
- The authors assume no liability for any consequences resulting from use of this tool
- This is an independent, open-source project for personal data preservation

**Recommended Use:**
- Only download data you own or have explicit permission to access
- Respect copyright and intellectual property rights
- Use responsibly and do not overload Ancestry.com's servers
- Consider this tool for archival/backup purposes of accessible data

## 🚀 Installation

### Prerequisites

- **Go 1.21+** - [Install Go](https://golang.org/doc/install)
- **Chrome or Chromium browser** - Required for browser automation

### Install from Source

```bash
# Clone the repository
git clone https://github.com/chrisrob11/ancestrydl.git
cd ancestrydl

# Build the binary
make build

# Or install directly
go install
```

The binary will be available as `ancestrydl` in your `$GOPATH/bin`.

## 📚 Usage

### 1. Login to Ancestry.com

First, authenticate with your Ancestry.com credentials:

```bash
ancestrydl login -u your-username -p your-password
```

**With 2FA/Two-Factor Authentication:**

If you have two-factor authentication enabled, specify your preferred method:

```bash
# Use phone (SMS) for 2FA
ancestrydl login -u your-username -p your-password --2fa phone

# Use email for 2FA
ancestrydl login -u your-username -p your-password --2fa email

# Use authenticator app for 2FA
ancestrydl login -u your-username -p your-password --2fa app
```

The tool will open a browser window and wait for you to complete the 2FA process.

**What happens during login:**
- A browser window opens automatically
- The tool fills in your credentials
- If 2FA is enabled, you complete the verification in the browser
- Session cookies are securely stored in your system keyring
- The browser closes after successful authentication

### 2. List Available Trees

See all family trees you have access to:

```bash
ancestrydl list-trees
```

This shows:
- Tree ID (needed for downloads)
- Tree name
- Owner information
- Creation and modification dates
- Share settings

### 3. List People in a Tree

View all people in a specific tree:

```bash
ancestrydl list-people <tree-id>
```

Or set a default tree and omit the ID:

```bash
ancestrydl config set-default-tree <tree-id>
ancestrydl list-people
```

### 4. Download Complete Tree

Download all data from a family tree:

```bash
ancestrydl download-tree <tree-id>
```

**With custom output directory:**

```bash
ancestrydl download-tree <tree-id> --output ./my-family-tree
```

**With verbose logging (for debugging):**

```bash
ancestrydl download-tree <tree-id> --verbose
```

This creates HTTP request/response logs in `http_log.txt`.

**What gets downloaded:**

```
tree-<id>/
├── index.html              # Interactive viewer (open in browser)
├── people.json             # All persons with full details
├── metadata.json           # Tree information
├── media/
│   ├── photos/            # Photos and images
│   │   ├── Person-123-portrait-001.jpg
│   │   └── ...
│   └── documents/         # Documents and stories
│       ├── Person-456-certificate-001.pdf
│       └── ...
```

### 5. Configuration

Manage settings for easier usage:

```bash
# Set default tree (so you don't need to specify tree-id every time)
ancestrydl config set-default-tree <tree-id>

# View current default tree
ancestrydl config get-default-tree
```

### 6. Logout

Remove stored credentials:

```bash
ancestrydl logout
```

This deletes your session cookies from the system keyring.

## 🎯 Common Workflows

### Backup a Shared Family Tree

```bash
# 1. Login once
ancestrydl login -u your-email -p your-password --2fa phone

# 2. Find the tree you want to backup
ancestrydl list-trees

# 3. Download it
ancestrydl download-tree 123456789 --output ./family-backup

# 4. Open the viewer
open ./family-backup/index.html  # macOS
# or
xdg-open ./family-backup/index.html  # Linux
# or just double-click index.html in File Explorer (Windows)
```

### Download Multiple Trees

```bash
ancestrydl login -u your-email -p your-password

# Download each tree to a separate directory
ancestrydl download-tree 111111111 --output ./smiths
ancestrydl download-tree 222222222 --output ./johnsons
ancestrydl download-tree 333333333 --output ./williams
```

### Quick Exploration

```bash
ancestrydl login -u your-email -p your-password
ancestrydl config set-default-tree 123456789

# Now you can use commands without tree-id
ancestrydl list-people
ancestrydl download-tree --output ./quick-backup
```

## 🔍 Output Formats

### Interactive HTML Viewer (`index.html`)

A self-contained, single-file HTML viewer that works offline:
- Browse all people in the tree
- View relationships (parents, spouses, children)
- See life events with dates and places
- Access photos and documents
- Search and filter
- No internet connection required after download

### JSON Data Files

**`people.json`** - Structured person data:
```json
[
  {
    "personId": "12345",
    "fullName": "John Smith",
    "givenName": "John",
    "surname": "Smith",
    "gender": "m",
    "isLiving": false,
    "events": [
      {
        "type": "Birth",
        "date": "1850",
        "place": "New York, USA"
      }
    ],
    "parents": [...],
    "spouses": [...],
    "children": [...],
    "media": [...]
  }
]
```

**`metadata.json`** - Tree information:
```json
{
  "treeId": "123456789",
  "treeName": "Smith Family Tree",
  "exportDate": "2025-11-11T10:30:00Z",
  "personCount": 234
}
```

## 🛠️ Troubleshooting

### "Chrome/Chromium not found"

Install Chrome or Chromium browser:
- **macOS**: `brew install --cask google-chrome`
- **Ubuntu/Debian**: `sudo apt install chromium-browser`
- **Windows**: Download from [google.com/chrome](https://www.google.com/chrome/)

### "Cloudflare challenge detected"

If you see a Cloudflare security check:
1. Complete the challenge in the browser window that opens
2. The tool will automatically continue once you're verified
3. You have 30 seconds to complete the challenge

### "Session expired" errors

Your login session has expired. Simply run:
```bash
ancestrydl login -u your-email -p your-password
```

### Media downloads fail

- Check your internet connection
- Ensure you have permission to view the tree
- Try running with `--verbose` to see detailed error messages

## 🏗️ Architecture

This tool uses:
- **Browser Automation** ([go-rod](https://github.com/go-rod/rod)) - For authentication and cookie extraction
- **API Reverse Engineering** - Direct calls to Ancestry's internal JSON APIs
- **System Keyring** - Secure credential storage via OS-native keychains
- **No GEDCOM** - Extracts data directly, bypassing GEDCOM limitations for non-owner trees

**Why no GEDCOM?**

GEDCOM export is a feature only available to tree owners through Ancestry's official interface. When you can view but don't own a tree, there's no "Export Tree" button. This tool works around that limitation by accessing the same APIs that power Ancestry's web interface, allowing you to preserve data from trees shared with you.

## 📖 Advanced Usage

### Test Browser Automation

Verify your browser setup:

```bash
ancestrydl test-browser -u your-email -p your-password
```

This opens a browser, logs in, and keeps it open for 60 seconds so you can verify everything works.

### Network Capture (for debugging)

Capture API requests during a test login:

```bash
ancestrydl test-browser -u your-email -p your-password --capture-network -o api-log.json
```

Useful for debugging or understanding the API structure.

## 🤝 Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Submit a pull request

## 📄 License

MIT License - See [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

This project builds upon the pioneering work of other developers who have reverse-engineered Ancestry.com's internal APIs:

- **[cdhorn/ancestry-tools](https://github.com/cdhorn/ancestry-tools)** - Python tools for extracting and refactoring GEDCOM files with media. Provided insights into media API endpoints and GEDCOM processing.
- **[neRok00/ancestry-image-downloader](https://github.com/neRok00/ancestry-image-downloader)** - Script for downloading images from Ancestry trees. Documented the approach for querying the media API to obtain authorized download links.

Thank you to these developers for documenting their findings and sharing their knowledge with the community.

## 📮 Support

- **Issues**: [GitHub Issues](https://github.com/chrisrob11/ancestrydl/issues)
- **Discussions**: [GitHub Discussions](https://github.com/chrisrob11/ancestrydl/discussions)

---

**Remember**: This tool is for personal archival and backup purposes. Always respect copyright, intellectual property rights, and terms of service. Use responsibly.
