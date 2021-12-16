//go:build integration
// +build integration

package e2e

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/efficientgo/e2e"
	"github.com/efficientgo/tools/core/pkg/testutil"
)

// Copies static configuration to the shared directory.
func prepareConfigs(t *testing.T, tt testType, e e2e.Environment) {
	testutil.Ok(t, exec.Command("cp", "-r", "../config", filepath.Join(e.SharedDir(), configSharedDir)).Run())
}
