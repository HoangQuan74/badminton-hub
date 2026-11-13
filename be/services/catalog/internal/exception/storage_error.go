package exception

import "errors"

var (
	ErrIdentityCutoverRequired = errors.New("Identity tenant data transfer must be verified before Catalog startup")
	ErrS3BucketRequired        = errors.New("CATALOG_S3_BUCKET is required when storage is enabled")
	ErrS3CredentialsRequired   = errors.New("both Catalog S3 access key and secret key are required")
	ErrStoredImageTooLarge     = errors.New("stored object exceeds image size limit")
)
