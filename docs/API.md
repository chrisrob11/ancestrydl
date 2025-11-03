# Ancestry.com Internal API Documentation

This document maps the reverse-engineered internal API endpoints used by Ancestry.com.

**Source Attribution:**
- Endpoints discovered through analysis of [cdhorn/ancestry-tools](https://github.com/cdhorn/ancestry-tools) and [neRok00/ancestry-image-downloader](https://github.com/neRok00/ancestry-image-downloader)
- Additional endpoints may be discovered through network inspection

---

## Authentication

### Login (Form-Based)
```
POST https://www.ancestry.com/account/signin/frame/authenticate
Content-Type: application/x-www-form-urlencoded

username={username}&password={password}
```

**Response:**
- Success: HTTP 200 with session cookies
- Failure: Response contains `"status":"invalidCredentials"`

**Required Cookies:**
Session cookies are automatically set after successful authentication and must be included in subsequent requests.

---

## Trees & GEDCOM

### List User Trees
```
GET https://www.ancestry.com/family-tree/tree/{user_id}/trees
```
*Note: Endpoint needs verification. May require browser automation to discover.*

### Download Tree (GEDCOM Export)
```
GET https://www.ancestry.com/tree/{tree_id}/export
```
*Note: GEDCOM export typically initiated through web UI. Exact API endpoint needs verification.*

---

## People & Records

### Get Person Page
```
GET https://search.ancestry.com/cgi-bin/sse.dll?indiv={indiv}&dbid={dbid}&h={pid}

Parameters:
  - indiv: Individual ID
  - dbid: Database ID
  - h: Person/Record ID (pid)
```

**Response:** HTML page containing person details and embedded data

**Extracted Data:**
- Image IDs via regex: `var iid='([^\s']+)';`
- Person details embedded in HTML

---

## Media & Images

### Get Media Information
```
GET https://www.ancestry.com/interactive/api/v2/Media/GetMediaInfo/{dbid}/{iid}/{pid}

Parameters:
  - dbid: Database ID
  - iid: Image ID (extracted from person page)
  - pid: Person/Record ID
```

**Response:** JSON
```json
{
  "ImageServiceUrlForDownload": "https://...",
  "...": "..."
}
```

### Get Image Download Token
```
GET https://www.ancestry.com/imageviewer/api/media/token?dbId={dbid}&imageId={imageid}

Parameters:
  - dbId: Database ID
  - imageId: Image ID
```

**Response:** JSON containing `imageDownloadUrl`

### Download User Media
```
GET https://www.ancestry.com/[path]?f=image&guid={guid}

Parameters:
  - guid: Media GUID from GEDCOM FILE records
```

### Download Newspaper Clipping
```
GET https://www.newspapers.com/clippings/download/?id={clipping_id}

Parameters:
  - clipping_id: Newspaper clipping identifier
```

**Response:** PDF file

---

## Source Citations

### Get Database Collection Information
```
GET https://www.ancestry.com/search/collections/{dbid}

Parameters:
  - dbid: Database ID
```

**Response:** HTML page with publisher and source information

---

## Notes

### Authentication Pattern
1. Browser automation completes login with 2FA
2. Extract session cookies from browser
3. Use cookies in HTTP client for API requests

### Rate Limiting
- No official rate limits documented
- Implement respectful delays between requests
- Monitor for HTTP 429 (Too Many Requests) responses

### Cookies Required
- Session authentication cookies from login
- May include: `session_id`, `auth_token`, or similar
- Exact cookie names need verification through browser inspection

### Content-Type Handling
- Image downloads may return various Content-Types
- Normalize JPEG variants (image/jpg, image/jpeg) to .jpg extension
- Check Content-Type header for proper file extension

---

## TODO: Endpoints to Discover

The following endpoints need to be identified through network inspection:

- [ ] List all trees for authenticated user
- [ ] Get tree metadata (name, person count, owner)
- [ ] Get tree members/people list
- [ ] Get detailed person information (dates, places, relationships)
- [ ] Search within tree
- [ ] Get hints/suggestions for person
- [ ] DNA matches endpoints (if applicable)

### How to Discover

Use browser DevTools Network tab while:
1. Viewing "My Trees" page
2. Opening a specific tree
3. Viewing person details
4. Exporting GEDCOM
5. Downloading media

Filter for XHR/Fetch requests and document:
- Request URL and method
- Request headers (especially auth)
- Request body/parameters
- Response format and structure
