package logging

import (
	"testing"

	"github.com/go-production-grade/errors"
)

func TestLog(t *testing.T) {
	Init(
		"sc-integrated-service",
		WithPrettyPrint(true),
	)

	Info(
		"SUCCESS.INIT",
		"Database connection success...",
		WithAppVersion(100),
		WithPlatform("ANDROID"),
		WithDeviceID("6a306da5390d9eb0"),
	)
	Debug("Debug", "", WithStackTrace(true))
	Warning("Waring", "")
	err := errors.New(
		"CB001",
		"Failed",
		errors.WithData(map[string]interface{}{
			"data": "1",
		}),
	)
	Error("Error", "", WithTraceError(err))

	Critical(
		"ERR.QUERY.DB", "error",
		WithIdentity("SL20220601"),
		WithService("GOOGLE"),
		WithReference(map[string]interface{}{"data": "1"}),
		WithSessionID("1234567890"),
	)
	if err := SetLevel("DEBUG"); err != nil {
		t.Errorf("SetLevel('DEBUG') failed: %v", err)
	}
	if err := SetLevel("NO_FOUND"); err == nil {
		t.Error("SetLevel('NO_FOUND') should have returned an error")
	}
}

// WithAppType and WithInstruction are examples of options a 3rd party app
// builds on top of WithField to inject its own fields into the log.
func WithAppType(v string) Option     { return WithField("appType", v) }
func WithInstruction(v string) Option { return WithField("instruction", v) }

func TestWithField(t *testing.T) {
	opt := option{}
	WithAppType("merchant")(&opt)
	WithInstruction("settle")(&opt)
	WithField("retry", 3)(&opt)

	cases := map[string]any{
		"appType":     "merchant",
		"instruction": "settle",
		"retry":       3,
	}
	for k, want := range cases {
		if got := opt.fields[k]; got != want {
			t.Errorf("opt.fields[%q] = %v, want %v", k, got, want)
		}
	}

	// reserved fields must not be overwritten by custom fields
	Init("sc-integrated-service", WithPrettyPrint(true))
	Info("SUCCESS.INIT", "custom fields",
		WithAppType("merchant"),
		WithInstruction("settle"),
		WithField("app", "should-not-override"),
	)
}

func TestSetLevel(t *testing.T) {
	if err := SetLevel("DEBUG"); err != nil {
		t.Errorf("SetLevel('DEBUG') failed: %v", err)
	}
	if err := SetLevel("NO_FOUND"); err == nil {
		t.Error("SetLevel('NO_FOUND') should have returned an error")
	}
}
