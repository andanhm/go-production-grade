package logging

import (
	log "github.com/sirupsen/logrus"
)

type Option func(*option)

// Options holds the DB configuration
type option struct {
	reference   map[string]any
	identity    string
	service     string
	errors      []error
	appVersion  int
	prettyPrint bool
	stackTrace  bool
	referenceID string
	deviceID    string
	ip          string
	platform    string // ANDROID, IOS
	skipLog     bool
	callerLevel int
	sessionID   string
	fields      log.Fields
}

// WithReference is used to set the reference for the log
func WithReference(ref map[string]any) func(*option) {
	return func(h *option) {
		h.reference = ref
	}
}

// WithIdentity is used to set the identity for the log
func WithIdentity(userID string) func(*option) {
	return func(h *option) {
		h.identity = userID
	}
}

// WithService is used to set the service for the log
func WithService(service string) func(*option) {
	return func(h *option) {
		h.service = service
	}
}

// WithTraceError keeps the original error for back tracing
func WithTraceError(err error) func(*option) {
	return func(o *option) {
		if err != nil {
			o.errors = append(o.errors, err)
		}
	}
}

// WithStackTrace is used to set the stack trace for the log
// If enable is true, the stack trace will be printed
// If enable is false, the stack trace will not be printed
func WithStackTrace(enable bool) func(*option) {
	return func(o *option) {
		o.stackTrace = enable
	}
}

// WithAppVersion is used to set the app version for the log
func WithAppVersion(code int) func(*option) {
	return func(o *option) {
		o.appVersion = code
	}
}

// WithPrettyPrint is used to set the pretty print for the log
// If enable is true, the log will be printed in pretty format
// If enable is false, the log will be printed in JSON format
func WithPrettyPrint(enable bool) func(*option) {
	return func(o *option) {
		o.prettyPrint = enable
	}
}

// WithReferenceID is used to set the reference ID for the log
func WithReferenceID(referenceID string) func(*option) {
	return func(h *option) {
		h.referenceID = referenceID
	}
}

// WithDeviceID is used to set the device ID for the log
func WithDeviceID(id string) func(*option) {
	return func(h *option) {
		h.deviceID = id
	}
}

// WithPlatform is used to set the platform for the log
func WithPlatform(platform string) func(*option) {
	return func(h *option) {
		h.platform = platform
	}
}

// WithIP is used to set the IP for the log
func WithIP(ip string) func(*option) {
	return func(h *option) {
		h.ip = ip
	}
}

// WithSkip is used to skip the log
func WithSkip(skip bool) func(*option) {
	return func(h *option) {
		h.skipLog = skip
	}
}

// WithCallerLevel is used to set the caller level for the log
// 0: caller level is 0
// 1: caller level is 1
func WithCallerLevel(level int) func(*option) {
	return func(h *option) {
		h.callerLevel = level
	}
}

// WithSessionID is used to set the session ID for the log
func WithSessionID(id string) func(*option) {
	return func(h *option) {
		h.sessionID = id
	}
}

// WithField adds a single dynamic key/value field to the log. It is the
// building block 3rd party apps use to create their own option functions,
// e.g.
//
//	func WithAppType(v string) logging.Option { return logging.WithField("appType", v) }
//	func WithInstruction(v string) logging.Option { return logging.WithField("instruction", v) }
//
// Reserved fields set by the logger are not overwritten.
func WithField(key string, value any) func(*option) {
	return func(h *option) {
		h.addField(key, value)
	}
}

func (h *option) addField(key string, value any) {
	if h.fields == nil {
		h.fields = log.Fields{}
	}
	h.fields[key] = value
}
