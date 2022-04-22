package pkg

import (
	"os"
	"testing"
)

const (
	ModeAll = "all"
	ModeUnit = "unit"
	ModeIntegration = "integration"
)

func RunTestWhen(t *testing.T, mode string) {
	envMode := os.Getenv("TEST_MODE")
	if envMode == "" || envMode == ModeAll {
		return
	}

	if envMode != mode {
		t.Skip()
	}
}