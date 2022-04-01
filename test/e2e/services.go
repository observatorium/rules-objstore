//go:build integration
// +build integration

package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/efficientgo/e2e"
	e2edb "github.com/efficientgo/e2e/db"
	"github.com/efficientgo/tools/core/pkg/testutil"
)

const (
	rulesObjstoreImage = "quay.io/observatorium/rules-objstore:local_e2e_test" // Image that is built if you run `make container-test`.

	logLevelError = "error"
	logLevelDebug = "debug"
)

func newRulesObjstoreService(e e2e.Environment) (*e2e.InstrumentedRunnable, error) {
	ports := map[string]int{
		"http":          8080,
		"http-internal": 8081,
	}

	args := e2e.BuildArgs(map[string]string{
		"-web.listen":           ":" + strconv.Itoa(ports["http"]),
		"-web.internal.listen":  ":" + strconv.Itoa(ports["http-internal"]),
		"-web.healthchecks.url": "http://127.0.0.1:" + strconv.Itoa(ports["http"]),
		"-objstore.config-file": filepath.Join(configsContainerPath, "objstore.yaml"),
		"-log.level":            logLevelDebug,
	})

	return e2e.NewInstrumentedRunnable(e, "rules_objstore", ports, "http-internal").Init(
		e2e.StartOptions{
			Image:     rulesObjstoreImage,
			Command:   e2e.NewCommandWithoutEntrypoint("rules-objstore", args...),
			Readiness: e2e.NewHTTPReadinessProbe("http-internal", "/ready", 200, 200),
			User:      strconv.Itoa(os.Getuid()),
		},
	), nil
}

func newMinio(t *testing.T, e e2e.Environment) {
	// Create S3 replacement for rules backend
	bucket := "obs_rules_test"
	userID := strconv.Itoa(os.Getuid())
	ports := map[string]int{e2edb.AccessPortName: 8090}
	envVars := []string{
		"MINIO_ROOT_USER=" + e2edb.MinioAccessKey,
		"MINIO_ROOT_PASSWORD=" + e2edb.MinioSecretKey,
		"MINIO_BROWSER=" + "off",
		"ENABLE_HTTPS=" + "0",
		// https://docs.min.io/docs/minio-kms-quickstart-guide.html
		"MINIO_KMS_KES_ENDPOINT=" + "https://play.min.io:7373",
		"MINIO_KMS_KES_KEY_FILE=" + "root.key",
		"MINIO_KMS_KES_CERT_FILE=" + "root.cert",
		"MINIO_KMS_KES_KEY_NAME=" + "my-minio-key",
	}
	f := e2e.NewInstrumentedRunnable(e, "rules-minio", ports, e2edb.AccessPortName)
	runnable := f.Init(
		e2e.StartOptions{
			Image: "minio/minio:RELEASE.2022-03-03T21-21-16Z",
			// Create the required bucket before starting minio.
			Command: e2e.NewCommandWithoutEntrypoint("sh", "-c", fmt.Sprintf(
				// Hacky: Create user that matches ID with host ID to be able to remove .minio.sys details on the start.
				// Proper solution would be to contribute/create our own minio image which is non root.
				"useradd -G root -u %v me && mkdir -p %s && chown -R me %s &&"+
					"curl -sSL --tlsv1.2 -O 'https://raw.githubusercontent.com/minio/kes/master/root.key' -O 'https://raw.githubusercontent.com/minio/kes/master/root.cert' && "+
					"cp root.* /home/me/ && "+
					"su - me -s /bin/sh -c 'mkdir -p %s && %s /opt/bin/minio server --address :%v --quiet %v'",
				userID, f.InternalDir(), f.InternalDir(), filepath.Join(f.InternalDir(), bucket), strings.Join(envVars, " "), ports[e2edb.AccessPortName], f.InternalDir()),
			),
			Readiness: e2e.NewHTTPReadinessProbe(e2edb.AccessPortName, "/minio/health/live", 200, 200),
		},
	)

	testutil.Ok(t, e2e.StartAndWaitReady(runnable))

	createObjstoreYAML(t, e, bucket, e2edb.MinioAccessKey, e2edb.MinioSecretKey, runnable.InternalEndpoint(e2edb.AccessPortName))
}
