# Ancestry Downloader

## 🎯 Purpose

Download and preserve family tree data from Ancestry.com for **trees you can view but don't own**, where GEDCOM export functionality is unavailable.

## ⚠️ Why No GEDCOM?

**GEDCOM export is owner-only.** Ancestry's GEDCOM export feature is only available to tree owners. If you're viewing a tree shared with you by a family member, there's no "Export Tree" button.

This tool works around that limitation by accessing Ancestry's internal JSON APIs directly—the same APIs used by the web interface when you browse a tree. This approach works regardless of ownership.

## ✅ Features

### Core Functionality
- **Authentication**: Login with 2FA support (phone/email/app)
- **Tree Discovery**: List all trees you have access to
- **Data Extraction**: Download complete tree data including:
  - People (names, dates, relationships)
  - Life events with dates and places
  - Family relationships (parents, spouses, children, siblings)
  - Photos and documents
- **Smart Processing**:
  - Relationship inference from family connections
  - Event type inference from family events
  - Place extraction from nested structures
- **Output**:
  - Interactive HTML viewer (offline, self-contained)
  - JSON data files for programmatic access
  - Organized media downloads

### API Endpoints
- `/api/trees` - List accessible trees
- `/api/treeviewer/tree/allpersons/{treeId}` - Get all persons (paginated)
- `/api/treeviewer/tree/newfamilyview/{treeId}` - Get relationships and events
- `/api/media/viewer/v1/trees/{treeId}/people/{personId}` - Get person media
- `/facts/page/{treeId}/person/{personId}` - Get detailed person facts

### Commands
- `login` - Authenticate with 2FA support
- `logout` - Remove stored credentials
- `list-trees` - List all accessible trees
- `list-people` - List all people in a tree
- `config` - Manage configuration (default tree)
- `download-tree` - Complete tree download
- `test-browser` - Test browser automation

## 🚧 Known Limitations

### By Design
1. **No GEDCOM Output**: This tool does not generate GEDCOM files. Use the downloaded JSON data or HTML viewer instead.
2. **API Dependency**: Relies on Ancestry's internal APIs which may change without notice.
3. **Authentication Required**: Requires valid Ancestry.com credentials and appropriate permissions.
4. **Rate Limiting**: Large trees (1000+ people) may take time to download due to respectful rate limiting.

### Technical
1. **Browser Requirement**: Chrome/Chromium must be installed for authentication.
2. **Session Expiration**: Cookies expire after days/weeks; re-login required periodically.
3. **Privacy Restrictions**: Living persons may have limited visibility based on Ancestry's privacy settings.

## 🔧 Technical Stack

- **Language**: Go 1.21+
- **Browser Automation**: [go-rod](https://github.com/go-rod/rod)
- **Secure Storage**: System keyring integration
- **Architecture**: Direct API calls, no GEDCOM parsing

---

**Status**: ✅ **Production Ready**

The tool is feature-complete for downloading and preserving family tree data from Ancestry.com trees you can view but don't own.
