package migration

import (
	"log/slog"
	"os"
)

type envVar struct {
	key   string
	field string
}

type auditEnvs []envVar

func (e auditEnvs) Fields() []slog.Attr {
	attrs := make([]slog.Attr, 0, len(e))
	for _, v := range e {
		val := os.Getenv(v.key)
		if val == "" {
			continue
		}
		attrs = append(attrs, slog.String(v.field, val))
	}
	return attrs
}

var canonicalAuditEnvs = auditEnvs{
	{key: "CI_BUILD_ID", field: "build_id"},
	{key: "GITHUB_RUN_ID", field: "github_run_id"},
	{key: "K8S_POD_NAME", field: "pod_name"},
	{key: "HOSTNAME", field: "hostname"},
	{key: "K8S_NAMESPACE", field: "k8s_namespace"},
	{key: "TRIGGERED_BY", field: "triggered_by"},
	{key: "GIT_COMMIT_SHA", field: "git_commit_sha"},
}

// AuditFields returns slog.Attr entries for each non-empty audit env variable.
func AuditFields() []slog.Attr {
	return canonicalAuditEnvs.Fields()
}
