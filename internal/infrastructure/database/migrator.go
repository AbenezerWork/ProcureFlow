package database

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	dbmigrations "github.com/AbenezerWork/ProcureFlow/db/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

const ensureSchemaMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
)
`

type Migrator struct {
	pool *pgxpool.Pool
	fs   fs.FS
}

type MigrationResult struct {
	Applied        []migrationFile
	CurrentVersion int64
	TargetVersion  int64
}

type migrationFile struct {
	version int64
	name    string
	sql     string
}

func NewMigrator(pool *pgxpool.Pool) *Migrator {
	return &Migrator{
		pool: pool,
		fs:   dbmigrations.Files,
	}
}

func (m *Migrator) Up(ctx context.Context) (MigrationResult, error) {
	if err := m.ensureMigrationsTable(ctx); err != nil {
		return MigrationResult{}, err
	}

	migrations, err := collectUpMigrations(m.fs)
	if err != nil {
		return MigrationResult{}, err
	}

	appliedVersions, err := m.appliedVersions(ctx)
	if err != nil {
		return MigrationResult{}, err
	}

	result := MigrationResult{
		CurrentVersion: highestAppliedVersion(appliedVersions),
		TargetVersion:  latestMigrationVersion(migrations),
	}

	for _, migration := range migrations {
		if _, ok := appliedVersions[migration.version]; ok {
			continue
		}

		if err := m.applyMigration(ctx, migration); err != nil {
			return MigrationResult{}, err
		}

		result.Applied = append(result.Applied, migration)
		result.CurrentVersion = migration.version
	}

	return result, nil
}

func (m *Migrator) Version(ctx context.Context) (int64, error) {
	if err := m.ensureMigrationsTable(ctx); err != nil {
		return 0, err
	}

	var version int64
	if err := m.pool.QueryRow(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&version); err != nil {
		return 0, fmt.Errorf("query current schema version: %w", err)
	}

	return version, nil
}

func (m *Migrator) ensureMigrationsTable(ctx context.Context) error {
	if _, err := m.pool.Exec(ctx, ensureSchemaMigrationsTable); err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}

	return nil
}

func (m *Migrator) appliedVersions(ctx context.Context) (map[int64]struct{}, error) {
	rows, err := m.pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("load applied schema versions: %w", err)
	}
	defer rows.Close()

	versions := make(map[int64]struct{})
	for rows.Next() {
		var version int64
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan applied schema version: %w", err)
		}

		versions[version] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied schema versions: %w", err)
	}

	return versions, nil
}

func (m *Migrator) applyMigration(ctx context.Context, migration migrationFile) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration %s transaction: %w", migration.name, err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	if _, err := tx.Exec(ctx, stripTransactionWrapper(migration.sql)); err != nil {
		return fmt.Errorf("apply migration %s: %w", migration.name, err)
	}

	if _, err := tx.Exec(
		ctx,
		`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`,
		migration.version,
		migration.name,
	); err != nil {
		return fmt.Errorf("record migration %s: %w", migration.name, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration %s: %w", migration.name, err)
	}

	committed = true
	return nil
}

func collectUpMigrations(fsys fs.FS) ([]migrationFile, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("read migrations directory: %w", err)
	}

	migrations := make([]migrationFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}

		version, name, err := parseMigrationFileName(entry.Name())
		if err != nil {
			return nil, err
		}

		contents, err := fs.ReadFile(fsys, entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		migrations = append(migrations, migrationFile{
			version: version,
			name:    name,
			sql:     string(contents),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	return migrations, nil
}

func highestAppliedVersion(versions map[int64]struct{}) int64 {
	var highest int64
	for version := range versions {
		if version > highest {
			highest = version
		}
	}

	return highest
}

func latestMigrationVersion(migrations []migrationFile) int64 {
	var latest int64
	for _, migration := range migrations {
		if migration.version > latest {
			latest = migration.version
		}
	}

	return latest
}

func parseMigrationFileName(name string) (int64, string, error) {
	base := filepath.Base(name)
	parts := strings.SplitN(base, "_", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid migration name %q: expected <version>_<name>.up.sql", name)
	}

	version, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("parse migration version from %q: %w", name, err)
	}

	return version, base, nil
}

func stripTransactionWrapper(sql string) string {
	lines := strings.Split(sql, "\n")
	start := 0
	end := len(lines)

	for start < end && strings.TrimSpace(lines[start]) == "" {
		start++
	}

	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}

	if start < end && strings.EqualFold(strings.TrimSpace(lines[start]), "BEGIN;") {
		start++
	}

	if end > start && strings.EqualFold(strings.TrimSpace(lines[end-1]), "COMMIT;") {
		end--
	}

	return strings.Join(lines[start:end], "\n")
}
