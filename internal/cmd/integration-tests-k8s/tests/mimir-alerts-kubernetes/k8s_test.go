package main

import (
	"testing"
	"time"

	"github.com/grafana/alloy/internal/cmd/integration-tests-k8s/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMimirAlerts(t *testing.T) {
	testDir := "./"

	cleanupFunc := util.BootstrapTest(testDir, "mimir-alerts-kubernetes")
	defer cleanupFunc()

	terminatePortFwd := util.ExecuteBackgroundCommand(
		"kubectl", []string{"port-forward", "service/mimir-nginx", "12346:80", "--namespace=mimir-test"},
		"Port forward Mimir")
	defer terminatePortFwd()

	expectedMimirConfig := `template_files:
    default_template: |-
        {{ define "__alertmanager" }}AlertManager{{ end }}
        {{ define "__alertmanagerURL" }}{{ .ExternalURL }}/#/alerts?receiver={{ .Receiver | urlquery }}{{ end }}
alertmanager_config: |
    global:
      resolve_timeout: 5m
      http_config:
        follow_redirects: true
        enable_http2: true
      smtp_hello: localhost
      smtp_require_tls: true
    route:
      receiver: "null"
      continue: false
    receivers:
    - name: "null"
    - name: alloy-namespace/global-config/myreceiver
    templates:
    - default_template
`

	require.EventuallyWithT(t, func(c *assert.CollectT) {
		actualMimirConfig := util.Curl(c, "http://localhost:12346/api/v1/alerts")
		require.Equal(c, expectedMimirConfig, actualMimirConfig)
	}, 15*time.Second, 100*time.Millisecond)

	// TODO: Try changing some k8s resources and check Mimir's config again.
}
