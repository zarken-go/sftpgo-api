package api

import "time"

// FilesystemProvider defines the supported storages
type FilesystemProvider int

// supported values for FilesystemProvider
const (
	LocalFilesystemProvider     FilesystemProvider = iota // Local
	S3FilesystemProvider                                  // AWS S3 compatible
	GCSFilesystemProvider                                 // Google Cloud Storage
	AzureBlobFilesystemProvider                           // Azure Blob Storage
	CryptedFilesystemProvider                             // Local encrypted
	SFTPFilesystemProvider                                // SFTP
)

// Filesystem defines cloud storage filesystem details
type Filesystem struct {
	Provider FilesystemProvider `json:"provider"`
	S3Config S3FsConfig         `json:"s3config,omitempty"`
	// TODO: GCSConfig    vfs.GCSFsConfig    `json:"gcsconfig,omitempty"`
	// TODO: AzBlobConfig vfs.AzBlobFsConfig `json:"azblobconfig,omitempty"`
	// TODO: CryptConfig  vfs.CryptFsConfig  `json:"cryptconfig,omitempty"`
	// TODO: SFTPConfig   vfs.SFTPFsConfig   `json:"sftpconfig,omitempty"`
}

// ExtensionsFilter defines filters based on file extensions.
// These restrictions do not apply to files listing for performance reasons, so
// a denied file cannot be downloaded/overwritten/renamed but will still be
// in the list of files.
// System commands such as Git and rsync interacts with the filesystem directly
// and they are not aware about these restrictions so they are not allowed
// inside paths with extensions filters
type ExtensionsFilter struct {
	// Virtual path, if no other specific filter is defined, the filter apply for
	// sub directories too.
	// For example if filters are defined for the paths "/" and "/sub" then the
	// filters for "/" are applied for any file outside the "/sub" directory
	Path string `json:"path"`
	// only files with these, case insensitive, extensions are allowed.
	// Shell like expansion is not supported so you have to specify ".jpg" and
	// not "*.jpg". If you want shell like patterns use pattern filters
	AllowedExtensions []string `json:"allowed_extensions,omitempty"`
	// files with these, case insensitive, extensions are not allowed.
	// Denied file extensions are evaluated before the allowed ones
	DeniedExtensions []string `json:"denied_extensions,omitempty"`
}

// PatternsFilter defines filters based on shell like patterns.
// These restrictions do not apply to files listing for performance reasons, so
// a denied file cannot be downloaded/overwritten/renamed but will still be
// in the list of files.
// System commands such as Git and rsync interacts with the filesystem directly
// and they are not aware about these restrictions so they are not allowed
// inside paths with extensions filters
type PatternsFilter struct {
	// Virtual path, if no other specific filter is defined, the filter apply for
	// sub directories too.
	// For example if filters are defined for the paths "/" and "/sub" then the
	// filters for "/" are applied for any file outside the "/sub" directory
	Path string `json:"path"`
	// files with these, case insensitive, patterns are allowed.
	// Denied file patterns are evaluated before the allowed ones
	AllowedPatterns []string `json:"allowed_patterns,omitempty"`
	// files with these, case insensitive, patterns are not allowed.
	// Denied file patterns are evaluated before the allowed ones
	DeniedPatterns []string `json:"denied_patterns,omitempty"`
}

// UserFilters defines additional restrictions for a user
type UserFilters struct {
	// only clients connecting from these IP/Mask are allowed.
	// IP/Mask must be in CIDR notation as defined in RFC 4632 and RFC 4291
	// for example "192.0.2.0/24" or "2001:db8::/32"
	AllowedIP []string `json:"allowed_ip,omitempty"`
	// clients connecting from these IP/Mask are not allowed.
	// Denied rules will be evaluated before allowed ones
	DeniedIP []string `json:"denied_ip,omitempty"`
	// these login methods are not allowed.
	// If null or empty any available login method is allowed
	DeniedLoginMethods []string `json:"denied_login_methods,omitempty"`
	// these protocols are not allowed.
	// If null or empty any available protocol is allowed
	DeniedProtocols []string `json:"denied_protocols,omitempty"`
	// filters based on file extensions.
	// Please note that these restrictions can be easily bypassed.
	FileExtensions []ExtensionsFilter `json:"file_extensions,omitempty"`
	// filter based on shell patterns
	FilePatterns []PatternsFilter `json:"file_patterns,omitempty"`
	// max size allowed for a single upload, 0 means unlimited
	MaxUploadFileSize int64 `json:"max_upload_file_size,omitempty"`
}

type Users []User

// User defines a SFTPGo user
type User struct {
	// Database unique identifier
	ID int64 `json:"id"`
	// 1 enabled, 0 disabled (login is not allowed)
	Status int `json:"status"`
	// Username
	Username string `json:"username"`
	// Account expiration date as unix timestamp in milliseconds. An expired account cannot login.
	// 0 means no expiration
	ExpirationDate int64 `json:"expiration_date"`
	// Password used for password authentication.
	// For users created using SFTPGo REST API the password is be stored using argon2id hashing algo.
	// Checking passwords stored with bcrypt, pbkdf2, md5crypt and sha512crypt is supported too.
	Password string `json:"password,omitempty"`
	// PublicKeys used for public key authentication. At least one between password and a public key is mandatory
	PublicKeys []string `json:"public_keys,omitempty"`
	// The user cannot upload or download files outside this directory. Must be an absolute path
	HomeDir string `json:"home_dir"`
	// Mapping between virtual paths and filesystem paths outside the home directory.
	// Supported for local filesystem only
	// TODO: VirtualFolders []vfs.VirtualFolder `json:"virtual_folders,omitempty"`
	// If sftpgo runs as root system user then the created files and directories will be assigned to this system UID
	UID int `json:"uid"`
	// If sftpgo runs as root system user then the created files and directories will be assigned to this system GID
	GID int `json:"gid"`
	// Maximum concurrent sessions. 0 means unlimited
	MaxSessions int `json:"max_sessions"`
	// Maximum size allowed as bytes. 0 means unlimited
	QuotaSize int64 `json:"quota_size"`
	// Maximum number of files allowed. 0 means unlimited
	QuotaFiles int `json:"quota_files"`
	// List of the granted permissions
	Permissions map[string][]string `json:"permissions"`
	// Used quota as bytes
	UsedQuotaSize int64 `json:"used_quota_size"`
	// Used quota as number of files
	UsedQuotaFiles int `json:"used_quota_files"`
	// Last quota update as unix timestamp in milliseconds
	LastQuotaUpdate int64 `json:"last_quota_update"`
	// Maximum upload bandwidth as KB/s, 0 means unlimited
	UploadBandwidth int64 `json:"upload_bandwidth"`
	// Maximum download bandwidth as KB/s, 0 means unlimited
	DownloadBandwidth int64 `json:"download_bandwidth"`
	// Last login as unix timestamp in milliseconds
	LastLogin int64 `json:"last_login"`
	// Additional restrictions
	Filters UserFilters `json:"filters"`
	// Filesystem configuration details
	FsConfig Filesystem `json:"filesystem"`
	// free form text field for external systems
	AdditionalInfo string `json:"additional_info,omitempty"`
}

// S3FsConfig defines the configuration for S3 based filesystem
type S3FsConfig struct {
	Bucket string `json:"bucket,omitempty"`
	// KeyPrefix is similar to a chroot directory for local filesystem.
	// If specified then the SFTP user will only see objects that starts
	// with this prefix and so you can restrict access to a specific
	// folder. The prefix, if not empty, must not start with "/" and must
	// end with "/".
	// If empty the whole bucket contents will be available
	KeyPrefix    string `json:"key_prefix,omitempty"`
	Region       string `json:"region,omitempty"`
	AccessKey    string `json:"access_key,omitempty"`
	AccessSecret Secret `json:"access_secret,omitempty"`
	Endpoint     string `json:"endpoint,omitempty"`
	StorageClass string `json:"storage_class,omitempty"`
	// The buffer size (in MB) to use for multipart uploads. The minimum allowed part size is 5MB,
	// and if this value is set to zero, the default value (5MB) for the AWS SDK will be used.
	// The minimum allowed value is 5.
	// Please note that if the upload bandwidth between the SFTP client and SFTPGo is greater than
	// the upload bandwidth between SFTPGo and S3 then the SFTP client have to wait for the upload
	// of the last parts to S3 after it ends the file upload to SFTPGo, and it may time out.
	// Keep this in mind if you customize these parameters.
	UploadPartSize int64 `json:"upload_part_size,omitempty"`
	// How many parts are uploaded in parallel
	UploadConcurrency int `json:"upload_concurrency,omitempty"`
}

// SecretStatus defines the statuses of a Secret object
type SecretStatus = string

const (
	// SecretStatusPlain means the secret is in plain text and must be encrypted
	SecretStatusPlain SecretStatus = "Plain"
	// SecretStatusAES256GCM means the secret is encrypted using AES-256-GCM
	SecretStatusAES256GCM SecretStatus = "AES-256-GCM"
	// SecretStatusSecretBox means the secret is encrypted using a locally provided symmetric key
	SecretStatusSecretBox SecretStatus = "Secretbox"
	// SecretStatusGCP means we use keys from Google Cloud Platform’s Key Management Service
	// (GCP KMS) to keep information secret
	SecretStatusGCP SecretStatus = "GCP"
	// SecretStatusAWS means we use customer master keys from Amazon Web Service’s
	// Key Management Service (AWS KMS) to keep information secret
	SecretStatusAWS SecretStatus = "AWS"
	// SecretStatusVaultTransit means we use the transit secrets engine in Vault
	// to keep information secret
	SecretStatusVaultTransit SecretStatus = "VaultTransit"
	// SecretStatusRedacted means the secret is redacted
	SecretStatusRedacted SecretStatus = "Redacted"
)

type Secret struct {
	Status         SecretStatus `json:"status,omitempty"`
	Payload        string       `json:"payload,omitempty"`
	Key            string       `json:"key,omitempty"`
	AdditionalData string       `json:"additional_data,omitempty"`
	Mode           int          `json:"mode,omitempty"`
}

// ConnectionStatus returns the status for an active connection
type ConnectionStatus struct {
	// Logged in username
	Username string `json:"username,omitempty"`
	// Unique identifier for the connection
	ConnectionID string `json:"connection_id,omitempty"`
	// client's version string
	ClientVersion string `json:"client_version,omitempty"`
	// Remote address for this connection
	RemoteAddress string `json:"remote_address,omitempty"`
	// Connection time as unix timestamp in milliseconds
	ConnectionTime int64 `json:"connection_time,omitempty"`
	// Last activity as unix timestamp in milliseconds
	LastActivity int64 `json:"last_activity,omitempty"`
	// Protocol for this connection
	Protocol string `json:"protocol,omitempty"`
	// active uploads/downloads
	Transfers []ConnectionTransfer `json:"active_transfers,omitempty"`
	// SSH command or WebDAV method
	Command string `json:"command,omitempty"`
}

type ConnectionTransfer struct {
	OperationType string `json:"operation_type,omitempty"`
	StartTime     int64  `json:"start_time,omitempty"`
	Size          int64  `json:"size,omitempty"`
	VirtualPath   string `json:"path,omitempty"`
}

type UserQuotaScans []UserQuotaScan
type UserQuotaScan struct {
	Username string `json:"username"`
	Start    int64  `json:"start_time"`
}

func (Scan UserQuotaScan) Time() time.Time {
	return time.Unix(0, Scan.Start*1000000)
}

type UserFilterFunc func(user User) bool

func (users Users) Filter(f UserFilterFunc) Users {
	var Results Users
	for _, User := range users {
		if f(User) {
			Results = append(Results, User)
		}
	}
	return Results
}
