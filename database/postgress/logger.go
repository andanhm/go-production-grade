package database

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
}

// Logger struct
type Logger struct{}

// Printf - GORM-compatible log formatter
// GORM calls: Printf(format string, values ...interface{})
func (l *Logger) Printf(format string, queries ...any) {
	// minimal guard
	if format == "" && len(queries) == 0 {
		return
	}

	// Safely build the formatted message using the format string GORM provides.
	// If format itself is not a format string, Sprintf will simply return it.
	// Protect against panics if the format and args don't match perfectly
	// (Sprintf itself won't panic but it's safer to try/catch complex types)
	formatted := fmt.Sprintf(format, queries...)

	// Build reference for structured logging
	ref := map[string]any{
		"module": "gorm",
		"type":   "sql",
		"msg":    formatted,
		"raw":    queries, // raw args from GORM (be mindful of serialization)
	}

	// If the formatted message contains this particular unique constraint error, ignore it
	// (example: "SC_USER_CONSENT_user_id_key" and Postgres unique violation code 23505)
	if strings.Contains(formatted, "23505") && strings.Contains(formatted, "SC_USER_CONSENT_user_id_key") {
		return
	}

	// Try to extract a user identity token from the formatted string first,
	// otherwise from the individual query args
	identity := extractIdentity(formatted)
	if identity == "" {
		for _, q := range queries {
			s := fmt.Sprintf("%v", q)
			identity = extractIdentity(s)
			if identity != "" {
				break
			}
		}
	}

	ref["result"] = queries

	refBytes, _ := json.Marshal(ref)
	refStr := string(refBytes)

	// Decide log level
	if isContextCancellation(formatted) {
		fields := log.Fields{
			"level":    "warning",
			"code":     "ERR.DB.QUERY",
			"service":  "postgres",
			"identity": identity,
			"ref":      refStr,
		}
		log.WithFields(fields).Warn(formatted)
		return
	}

	fields := log.Fields{
		"level":    "error",
		"code":     "ERR.DB.QUERY",
		"service":  "postgres",
		"identity": identity,
		"ref":      refStr,
	}
	log.WithFields(fields).Error(formatted)
}

// extractIdentity finds likely identity tokens like SC123456 or SC_SOMETHING etc.
// Adjust the regex to match the tokens you actually have in logs.
// This pattern finds:
//   - SC followed by digits (e.g. SC800000067321)
//   - SC_... tokens (e.g. SC_REWARD_...)
//
// It returns the first match.
func extractIdentity(text string) string {
	if text == "" {
		return ""
	}
	// Combined patterns:
	//  - SC followed by 6+ digits
	//  - SC_ followed by word characters (letters/digits/underscore)
	re := regexp.MustCompile(`\b(SC[0-9]{6,})\b`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
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
