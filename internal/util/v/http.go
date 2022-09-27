package v

import "net/textproto"

var (
	HeaderAccountID     = textproto.CanonicalMIMEHeaderKey("X-Account-ID")
	HeaderUserID        = textproto.CanonicalMIMEHeaderKey("X-User-ID")
	HeaderRealAccountID = textproto.CanonicalMIMEHeaderKey("X-Real-Account-ID")
	HeaderRealUserID    = textproto.CanonicalMIMEHeaderKey("X-Real-User-ID")
	HeaderSource        = textproto.CanonicalMIMEHeaderKey("X-Source")

	HeaderTraceID = textproto.CanonicalMIMEHeaderKey("X-Trace-ID")

	HeaderCacheControl = textproto.CanonicalMIMEHeaderKey("Cache-Control")
)
