package migration

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuditFields_WhitelistCorrect(t *testing.T) {
	t.Parallel()
	keys := []string{"CI_BUILD_ID", "GITHUB_RUN_ID", "K8S_POD_NAME", "HOSTNAME",
		"K8S_NAMESPACE", "TRIGGERED_BY", "GIT_COMMIT_SHA"}
	for _, env := range canonicalAuditEnvs {
		assert.Contains(t, keys, env.key, "unexpected audit env key %q", env.key)
	}
}

func TestAuditFields_AbsentEnvsProduceNoFields(t *testing.T) {
	for _, v := range canonicalAuditEnvs {
		t.Setenv(v.key, "")
	}
	fields := AuditFields()
	assert.Empty(t, fields, "expected no audit fields when all envs are empty")
}

func TestAuditFields_PresentEnvsProduceFields(t *testing.T) {
	for _, v := range canonicalAuditEnvs {
		t.Setenv(v.key, "")
	}
	t.Setenv("CI_BUILD_ID", "build-123")
	t.Setenv("GIT_COMMIT_SHA", "abc123")

	fields := AuditFields()
	assert.Len(t, fields, 2)

	fieldMap := make(map[string]string, len(fields))
	for _, f := range fields {
		fieldMap[f.Key] = f.Value.String()
	}
	assert.Equal(t, "build-123", fieldMap["build_id"])
	assert.Equal(t, "abc123", fieldMap["git_commit_sha"])
}

func TestAuditFields_MixedPresence(t *testing.T) {
	for _, v := range canonicalAuditEnvs {
		t.Setenv(v.key, "")
	}
	t.Setenv("K8S_POD_NAME", "pod-1")

	fields := AuditFields()
	assert.Len(t, fields, 1)
	assert.Equal(t, slog.String("pod_name", "pod-1"), fields[0])
}
