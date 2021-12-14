//go:build integration
// +build integration

package e2e

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/efficientgo/e2e"
	e2edb "github.com/efficientgo/e2e/db"
	"github.com/efficientgo/tools/core/pkg/runutil"
	"github.com/efficientgo/tools/core/pkg/testutil"
	rulesspec "github.com/observatorium/api/rules"
)

var sampleRules = `
groups:
  - name: test-oidc
    interval: 5s
    rules:
      - record: trs
        expr: vector(1)`

var invalidRules = `
groups:
  - name: test-oidc
    interval: 5s
    rules:
      - record: trs
        expr: vector(1)
      - invalid: property`

func TestMetricsReadAndWrite(t *testing.T) {
	t.Parallel()

	e, err := e2e.NewDockerEnvironment(envReadWriteName)
	testutil.Ok(t, err)
	t.Cleanup(e.Close)

	prepareConfigs(t, readWrite, e)

	bucket := "obs_rules_test"

	m := e2edb.NewMinio(e, "rules-minio", bucket)
	testutil.Ok(t, e2e.StartAndWaitReady(m))

	createObjstoreYAML(t, e, bucket, e2edb.MinioAccessKey, e2edb.MinioSecretKey, m.InternalEndpoint(e2edb.AccessPortName))

	rules, err := newRulesObjstoreService(e)
	testutil.Ok(t, err)
	testutil.Ok(t, e2e.StartAndWaitReady(rules))

	client, err := rulesspec.NewClient("http://" + rules.Endpoint("http"))
	testutil.Ok(t, err)

	ctx := context.Background()
	tenantA := "tenant_a"

	t.Run("valid-rules-read-write", func(t *testing.T) {
		rctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		t.Cleanup(cancel)

		// Retrying the first request as minio takes some time to get ready, even after readiness check passes.
		// Details: https://github.com/efficientgo/e2e/issues/11.
		err = runutil.Retry(time.Second*3, rctx.Done(), func() error {
			res, err := client.SetRulesWithBody(ctx, tenantA, "application/yaml", strings.NewReader(sampleRules))
			if err != nil {
				return err
			}

			if res.StatusCode/100 != 2 {
				return fmt.Errorf("statuscode expected 200, got %d", res.StatusCode)
			}

			return nil
		})
		testutil.Ok(t, err)

		resp, err := client.ListRules(ctx, tenantA)
		testutil.Ok(t, err)
		testutil.Equals(t, http.StatusOK, resp.StatusCode)

		respRules, err := ioutil.ReadAll(resp.Body)
		testutil.Ok(t, err)

		testutil.Equals(t, sampleRules, string(respRules))
	})

	t.Run("invalid-rules-read-write", func(t *testing.T) {
		res, err := client.SetRulesWithBody(ctx, tenantA, "application/yaml", strings.NewReader(invalidRules))
		testutil.Equals(t, http.StatusBadRequest, res.StatusCode)
		testutil.Ok(t, err)

		resp, err := client.ListRules(ctx, tenantA)
		testutil.Ok(t, err)
		testutil.Equals(t, http.StatusOK, resp.StatusCode)

		respRules, err := ioutil.ReadAll(resp.Body)
		testutil.Ok(t, err)

		// The rules retrieved should still match the prevoiusly set rules.
		testutil.Equals(t, sampleRules, string(respRules))
	})
}
