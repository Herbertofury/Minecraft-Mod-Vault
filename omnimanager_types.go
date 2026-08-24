package main

import "time"

const librarySchemaVersion = 1

type LibraryHashes struct {
	SHA1             string `json:"sha1,omitempty"`
	SHA256           string `json:"sha256,omitempty"`
	SHA512           string `json:"sha512,omitempty"`
	CurseFingerprint uint32 `json:"curseFingerprint,omitempty"`
}

type LibraryDependency struct {
	ID       string `json:"id"`
	Version  string `json:"version,omitempty"`
	Kind     string `json:"kind,omitempty"`
	Required bool   `json:"required,omitempty"`
}

type LibrarySource struct {
	Provider      string   `json:"provider"`
	ProviderLabel string   `json:"providerLabel,omitempty"`
	ProjectID     string   `json:"projectId,omitempty"`
	Slug          string   `json:"slug,omitempty"`
	Title         string   `json:"title,omitempty"`
	Author        string   `json:"author,omitempty"`
	IconURL       string   `json:"iconUrl,omitempty"`
	PageURL       string   `json:"pageUrl,omitempty"`
	LatestVersion string   `json:"latestVersion,omitempty"`
	GameVersions  []string `json:"gameVersions,omitempty"`
	Loaders       []string `json:"loaders,omitempty"`
	Exact         bool     `json:"exact"`
	Confidence    float64  `json:"confidence"`
	Evidence      string   `json:"evidence"`
}

type LibraryItem struct {
	ID                   string              `json:"id"`
	Path                 string              `json:"path"`
	Filename             string              `json:"filename"`
	Name                 string              `json:"name"`
	Summary              string              `json:"summary,omitempty"`
	Description          string              `json:"description,omitempty"`
	Authors              []string            `json:"authors,omitempty"`
	Edition              string              `json:"edition"`
	Kind                 string              `json:"kind"`
	Profile              string              `json:"profile,omitempty"`
	FamilyID             string              `json:"familyId,omitempty"`
	InstalledVersion     string              `json:"installedVersion,omitempty"`
	LatestVersion        string              `json:"latestVersion,omitempty"`
	ModID                string              `json:"modId,omitempty"`
	UUID                 string              `json:"uuid,omitempty"`
	Loaders              []string            `json:"loaders,omitempty"`
	GameVersions         []string            `json:"gameVersions,omitempty"`
	MinEngineVersion     string              `json:"minEngineVersion,omitempty"`
	Modules              []string            `json:"modules,omitempty"`
	Capabilities         []string            `json:"capabilities,omitempty"`
	Dependencies         []LibraryDependency `json:"dependencies,omitempty"`
	License              string              `json:"license,omitempty"`
	Homepage             string              `json:"homepage,omitempty"`
	SourceURL            string              `json:"sourceUrl,omitempty"`
	Enabled              bool                `json:"enabled"`
	IsDir                bool                `json:"isDir"`
	Size                 int64               `json:"size"`
	Modified             string              `json:"modified"`
	Hashes               LibraryHashes       `json:"hashes,omitempty"`
	MetadataBy           string              `json:"metadataBy,omitempty"`
	LocalArtURL          string              `json:"localArtUrl,omitempty"`
	RemoteArtURL         string              `json:"remoteArtUrl,omitempty"`
	ArtOrigin            string              `json:"artOrigin,omitempty"`
	Sources              []LibrarySource     `json:"sources,omitempty"`
	UpdateStatus         string              `json:"updateStatus"`
	UpdateMessage        string              `json:"updateMessage,omitempty"`
	SafeUpdate           *UpdateCandidate    `json:"safeUpdate,omitempty"`
	Alternatives         []UpdateCandidate   `json:"alternatives,omitempty"`
	ProvenanceConfidence float64             `json:"provenanceConfidence"`
	MatchEvidence        []string            `json:"matchEvidence,omitempty"`
	Warnings             []string            `json:"warnings,omitempty"`
	ManagedRoot          string              `json:"managedRoot,omitempty"`
	ReceiptID            string              `json:"receiptId,omitempty"`
}

type LibrarySummary struct {
	Total       int `json:"total"`
	Java        int `json:"java"`
	Bedrock     int `json:"bedrock"`
	Server      int `json:"server"`
	Updates     int `json:"updates"`
	Current     int `json:"current"`
	Review      int `json:"review"`
	Unknown     int `json:"unknown"`
	Modified    int `json:"modified"`
	Disabled    int `json:"disabled"`
	WithArt     int `json:"withArt"`
	ExactSource int `json:"exactSource"`
}

type LibraryProfile struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Edition string `json:"edition"`
	Root    string `json:"root"`
	Exists  bool   `json:"exists"`
	Channel string `json:"channel,omitempty"`
}

type LibraryResponse struct {
	SchemaVersion int               `json:"schemaVersion"`
	GeneratedAt   string            `json:"generatedAt"`
	Enriched      bool              `json:"enriched"`
	Items         []LibraryItem     `json:"items"`
	Summary       LibrarySummary    `json:"summary"`
	Profiles      []LibraryProfile  `json:"profiles"`
	Warnings      []string          `json:"warnings,omitempty"`
	Errors        map[string]string `json:"errors,omitempty"`
	CacheAge      string            `json:"cacheAge,omitempty"`
}

type libraryCacheRecord struct {
	ItemID               string            `json:"itemId"`
	UpdatedAt            string            `json:"updatedAt"`
	Name                 string            `json:"name,omitempty"`
	Summary              string            `json:"summary,omitempty"`
	Authors              []string          `json:"authors,omitempty"`
	RemoteArtURL         string            `json:"remoteArtUrl,omitempty"`
	Sources              []LibrarySource   `json:"sources,omitempty"`
	LatestVersion        string            `json:"latestVersion,omitempty"`
	UpdateStatus         string            `json:"updateStatus,omitempty"`
	UpdateMessage        string            `json:"updateMessage,omitempty"`
	SafeUpdate           *UpdateCandidate  `json:"safeUpdate,omitempty"`
	Alternatives         []UpdateCandidate `json:"alternatives,omitempty"`
	ProvenanceConfidence float64           `json:"provenanceConfidence,omitempty"`
	MatchEvidence        []string          `json:"matchEvidence,omitempty"`
}

type libraryCacheFile struct {
	SchemaVersion int                           `json:"schemaVersion"`
	Records       map[string]libraryCacheRecord `json:"records"`
}

type LibraryActionRequest struct {
	Action string   `json:"action"`
	Paths  []string `json:"paths,omitempty"`
	IDs    []string `json:"ids,omitempty"`
}

type LibraryActionResult struct {
	OK       bool                 `json:"ok"`
	Action   string               `json:"action"`
	Results  []LibraryActionEntry `json:"results"`
	Receipts []string             `json:"receipts,omitempty"`
}

type LibraryActionEntry struct {
	Path      string `json:"path"`
	OK        bool   `json:"ok"`
	Result    string `json:"result,omitempty"`
	ReceiptID string `json:"receiptId,omitempty"`
	Error     string `json:"error,omitempty"`
}

type LibraryTransaction struct {
	SchemaVersion int            `json:"schemaVersion"`
	ID            string         `json:"id"`
	Action        string         `json:"action"`
	CreatedAt     string         `json:"createdAt"`
	SourcePaths   []string       `json:"sourcePaths"`
	TargetPaths   []string       `json:"targetPaths"`
	ItemNames     []string       `json:"itemNames,omitempty"`
	SHA512        []string       `json:"sha512,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	UndoneAt      string         `json:"undoneAt,omitempty"`
}

type LibraryHistoryResponse struct {
	Transactions []LibraryTransaction `json:"transactions"`
}

type BedrockInstallResult struct {
	OK             bool     `json:"ok"`
	Profile        string   `json:"profile"`
	Package        string   `json:"package"`
	InstalledPaths []string `json:"installedPaths"`
	ReceiptID      string   `json:"receiptId"`
	Kinds          []string `json:"kinds"`
}

type BedrockActivationRequest struct {
	WorldPath string `json:"worldPath"`
	PackUUID  string `json:"packUuid"`
	Version   string `json:"version"`
	PackKind  string `json:"packKind"`
}

func newLibraryResponse(items []LibraryItem, profiles []LibraryProfile, enriched bool) LibraryResponse {
	resp := LibraryResponse{SchemaVersion: librarySchemaVersion, GeneratedAt: time.Now().UTC().Format(time.RFC3339), Enriched: enriched, Items: items, Profiles: profiles, Errors: map[string]string{}}
	for _, item := range items {
		resp.Summary.Total++
		switch item.Edition {
		case "bedrock":
			resp.Summary.Bedrock++
		case "server":
			resp.Summary.Server++
		default:
			resp.Summary.Java++
		}
		switch item.UpdateStatus {
		case "update":
			resp.Summary.Updates++
		case "current":
			resp.Summary.Current++
		case "review":
			resp.Summary.Review++
		case "modified":
			resp.Summary.Modified++
		default:
			resp.Summary.Unknown++
		}
		if !item.Enabled {
			resp.Summary.Disabled++
		}
		if item.LocalArtURL != "" || item.RemoteArtURL != "" {
			resp.Summary.WithArt++
		}
		for _, source := range item.Sources {
			if source.Exact {
				resp.Summary.ExactSource++
				break
			}
		}
	}
	return resp
}
