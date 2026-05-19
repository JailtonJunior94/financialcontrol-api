package redactor_test

import (
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/redactor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- helpers ----------------------------------------------------------------

type kv struct{ key, val string }

func (a kv) Key() string { return a.key }

// ---- Contains ----------------------------------------------------------------

func TestContains(t *testing.T) {
	dl := redactor.MustCompile("password", "cpf", "card_number")

	cases := []struct {
		name string
		key  string
		want bool
	}{
		{"exact match", "password", true},
		{"uppercase", "PASSWORD", true},
		{"mixed case", "Cpf", true},
		{"not listed", "name", false},
		{"empty", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, dl.Contains(tc.key))
		})
	}
}

// ---- RedactMap ---------------------------------------------------------------

func TestRedactMap_RootPII(t *testing.T) {
	dl := redactor.MustCompile("password", "cpf", "card_number")

	src := map[string]any{
		"name":        "Alice",
		"password":    "s3cr3t",
		"cpf":         "123.456.789-00",
		"card_number": "4111111111111111",
	}

	got := dl.RedactMap(src)

	assert.Equal(t, "Alice", got["name"])
	assert.Equal(t, redactor.Sentinel, got["password"])
	assert.Equal(t, redactor.Sentinel, got["cpf"])
	assert.Equal(t, redactor.Sentinel, got["card_number"])
}

func TestRedactMap_NestedThreeLevels(t *testing.T) {
	// "location" and "profile" are not PII keys; "zipcode" and "cpf" are.
	// This tests redaction at depth ≥ 3: user → profile → location → zipcode.
	dl := redactor.DefaultDenylist

	src := map[string]any{
		"user": map[string]any{
			"profile": map[string]any{
				"location": map[string]any{
					"zipcode": "01310-100",
					"street":  "Av. Paulista",
				},
				"cpf": "123.456.789-00",
			},
		},
	}

	got := dl.RedactMap(src)

	user := got["user"].(map[string]any)
	profile := user["profile"].(map[string]any)
	location := profile["location"].(map[string]any)

	assert.Equal(t, redactor.Sentinel, location["zipcode"])
	assert.Equal(t, "Av. Paulista", location["street"])
	assert.Equal(t, redactor.Sentinel, profile["cpf"])
}

func TestRedactMap_CapitalisationVariants(t *testing.T) {
	dl := redactor.MustCompile("cpf")

	cases := []struct{ key string }{
		{"cpf"},
		{"CPF"},
		{"Cpf"},
		{"cPF"},
	}

	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			src := map[string]any{tc.key: "123.456.789-00"}
			got := dl.RedactMap(src)
			assert.Equal(t, redactor.Sentinel, got[tc.key])
		})
	}
}

func TestRedactMap_NoPII_DeepEqual(t *testing.T) {
	dl := redactor.MustCompile("password", "cpf")

	src := map[string]any{
		"name":   "Bob",
		"age":    42,
		"active": true,
	}

	got := dl.RedactMap(src)
	assert.Equal(t, src, got)
}

func TestRedactMap_ArrayOfObjects(t *testing.T) {
	dl := redactor.MustCompile("card_number")

	src := map[string]any{
		"transactions": []any{
			map[string]any{"id": "t1", "card_number": "4111111111111111"},
			map[string]any{"id": "t2", "card_number": "5500005555555559"},
		},
	}

	got := dl.RedactMap(src)
	txs := got["transactions"].([]any)

	for _, tx := range txs {
		m := tx.(map[string]any)
		assert.Equal(t, redactor.Sentinel, m["card_number"])
		assert.NotEmpty(t, m["id"])
	}
}

func TestRedactMap_NilInput(t *testing.T) {
	dl := redactor.MustCompile("password")
	assert.Nil(t, dl.RedactMap(nil))
}

func TestRedactMap_EmptyDenylist(t *testing.T) {
	dl := redactor.MustCompile()
	src := map[string]any{"password": "s3cr3t"}
	got := dl.RedactMap(src)
	assert.Equal(t, "s3cr3t", got["password"])
}

// ---- RedactStatement --------------------------------------------------------

func TestRedactStatement(t *testing.T) {
	dl := redactor.MustCompile("cpf", "password", "cvv")

	cases := []struct {
		name  string
		input string
		check func(t *testing.T, got string)
	}{
		{
			name:  "single-quoted value",
			input: "UPDATE users SET cpf='123.456.789-00' WHERE id=1",
			check: func(t *testing.T, got string) {
				assert.Contains(t, got, "cpf="+redactor.Sentinel)
				assert.NotContains(t, got, "123.456.789-00")
			},
		},
		{
			name:  "unquoted value",
			input: "UPDATE users SET password=s3cr3t WHERE id=1",
			check: func(t *testing.T, got string) {
				assert.Contains(t, got, "password="+redactor.Sentinel)
				assert.NotContains(t, got, "s3cr3t")
			},
		},
		{
			name:  "parameter placeholder",
			input: "UPDATE cards SET cvv=@p1 WHERE id=@p2",
			check: func(t *testing.T, got string) {
				assert.Contains(t, got, "cvv="+redactor.Sentinel)
				assert.NotContains(t, got, "@p1")
			},
		},
		{
			name:  "column name without assignment is preserved",
			input: "SELECT cpf FROM users WHERE cpf='123'",
			check: func(t *testing.T, got string) {
				// The cpf after SELECT must not be touched (no = follows)
				assert.Contains(t, got, "SELECT cpf FROM users")
				assert.Contains(t, got, "cpf="+redactor.Sentinel)
				assert.NotContains(t, got, "123")
			},
		},
		{
			name:  "no PII keys",
			input: "SELECT name FROM users WHERE id=1",
			check: func(t *testing.T, got string) {
				assert.Equal(t, "SELECT name FROM users WHERE id=1", got)
			},
		},
		{
			name:  "multiple PII keys in one statement",
			input: "UPDATE t SET cpf='123', password='abc' WHERE id=1",
			check: func(t *testing.T, got string) {
				assert.NotContains(t, got, "123")
				assert.NotContains(t, got, "abc")
				assert.Contains(t, got, "cpf="+redactor.Sentinel)
				assert.Contains(t, got, "password="+redactor.Sentinel)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := dl.RedactStatement(tc.input)
			tc.check(t, got)
		})
	}
}

func TestRedactStatement_EmptyDenylist(t *testing.T) {
	dl := redactor.MustCompile()
	input := "SELECT password FROM users WHERE password='abc'"
	assert.Equal(t, input, dl.RedactStatement(input))
}

// ---- RedactAttrs ------------------------------------------------------------

func TestRedactAttrs(t *testing.T) {
	dl := redactor.MustCompile("password", "token")

	attrs := []kv{
		{"username", "alice"},
		{"password", "s3cr3t"},
		{"token", "abc123"},
		{"request_id", "req-42"},
	}

	got := redactor.RedactAttrs(dl, attrs,
		func(a kv) string { return a.key },
		func(a kv) kv { return kv{a.key, redactor.Sentinel} },
	)

	require.Len(t, got, 4)
	assert.Equal(t, "alice", got[0].val)
	assert.Equal(t, redactor.Sentinel, got[1].val)
	assert.Equal(t, redactor.Sentinel, got[2].val)
	assert.Equal(t, "req-42", got[3].val)
}

// ---- DefaultDenylist coverage -----------------------------------------------

func TestDefaultDenylist_AllKeysPresent(t *testing.T) {
	dl := redactor.DefaultDenylist

	required := []string{
		// PCI
		"pan", "card_number", "cardnumber", "cvv", "cvc",
		"password", "passwd", "secret", "token",
		"access_token", "refresh_token", "authorization",
		// BR
		"cpf", "cnpj", "rg",
		// Contact
		"email", "phone", "telefone", "address", "endereco", "zipcode", "cep",
		// Banking
		"account_number", "agency", "bank_account",
		"pix_key", "pix_chave", "full_name", "holder_name", "nome_completo",
	}

	for _, k := range required {
		assert.True(t, dl.Contains(k), "expected key %q to be in DefaultDenylist", k)
	}
}

func TestDefaultDenylist_SentinelIsOnlySubstitute(t *testing.T) {
	dl := redactor.DefaultDenylist

	src := map[string]any{
		"cpf":         "123.456.789-00",
		"password":    "secret",
		"card_number": "4111111111111111",
		"name":        "Alice",
	}

	got := dl.RedactMap(src)

	for k, v := range got {
		if dl.Contains(k) {
			assert.Equal(t, redactor.Sentinel, v,
				"key %q should be replaced only by Sentinel", k)
		}
	}
}

// ---- Benchmark --------------------------------------------------------------

func BenchmarkRedactMap(b *testing.B) {
	dl := redactor.DefaultDenylist

	input := map[string]any{
		"name":        "Alice",
		"card_number": "4111111111111111",
		"amount":      10000,
		"cpf":         "123.456.789-00",
		"address": map[string]any{
			"street":  "Main St",
			"zipcode": "01310-100",
		},
		"transactions": []any{
			map[string]any{"id": "t1", "card_number": "4111111111111111"},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		dl.RedactMap(input)
	}
}
