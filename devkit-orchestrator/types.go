package main

import "time"

type Registry struct {
	Schema    int               `json:"schema"`
	Drive     DriveConfig       `json:"drive"`
	Defaults  Target            `json:"defaults"`
	Artifacts []ManagedArtifact `json:"artifacts"`
}
type DriveConfig struct {
	RootFolderID    string `json:"rootFolderId,omitempty"`
	RuntimeFolderID string `json:"runtimeFolderId,omitempty"`
	SourceFolderID  string `json:"sourceFolderId,omitempty"`
	AccessTokenEnv  string `json:"accessTokenEnv,omitempty"`
	APIBase         string `json:"apiBase,omitempty"`
	UploadBase      string `json:"uploadBase,omitempty"`
}
type Target struct {
	Minecraft string `json:"minecraft,omitempty"`
	Loader    string `json:"loader,omitempty"`
	OS        string `json:"os,omitempty"`
	Arch      string `json:"arch,omitempty"`
	Channel   string `json:"channel,omitempty"`
}
type ManagedArtifact struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Kind          string        `json:"kind"`
	Version       string        `json:"version,omitempty"`
	Filename      string        `json:"filename,omitempty"`
	SHA256        string        `json:"sha256,omitempty"`
	DriveFileID   string        `json:"driveFileId,omitempty"`
	SourceDriveID string        `json:"sourceDriveId,omitempty"`
	Target        Target        `json:"target,omitempty"`
	Providers     []ProviderRef `json:"providers"`
	Dependencies  []string      `json:"dependencies,omitempty"`
	UpdatePolicy  UpdatePolicy  `json:"updatePolicy,omitempty"`
}
type UpdatePolicy struct {
	Pinned          bool `json:"pinned,omitempty"`
	AllowPrerelease bool `json:"allowPrerelease,omitempty"`
	AllowDowngrade  bool `json:"allowDowngrade,omitempty"`
	MirrorSource    bool `json:"mirrorSource,omitempty"`
	KeepFilename    bool `json:"keepFilename,omitempty"`
}
type ProviderRef struct {
	Type       string `json:"type"`
	Project    string `json:"project,omitempty"`
	ProjectID  string `json:"projectId,omitempty"`
	Repo       string `json:"repo,omitempty"`
	Branch     string `json:"branch,omitempty"`
	AssetRegex string `json:"assetRegex,omitempty"`
	Source     bool   `json:"source,omitempty"`
	Maven      string `json:"maven,omitempty"`
	URL        string `json:"url,omitempty"`
	Priority   int    `json:"priority,omitempty"`
}
type Candidate struct {
	Provider       string       `json:"provider"`
	ProjectID      string       `json:"projectId,omitempty"`
	VersionID      string       `json:"versionId,omitempty"`
	Version        string       `json:"version"`
	Filename       string       `json:"filename"`
	URL            string       `json:"url"`
	PageURL        string       `json:"pageUrl,omitempty"`
	Published      time.Time    `json:"published,omitempty"`
	SHA256         string       `json:"sha256,omitempty"`
	SHA512         string       `json:"sha512,omitempty"`
	SHA1           string       `json:"sha1,omitempty"`
	Size           int64        `json:"size,omitempty"`
	GameVersions   []string     `json:"gameVersions,omitempty"`
	Loaders        []string     `json:"loaders,omitempty"`
	Dependencies   []Dependency `json:"dependencies,omitempty"`
	SourceURL      string       `json:"sourceUrl,omitempty"`
	SourceArchive  string       `json:"sourceArchive,omitempty"`
	SourceRef      string       `json:"sourceRef,omitempty"`
	ReleaseChannel string       `json:"releaseChannel,omitempty"`
}
type Dependency struct {
	Provider  string `json:"provider"`
	ProjectID string `json:"projectId,omitempty"`
	VersionID string `json:"versionId,omitempty"`
	FileID    int64  `json:"fileId,omitempty"`
	Required  bool   `json:"required"`
}
type ArtifactPlan struct {
	ArtifactID string     `json:"artifactId"`
	Current    LocalState `json:"current"`
	Candidate  *Candidate `json:"candidate,omitempty"`
	Status     string     `json:"status"`
	Reason     string     `json:"reason"`
	Deps       []DepPlan  `json:"dependencies,omitempty"`
}
type DepPlan struct {
	Provider  string     `json:"provider"`
	ProjectID string     `json:"projectId"`
	Candidate *Candidate `json:"candidate,omitempty"`
	Status    string     `json:"status"`
	Children  []DepPlan  `json:"dependencies,omitempty"`
}
type LocalState struct {
	Version     string `json:"version,omitempty"`
	Filename    string `json:"filename,omitempty"`
	SHA256      string `json:"sha256,omitempty"`
	DriveFileID string `json:"driveFileId,omitempty"`
}
type UpdatePlan struct {
	Schema    int            `json:"schema"`
	CreatedAt time.Time      `json:"createdAt"`
	Target    Target         `json:"target"`
	Artifacts []ArtifactPlan `json:"artifacts"`
}
type LockFile struct {
	Schema    int         `json:"schema"`
	UpdatedAt time.Time   `json:"updatedAt"`
	Entries   []LockEntry `json:"entries"`
}
type LockEntry struct {
	ArtifactID    string    `json:"artifactId"`
	Provider      string    `json:"provider"`
	ProjectID     string    `json:"projectId,omitempty"`
	VersionID     string    `json:"versionId,omitempty"`
	Version       string    `json:"version"`
	Filename      string    `json:"filename"`
	SHA256        string    `json:"sha256"`
	DriveFileID   string    `json:"driveFileId,omitempty"`
	SourceDriveID string    `json:"sourceDriveId,omitempty"`
	SourceURL     string    `json:"sourceUrl,omitempty"`
	SourceRef     string    `json:"sourceRef,omitempty"`
	ResolvedAt    time.Time `json:"resolvedAt"`
}
