// Package logging provides a lightweight logging library dedicated to JSON logging.
package logging

import (
	"encoding/json"
	"errors"
	"os"
	"runtime/debug"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	logLevels = map[string]log.Level{
		"DEBUG":    log.DebugLevel,
		"INFO":     log.InfoLevel,
		"WARNING":  log.WarnLevel,
		"ERROR":    log.ErrorLevel,
		"CRITICAL": log.FatalLevel,
	}
	logLevel    = log.InfoLevel
	appName     = ""
	callerLevel = 3
)

func Init(app string, opts ...Option) {
	opt := option{callerLevel: 3}
	for _, o := range opts {
		o(&opt)
	}
	log.SetFormatter(&log.JSONFormatter{
		FieldMap:    log.FieldMap{},
		PrettyPrint: opt.prettyPrint,
	})
	appName = app
	callerLevel = opt.callerLevel
}

func logging(level log.Level, code string, message string, opts ...Option) {
	if logLevel < level {
		return
	}
	opt := option{
		service:   "APP",
		reference: map[string]any{},
	}
	for _, o := range opts {
		o(&opt)
	}
	if opt.skipLog {
		return
	}
	file, line, _ := getFileAndLine(callerLevel)

	method := getMethod(callerLevel)
	host, _ := os.Hostname()
	ref, _ := json.Marshal(opt.reference)
	traces := make([]string, 0)
	if len(opt.errors) > 0 {
		for _, e := range opt.errors {
			if data, err := json.Marshal(e); err == nil {
				traces = append(traces, string(data))
				continue
			}
			traces = append(traces, e.Error())
		}
	}
	fields := log.Fields{
		"app":      appName,
		"host":     host,
		"code":     code,
		"file":     file,
		"line":     line,
		"method":   method,
		"service":  opt.service,
		"identity": opt.identity,
		"ref":      string(ref),
		"trace":    traces,
		"id":       opt.referenceID,
		"version":  opt.appVersion,
		"stack":    "",
	}
	if len(opt.deviceID) > 0 {
		fields["deviceID"] = opt.deviceID
	}
	if len(opt.platform) > 0 {
		fields["platform"] = opt.platform
	}
	if level == log.FatalLevel || opt.stackTrace {
		fields["stack"] = string(debug.Stack())
	}
	if len(opt.ip) > 0 {
		fields["ip"] = opt.ip
	}
	if len(opt.sessionID) > 0 {
		fields["sessionID"] = opt.sessionID
	}
	for k, v := range opt.fields {
		if _, exists := fields[k]; !exists {
			fields[k] = v
		}
	}
	log.WithFields(fields).Log(level, message)
}

func Error(code string, message string, opts ...Option) {
	logging(log.ErrorLevel, code, message, opts...)
}

func Info(code string, message string, opts ...Option) {
	logging(log.InfoLevel, code, message, opts...)

}

func Warning(code string, message string, opts ...Option) {
	logging(log.WarnLevel, code, message, opts...)
}

func Critical(code string, message string, opts ...Option) {
	logging(log.FatalLevel, code, message, opts...)
}

func Debug(code string, message string, opts ...Option) {
	logging(log.DebugLevel, code, message, opts...)
}

// SetLevel sets level of logging.
// level can be "CRITICAL", "ERROR", "WARNING", "INFO" or "DEBUG"
func SetLevel(level string) error {
	level = strings.ToUpper(level)
	if numLevel, ok := logLevels[level]; ok {
		logLevel = numLevel
		return nil
	}
	return errors.New("invalid log level: " + level)
}
