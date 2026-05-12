package mssql

import (
	"regexp"
	"strings"
	"testing"
)

// BUG-DEL-001 regression: softDeleteInstallmentsByTransaction must NOT filter
// the parent transaction by t.[DeletedAt] IS NULL. DeleteTransaction.Execute
// soft-deletes the parent row first inside the same database.Do transaction,
// so a DeletedAt-IS-NULL predicate on the joined Transactions row would
// exclude the just-soft-deleted parent and the UPDATE would affect zero
// installments — silently leaving them active (RF-44/RF-55 regression).
func TestSoftDeleteInstallmentsByTransaction_DoesNotFilterParentByDeletedAt(t *testing.T) {
	q := softDeleteInstallmentsByTransaction

	parentDeletedAtFilter := regexp.MustCompile(`(?i)\bt\s*\.\s*\[?DeletedAt\]?\s+IS\s+NULL`)
	if parentDeletedAtFilter.MatchString(q) {
		t.Fatalf("softDeleteInstallmentsByTransaction must not filter the joined parent by DeletedAt IS NULL — the parent is soft-deleted earlier in the same TX (RF-44/RF-55). Query:\n%s", q)
	}

	// Sanity: child predicate stays in place and the user-id auth join is preserved.
	if !strings.Contains(q, "i.[DeletedAt] IS NULL") {
		t.Error("softDeleteInstallmentsByTransaction must still skip already soft-deleted installments via i.[DeletedAt] IS NULL")
	}
	if !regexp.MustCompile(`(?i)t\.\[UserId\]\s*=\s*@userId`).MatchString(q) {
		t.Error("softDeleteInstallmentsByTransaction must enforce user authorization via t.[UserId] = @userId")
	}
}
