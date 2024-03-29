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
	"github.com/prometheus/prometheus/pkg/rulefmt"
)

var sampleRulesA = `
groups:
  - name: test-oidc
    interval: 5s
    rules:
      - record: trs
        expr: vector(1)
      - alert: HighRequestLatency
        expr: job:request_latency_seconds:mean5m{job="myjob"} > 0.5
        for: 10m
        labels:
          severity: page
        annotations:
          summary: High request latency`

var sampleRulesB = `
groups:
  - name: test-oidc
    interval: 5s
    rules:
      - record: btrs
        expr: vector(1)
        labels:
          dummy: yes
      - alert: HighRequestLatency
        expr: job:request_latency_seconds:mean5m{job="second"} > 0.5
        for: 10m`

var allRulesMerged = `groups:
    - name: tenant_a.test-oidc
      interval: 5s
      rules:
        - record: trs
          expr: vector(1)
        - alert: HighRequestLatency
          expr: job:request_latency_seconds:mean5m{job="myjob"} > 0.5
          for: 10m
          labels:
            severity: page
          annotations:
            summary: High request latency
    - name: tenant_b.test-oidc
      interval: 5s
      rules:
        - record: btrs
          expr: vector(1)
          labels:
            dummy: "yes"
        - alert: HighRequestLatency
          expr: job:request_latency_seconds:mean5m{job="second"} > 0.5
          for: 10m
`

var invalidRules = `
groups:
  - name: test-oidc
    interval: 5s
    rules:
      - record: trs
        expr: vector(1)
      - invalid: property`

var emptyRules = `
groups: []`

var nilRules = ``

func TestRulesReadAndWrite(t *testing.T) {
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
	tenantB := "tenant_b"

	t.Run("valid-rules-read-write", func(t *testing.T) {
		rctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		t.Cleanup(cancel)

		// Retrying the first request as minio takes some time to get ready, even after readiness check passes.
		// Details: https://github.com/efficientgo/e2e/issues/11.
		testutil.Ok(t, runutil.Retry(time.Second*3, rctx.Done(), func() error {
			res, err := client.SetRulesWithBody(ctx, tenantA, "application/yaml", strings.NewReader(sampleRulesA))
			if err != nil {
				return err
			}

			if res.StatusCode/100 != 2 {
				return fmt.Errorf("statuscode expected 200, got %d", res.StatusCode)
			}

			return nil
		}))

		res, err := client.SetRulesWithBody(ctx, tenantB, "application/yaml", strings.NewReader(sampleRulesB))
		testutil.Ok(t, err)
		testutil.Equals(t, http.StatusOK, res.StatusCode)

		checkRules(t, ctx, client, tenantA, sampleRulesA)
		checkRules(t, ctx, client, tenantB, sampleRulesB)
	})

	t.Run("invalid-rules-read-write", func(t *testing.T) {
		res, err := client.SetRulesWithBody(ctx, tenantA, "application/yaml", strings.NewReader(invalidRules))
		testutil.Equals(t, http.StatusBadRequest, res.StatusCode)
		testutil.Ok(t, err)

		// The rules retrieved should still match the prevoiusly set rules.
		checkRules(t, ctx, client, tenantA, sampleRulesA)
	})

	t.Run("empty-rules-read-write", func(t *testing.T) {
		res, err := client.SetRulesWithBody(ctx, tenantA, "application/yaml", strings.NewReader(emptyRules))
		testutil.Equals(t, http.StatusOK, res.StatusCode)
		testutil.Ok(t, err)

		checkRules(t, ctx, client, tenantA, emptyRules)
	})

	t.Run("nil-rules-read-write", func(t *testing.T) {
		res, err := client.SetRulesWithBody(ctx, tenantA, "application/yaml", strings.NewReader(nilRules))
		testutil.Equals(t, http.StatusOK, res.StatusCode)
		testutil.Ok(t, err)

		checkRules(t, ctx, client, tenantA, nilRules)
	})

	t.Run("all-rules", func(t *testing.T) {
		res, err := client.ListAllRules(ctx)
		testutil.Ok(t, err)

		respRules, err := ioutil.ReadAll(res.Body)
		testutil.Ok(t, err)

		_, errs := rulefmt.Parse(respRules)
		testutil.Equals(t, 0, len(errs))

		testutil.Equals(t, allRulesMerged, string(respRules))
	})
}

func checkRules(t *testing.T, ctx context.Context, client *rulesspec.Client, tenant, rules string) {
	resp, err := client.ListRules(ctx, tenant)
	testutil.Ok(t, err)
	testutil.Equals(t, http.StatusOK, resp.StatusCode)

	respRules, err := ioutil.ReadAll(resp.Body)
	testutil.Ok(t, err)

	testutil.Equals(t, rules, string(respRules))
}
