package pkg_test

import (
	"testing"
	"github.com/josestg/multi-mode-testing-in-go/pkg"
)


func TestInsertUser(t *testing.T) {
	pkg.RunTestWhen(t, pkg.ModeIntegration)
    t.Log("integration test: InsertUser")
}

/* ---- using build tags
//go:build with_integration_test

package pkg_test

import (
	"testing"
)


func InsertUser(t *testing.T) {
    t.Log("integration test: InsertUser")
}
*/