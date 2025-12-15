package ancestry

import (
	"fmt"
	"strings"
	"time"
)

// Tree represents an Ancestry family tree
// This struct supports both the old and new tree list APIs
type Tree struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"desc,omitempty"`
	Owner             string    `json:"owner,omitempty"`        // From /api/media/viewer/api/trees/list
	OwnerUserID       string    `json:"ownerUserId,omitempty"`  // From old API
	DateModified      time.Time `json:"dateModified,omitempty"` // Old format
	DateCreated       time.Time `json:"dateCreated,omitempty"`  // Old format
	CreatedOn         time.Time `json:"createdOn,omitempty"`    // New format
	ModifiedOn        time.Time `json:"modifiedOn,omitempty"`   // New format
	CanSeeLiving      bool      `json:"canSeeLiving,omitempty"` // From /api/media/viewer/api/trees/list
	TotalInvitedCount int       `json:"totalInvitedCount,omitempty"`
	SH                bool      `json:"sh,omitempty"`
}

// TreeListResponse represents the response from the tree list API
type TreeListResponse struct {
	Trees []Tree `json:"trees"`
	Count int    `json:"count"`
}

// Person represents a person in the family tree (matches Ancestry API structure)
type Person struct {
	GID           map[string]interface{} `json:"gid"`
	Names         []Name                 `json:"Names,omitempty"`
	Genders       []Gender               `json:"Genders,omitempty"`
	Events        []Event                `json:"Events,omitempty"`
	L             interface{}            `json:"l,omitempty"`
	Lus           interface{}            `json:"lus,omitempty"`
	MD            string                 `json:"md,omitempty"` // Modified date
	CD            string                 `json:"cd,omitempty"` // Created date
	Kinships      []interface{}          `json:"Kinships,omitempty"`
	KinshipLabel  string                 `json:"kinshipLabel,omitempty"`
	Family        []FamilyMember         `json:"Family,omitempty"` // Family relationships
	PID           string                 `json:"pid,omitempty"`
	IsLiving      bool                   `json:"isLiving,omitempty"`
	GivenName     string                 `json:"gname,omitempty"` // Flat field for given name
	Surname       string                 `json:"sname,omitempty"` // Flat field for surname
	Gender        string                 `json:"gender,omitempty"`
	EventsSummary []interface{}          `json:"events,omitempty"`
}

// FamilyMember represents a family relationship
type FamilyMember struct {
	Type string                 `json:"t"`    // F=Father, M=Mother, H=Husband, W=Wife, C=Child
	TGID map[string]interface{} `json:"tgid"` // Target person ID
	CD   string                 `json:"cd,omitempty"`
	MD   string                 `json:"md,omitempty"`
	Mod  string                 `json:"mod,omitempty"`
}

// GetPersonID extracts the person ID from the GID map
// The person ID is stored in gid.v (e.g., {"v":"232573524428:1030:197283789"})
// Returns the full ID including tree context
func (p *Person) GetPersonID() string {
	// First check if PID is already set (for backwards compatibility)
	if p.PID != "" {
		return p.PID
	}

	// Extract from GID map
	if p.GID != nil {
		if v, ok := p.GID["v"].(string); ok {
			return v
		}
	}

	return ""
}

// GetShortPersonID extracts just the person identifier without tree context
// For example, "232573524428:1030:197283789" becomes "232573524428"
func (p *Person) GetShortPersonID() string {
	fullID := p.GetPersonID()
	if fullID == "" {
		return ""
	}

	// Split on colon and return first part
	parts := strings.Split(fullID, ":")
	return parts[0]
}

// GetDisplayName returns the person's name from the Names array or falls back to flat fields
func (p *Person) GetDisplayName() string {
	// Try to get name from Names array first
	if len(p.Names) > 0 {
		givenName := p.Names[0].GivenName
		surname := p.Names[0].Surname
		if givenName != "" || surname != "" {
			return strings.TrimSpace(fmt.Sprintf("%s %s", givenName, surname))
		}
	}

	// Fall back to flat fields
	if p.GivenName != "" || p.Surname != "" {
		return strings.TrimSpace(fmt.Sprintf("%s %s", p.GivenName, p.Surname))
	}

	return ""
}

// Name represents a person's name in structured format
type Name struct {
	ID        string `json:"id"`
	GivenName string `json:"g"` // Given name / first name
	Surname   string `json:"s"` // Surname / last name
}

// Gender represents a person's gender
type Gender struct {
	ID     string `json:"id"`
	Gender string `json:"g"` // "m" or "f"
}

// Event represents a life event (birth, death, marriage, etc.)
type Event struct {
	ID          string                   `json:"id,omitempty"`
	Type        string                   `json:"t,omitempty"`
	Date        interface{}              `json:"d,omitempty"`
	NPS         []map[string]interface{} `json:"nps,omitempty"`  // Nested place structure
	Description string                   `json:"desc,omitempty"` // Event description/notes
}

// FamilyViewResponse represents the response from the newfamilyview API
type FamilyViewResponse struct {
	V       string      `json:"v"` // Version string like "3.0"
	Persons []Person    `json:"Persons"`
	Focus   interface{} `json:"focus,omitempty"` // Can be object or string
}

// TreeInfo represents tree metadata
type TreeInfo struct {
	TreeID          string `json:"treeId"`
	TreeName        string `json:"treeName"`
	TreeDescription string `json:"treeDescription,omitempty"`
	IsPrivate       bool   `json:"isPrivate"`
	PersonCount     int    `json:"personCount,omitempty"`
}

// FocusHistoryResponse represents the focus history with person data
type FocusHistoryResponse struct {
	History []FocusHistoryItem `json:"History"`
	Persons map[string]Person  `json:"Persons"`
}

// FocusHistoryItem represents a single item in focus history
type FocusHistoryItem struct {
	PID       string `json:"pid"`
	Timestamp int64  `json:"ts"`
}

// UserData represents user account information
type UserData struct {
	User                     map[string]interface{} `json:"user"`
	HasHints                 bool                   `json:"hasHints"`
	MostRecentlyViewedTreeID interface{}            `json:"mostRecentlyViewedTreeId"`
	NotificationsCount       int                    `json:"notificationsCount"`
	HintCount                int                    `json:"hintcount"`
}

// PersonMedia represents media attached to a person
type PersonMedia struct {
	PersonID   string      `json:"personId,omitempty"`
	TreeID     string      `json:"treeId,omitempty"`
	MediaItems []MediaItem `json:"mediaItems,omitempty"`
}

// MediaItem represents a single media item (photo, document, etc.)
type MediaItem struct {
	MediaID     string                 `json:"mediaId,omitempty"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type,omitempty"`
	URL         string                 `json:"url,omitempty"`
	ThumbURL    string                 `json:"thumbUrl,omitempty"`
	Category    string                 `json:"category,omitempty"`
	Raw         map[string]interface{} `json:"-"` // Store raw response for debugging
}

// InitialState represents the window.INITIAL_STATE object in person pages
type InitialState struct {
	Redux ReduxState `json:"redux"`
}

// ReduxState is part of the INITIAL_STATE
type ReduxState struct {
	Person PersonState `json:"person"`
}

// PersonState holds person-specific page data
type PersonState struct {
	PageData PageData `json:"pageData"`
}

// PageData contains the person facts
type PageData struct {
	PersonFacts map[string]interface{} `json:"personFacts"`
}

// PersonFact represents a single fact, which may have media
type PersonFact struct {
	PrimaryMediaItem *PrimaryMediaItem `json:"primaryMediaItem"`
}

// PrimaryMediaItem contains direct URLs to media with metadata
type PrimaryMediaItem struct {
	Type        string `json:"type"`
	URL         string `json:"url"`
	PreviewURL  string `json:"previewUrl"`
	MediaID     string `json:"mediaId"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Subcategory string `json:"subcategory"`
	Description string `json:"description"`
	Date        string `json:"date"`
}

// MediaViewerResponse represents the response from /api/media/viewer/v1/trees/{treeId}/people/{personId}
type MediaViewerResponse struct {
	MediaCount int                 `json:"mediaCount"`
	HasMedia   bool                `json:"hasMedia"`
	Objects    []MediaViewerObject `json:"objects"`
}

// MediaViewerObject represents a single media item from the media viewer API
type MediaViewerObject struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Type         string `json:"type"`
	Category     string `json:"category"`
	Subcategory  string `json:"subcategory"`
	URL          string `json:"url"`
	CollectionID int    `json:"collectionId"`
	Description  string `json:"description"`
	Date         string `json:"date"`
	PreviewURL   string `json:"previewUrl"`
}

// ResearchData represents the window.researchData object embedded in Facts pages
type ResearchData struct {
	PersonFacts   []PersonFactDetail   `json:"PersonFacts"`
	PersonSources []PersonSourceDetail `json:"PersonSources"`
}

// PersonSourceDetail represents a detailed source record from the PersonSources array in ResearchData
type PersonSourceDetail struct {
	AssertionIds          string `json:"AssertionIds"` // Space-separated string of Assertion IDs
	CitationId            string `json:"CitationId"`
	DatabaseId            string `json:"DatabaseId"`
	RecordId              string `json:"RecordId"`
	SourceId              string `json:"SourceId"` // Often same as DatabaseId but not always
	Title                 string `json:"Title"`
	RecordImageUrl        string `json:"RecordImageUrl"`
	RecordImagePreviewUrl string `json:"RecordImagePreviewUrl"`
	ViewRecordUrl         string `json:"ViewRecordUrl"`
	ViewRecordImageUrl    string `json:"ViewRecordImageUrl"`
}

// PersonFactDetail represents a single fact/event with complete details
type PersonFactDetail struct {
	Type              int                    `json:"Type"`
	TypeString        string                 `json:"TypeString"`
	Title             string                 `json:"Title,omitempty"` // Custom event title (like "Prison")
	Place             string                 `json:"Place,omitempty"`
	PlaceGpids        map[string]interface{} `json:"PlaceGpids,omitempty"`
	Description       string                 `json:"Description,omitempty"`
	Date              interface{}            `json:"Date,omitempty"`
	SourceCitationIDs interface{}            `json:"SourceCitationIDs,omitempty"` // Can be string (comma sep) or null
	AssertionID       string                 `json:"AssertionId,omitempty"`
}

// FactEditData represents the window.getFactEditData object embedded in source edit pages
type FactEditData struct {
	CitationID              string        `json:"CitationId"`
	SourceID                string        `json:"SourceId"`
	RepositoryID            string        `json:"RepositoryId"`
	DatabaseID              string        `json:"DatabaseId"`
	RecordID                string        `json:"RecordId"`
	CitationDate            interface{}   `json:"CitationDate"` // Can be null
	CitationPage            string        `json:"CitationPage"`
	CitationNote            string        `json:"CitationNote"`
	CitationText            string        `json:"CitationText"`
	CitationTitle           string        `json:"CitationTitle"`
	CitationURL             string        `json:"CitationUrl"`
	SourceAuthor            string        `json:"SourceAuthor"`
	SourceCallNumber        string        `json:"SourceCallNumber"`
	SourceNote              string        `json:"SourceNote"`
	SourcePublisher         string        `json:"SourcePublisher"`
	SourcePublisherDate     string        `json:"SourcePublisherDate"`
	SourcePublisherLocation string        `json:"SourcePublisherLocation"`
	SourceRefn              string        `json:"SourceRefn"`
	SourceTitle             string        `json:"SourceTitle"`
	SourceType              int           `json:"SourceType"`
	SourceFTMTemplate       bool          `json:"SourceFTMTemplate"`
	RepositoryAddress       string        `json:"RepositoryAddress"`
	RepositoryCallNumber    string        `json:"RepositoryCallNumber"`
	RepositoryEmail         string        `json:"RepositoryEmail"`
	RepositoryName          string        `json:"RepositoryName"`
	RepositoryNote          string        `json:"RepositoryNote"`
	RepositoryPhone         string        `json:"RepositoryPhone"`
	RepositoryURL           string        `json:"RepositoryUrl"`
	EditCitationDetailURL   string        `json:"EditCitationDetailUrl"`
	ErrorMessage            interface{}   `json:"ErrorMessage"` // Can be null
	FailurePoint            interface{}   `json:"FailurePoint"` // Can be null
	HasError                bool          `json:"HasError"`
	ErrorMessages           []interface{} `json:"ErrorMessages"`
	FailurePoints           []interface{} `json:"FailurePoints"`
	CanEdit                 bool          `json:"canEdit"`
	SrcFTMTemplate          bool          `json:"srcFTMTemplate"`
	CitationMedia           []interface{} `json:"citationMedia"` // Array of media items, if any
	Name                    string        `json:"name"`          // Person's name associated with the source
	Events                  []interface{} `json:"events"`        // Events associated with the source
	// Added fields for media downloading
	RecordImageUrl        string `json:"recordImageUrl,omitempty"`
	RecordImagePreviewUrl string `json:"recordImagePreviewUrl,omitempty"`
	LocalMediaFilePath    string `json:"localMediaFilePath,omitempty"`
}

// DiscoveryDiscoveryRecordResponse represents the response from the record API
type DiscoveryDiscoveryRecordResponse struct {
	DiscoveryRecords []DiscoveryRecord `json:"DiscoveryRecords"`
}

// DiscoveryRecord represents a single record
type DiscoveryRecord struct {
	DiscoveryRecordID int64           `json:"DiscoveryRecordId"`
	HouseHoldID       int64           `json:"HouseHoldId"`
	Name              string          `json:"Name"`
	DiscoveryCells    []DiscoveryCell `json:"DiscoveryCells"`
}

// DiscoveryCell represents a single cell in a record
type DiscoveryCell struct {
	X                      int                     `json:"X"`
	Y                      int                     `json:"Y"`
	Width                  int                     `json:"Width"`
	Height                 int                     `json:"Height"`
	DiscoveryDisplayFields []DiscoveryDisplayField `json:"DiscoveryDisplayFields"`
}

// DiscoveryDisplayField represents a single display field in a cell
type DiscoveryDisplayField struct {
	DisplayLabel string   `json:"DisplayLabel"`
	DisplayValue string   `json:"DisplayValue"`
	FieldNames   []string `json:"FieldNames"`
}

// CapturedRequest represents a captured network request/response
type CapturedRequest struct {
	URL          string    `json:"url"`
	Method       string    `json:"method"`
	StatusCode   int       `json:"statusCode"`
	RequestBody  string    `json:"requestBody"`
	ResponseBody string    `json:"responseBody"`
	ContentType  string    `json:"contentType"`
	Timestamp    time.Time `json:"timestamp"`
}
