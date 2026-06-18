package database

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockDB struct {
}

func (d MockDB) URIBuilder() string {

	return ""
}

func (d MockDB) Name() string {
	return ""
}

func (d MockDB) EnableLogging() bool {
	return true
}

func Test_New(t *testing.T) {
	os.Setenv("TIER", "dev")
	_, err := New(MockDB{})
	assert.Error(t, err)
	log := Logger{}
	log.Printf("Slow Query %s", `INSERT INTO "SC_USER (user_id) VALUES ('SC012202601)"`)
	os.Unsetenv("TIER")
}
