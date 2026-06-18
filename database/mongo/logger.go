package mongo

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
}

// Logger struct
type Logger struct {
}

func (l *Logger) Info(level int, message string, keysAndValues ...any) {
	if os.Getenv("TIER") == "production" {
		return
	}
	userID := extractIdentityFromAll(message, keysAndValues)

	ref := buildRef(level, keysAndValues)
	refBytes, _ := json.Marshal(ref)
	refStr := string(refBytes)

	fields := log.Fields{
		"level":    "info",
		"code":     "DB.QUERY",
		"service":  "mongodb",
		"identity": userID,
		"ref":      refStr,
	}
	log.WithFields(fields).Info(message)
}

// Error logs an error message with the given key/value pairs
func (l *Logger) Error(err error, message string, keysAndValues ...any) {
	if message == "" && err != nil {
		message = err.Error()
	}
	ref := buildRef(0, keysAndValues)
	if err != nil {
		ref["error"] = err.Error()
	}
	userID := extractIdentityFromAll(message, keysAndValues)

	refBytes, _ := json.Marshal(ref)
	refStr := string(refBytes)

	if isContextCancellation(message) || (err != nil && isContextCancellation(err.Error())) {
		fields := log.Fields{
			"level":    "warning",
			"code":     "ERR.DB.QUERY",
			"service":  "mongodb",
			"identity": userID,
			"ref":      refStr,
		}
		log.WithFields(fields).Warn(message)
		return
	}

	fields := log.Fields{
		"level":    "error",
		"code":     "ERR.DB.QUERY",
		"service":  "mongodb",
		"identity": userID,
		"ref":      refStr,
	}
	log.WithFields(fields).Error(message)
}

func buildRef(level int, keysAndValues []any) map[string]any {
	ref := map[string]any{
		"module": "mongo",
		"type":   "nosql",
	}
	if level > 0 {
		ref["level"] = level
	}
	data := map[string]any{}
	for i := 0; i+1 < len(keysAndValues); i += 2 {
		data[fmt.Sprintf("%v", keysAndValues[i])] = keysAndValues[i+1]
	}
	if len(data) > 0 {
		ref["data"] = data
	}
	return ref
}

var identityRegexp = regexp.MustCompile(`\b(SC[0-9]{6,})\b`)

func extractIdentity(text string) string {
	if text == "" {
		return ""
	}
	matches := identityRegexp.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractIdentityFromAll(message string, keysAndValues []any) string {
	if id := extractIdentity(message); id != "" {
		return id
	}
	for _, v := range keysAndValues {
		if id := extractIdentity(fmt.Sprintf("%v", v)); id != "" {
			return id
		}
	}
	return ""
}

func isContextCancellation(msg string) bool {
	if msg == "" {
		return false
	}
	msg = strings.ToLower(msg)
	return strings.Contains(msg, "context canceled") ||
		strings.Contains(msg, "context deadline exceeded") ||
		strings.Contains(msg, "timeout")
}
