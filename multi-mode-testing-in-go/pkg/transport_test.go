package pkg_test

import (
	"testing"
	"github.com/josestg/multi-mode-testing-in-go/pkg"
)


func TestTransportWithRSASigner(t *testing.T) {
	pkg.RunTestWhen(t, pkg.ModeUnit)
    t.Log("unit test: TransportWithRSASigner")
}

/* ---- using build tags
//go:build with_unit_test

package pkg_test

import (
	"testing"
)


func TestTransportWithRSASigner(t *testing.T) {
    t.Log("unit test: TransportWithRSASigner")
}
*/