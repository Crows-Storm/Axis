package security

import "context"

// The DataSanitizer can Sanitize all about: password, key, private key, internal context, internal error info...
// implement it, service can build a
type DataSanitizer interface {
	sanitize(ctx context.Context)
}
