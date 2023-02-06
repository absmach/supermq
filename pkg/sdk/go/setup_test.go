package sdk_test

import (
	"os"
	"testing"
)

const (
	Identity = "identity"
	token    = "token"
)

var (
	limit  uint64 = 5
	offset uint64 = 0
	total  uint64 = 200
)

func TestMain(m *testing.M) {
	exitCode := m.Run()
	os.Exit(exitCode)
}
