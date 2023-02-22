package sdk_test

import (
	"os"
	"testing"
)

const (
	Identity    = "identity"
	contentType = "application/senml+json"
)

func TestMain(m *testing.M) {
	exitCode := m.Run()
	os.Exit(exitCode)
}
