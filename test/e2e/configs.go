//go:build integration
// +build integration

package e2e

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/efficientgo/e2e"
	"github.com/efficientgo/tools/core/pkg/testutil"
)

type testType string

const (
	readWrite testType = "read_write"

	dockerLocalSharedDir = "/shared"
	configSharedDir      = "config"

	configsContainerPath = dockerLocalSharedDir + "/" + configSharedDir

	envReadWriteName = "e2e_read_write"
)

const objstoreYamlTpl = `
type: S3
config:
  bucket: %s
  access_key: %s
  secret_key: %s
  endpoint: %s
  insecure: true
`

func createObjstoreYAML(
	t *testing.T,
	e e2e.Environment,
	bucket string,
	accessKey, secretKey string,
	endpoint string,
) {
	yamlContent := []byte(fmt.Sprintf(
		objstoreYamlTpl,
		bucket,
		accessKey,
		secretKey,
		endpoint,
	))

	err := ioutil.WriteFile(
		filepath.Join(e.SharedDir(), configSharedDir, "objstore.yaml"),
		yamlContent,
		os.FileMode(0o755),
	)
	testutil.Ok(t, err)
}
