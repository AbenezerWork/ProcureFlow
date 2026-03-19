package migrations

import "embed"

// Files exposes the SQL migrations for the embedded migration runner.
//
//go:embed *.sql
var Files embed.FS
