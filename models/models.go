package models

import "time"

type Version struct {
	Number string `json:"number"`
}

type InRequest struct {
	Source  Source   `json:"source"`
	Version Version  `json:"version"`
	Params  InParams `json:"params"`
}

type InResponse struct {
	Version  Version  `json:"version"`
	Metadata Metadata `json:"metadata"`
}

type InParams struct{}

type OutRequest struct {
	Source  Source    `json:"source"`
	Version Version   `json:"version"`
	Params  OutParams `json:"params"`
}

type OutResponse struct {
	Version  Version  `json:"version"`
	Metadata Metadata `json:"metadata"`
}

type OutParams struct {
	Action StatusAction `json:"action"`
}

type CheckRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
}

type CheckResponse []Version

type Source struct {
	Debug          string `json:"debug"`
	Driver         Driver `json:"driver"`
	InitialVersion string `json:"initial_version"`
	RequireReady   bool   `json:"require_ready"`
	RetryAfter     string `json:"retry_after"`

	Bucket               string `json:"bucket"`
	Key                  string `json:"key"`
	AccessKeyID          string `json:"access_key_id"`
	SecretAccessKey      string `json:"secret_access_key"`
	SessionToken		 string `json:"session_token"`
	UseIAMInstanceProfile bool `json:"use_iam_instance_profile"`
	RegionName           string `json:"region_name"`
	Endpoint             string `json:"endpoint"`
	DisableSSL           bool   `json:"disable_ssl"`
	ServerSideEncryption string `json:"server_side_encryption"`
	UseV2Signing         bool   `json:"use_v2_signing"`

	URI        string `json:"uri"`
	Branch     string `json:"branch"`
	PrivateKey string `json:"private_key"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	File       string `json:"file"`
	GitUser    string `json:"git_user"`

	OpenStack OpenStackOptions `json:"openstack"`

	JSONKey string `json:"json_key"`
}

// OpenStackOptions contains properties for authenticating and accessing
// the object storage system.
type OpenStackOptions struct {
	Container string `json:"container"`
	ItemName  string `json:"item_name"`
	Region    string `json:"region"`

	// Properties below are for authentication. Its a copy of
	// the properties required by gophercloud. Review documentation
	// in gophercloud for parameter usage as these are just passed in.
	IdentityEndpoint string `json:"identity_endpoint"`
	Username         string `json:"username"`
	UserID           string `json:"user_id"`
	Password         string `json:"password"`
	APIKey           string `json:"api_key"`
	DomainID         string `json:"domain_id"`
	DomainName       string `json:"domain_name"`
	TenantID         string `json:"tenant_id"`
	TenantName       string `json:"tenant_name"`
	AllowReauth      bool   `json:"allow_reauth"`
	TokenID          string `json:"token_id"`
}

type Metadata []MetadataField

type MetadataField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type BuildFailure struct {
	JobName    string `yaml:"job"`
	BuildName  string `yaml:"build"`
	DetailsURL string `yaml:"details"`
}

type PipelineStatus struct {
	Pipeline     string        `yaml:"pipeline"`
	Team         string        `yaml:"team"`
	BuildNumber  string        `yaml:"build"`
	LastModified string        `yaml:"last_modified"`
	State        PipelineState `yaml:"state"`
	Failure      *BuildFailure `yaml:"failure,omitempty"`
}

type Driver string
type PipelineState string
type StatusAction string

const (
	DriverUnspecified Driver = ""
	DriverS3          Driver = "s3"
	DriverGit         Driver = "git"
	DriverSwift       Driver = "swift"
	DriverGCS         Driver = "gcs"
)

const (
	StateReady   PipelineState = "READY"
	StateRunning PipelineState = "RUNNING"
)

const (
	Start  StatusAction = "start"
	Finish StatusAction = "finish"
	Fail   StatusAction = "fail"
)

const (
	DefaultRetryPeriod time.Duration = 1 * time.Minute
)

const (
	ISO8601DateFormat string = "2006-01-02T15:04:05-0700"
)
