//go:build integration
// +build integration

package e2e

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/efficientgo/e2e"
)

const (
	rulesObjstoreImage = "quay.io/observatorium/rules-objstore:local_e2e_test" // Image that is built if you run `make container-test`.

	logLevelError = "error"
	logLevelDebug = "debug"
)

func newRulesObjstoreService(e e2e.Environment) (e2e.InstrumentedRunnable, error) {
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

	return e2e.NewInstrumentedRunnable(e, "rules_objstore").WithPorts(ports, "http-internal").Init(e2e.StartOptions{
		Image:     rulesObjstoreImage,
		Command:   e2e.NewCommandWithoutEntrypoint("rules-objstore", args...),
		Readiness: e2e.NewHTTPReadinessProbe("http-internal", "/ready", 200, 200),
		User:      strconv.Itoa(os.Getuid()),
	}), nil
}
