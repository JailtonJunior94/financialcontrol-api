package migration

import "log/slog"

// Version, Commit, and BuildDate are injected at build time via -ldflags.
var (
	Version   = "dev"
	Commit    = "dev"
	BuildDate = "dev"
)

// BuildInfoFields returns slog.Attr entries for the binary build metadata.
func BuildInfoFields() []slog.Attr {
	return []slog.Attr{
		slog.String("binary.version", Version),
		slog.String("binary.commit", Commit),
		slog.String("binary.build_date", BuildDate),
	}
}
