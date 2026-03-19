package database

import (
	"testing"
	"testing/fstest"
)

func TestCollectUpMigrationsSortsAndFiltersFiles(t *testing.T) {
	t.Parallel()

	migrations, err := collectUpMigrations(fstest.MapFS{
		"000002_add_index.up.sql":   {Data: []byte("SELECT 2;")},
		"000001_init_schema.up.sql": {Data: []byte("SELECT 1;")},
		"000001_init_schema.down.sql": {
			Data: []byte("SELECT 0;"),
		},
		"notes.txt": {Data: []byte("ignore me")},
	})
	if err != nil {
		t.Fatalf("collectUpMigrations returned error: %v", err)
	}

	if len(migrations) != 2 {
		t.Fatalf("expected 2 up migrations, got %d", len(migrations))
	}

	if migrations[0].version != 1 || migrations[0].name != "000001_init_schema.up.sql" {
		t.Fatalf("unexpected first migration: %+v", migrations[0])
	}

	if migrations[1].version != 2 || migrations[1].name != "000002_add_index.up.sql" {
		t.Fatalf("unexpected second migration: %+v", migrations[1])
	}
}

func TestStripTransactionWrapper(t *testing.T) {
	t.Parallel()

	sql := "\nBEGIN;\nCREATE TABLE example (id INT);\nCOMMIT;\n"
	got := stripTransactionWrapper(sql)
	want := "CREATE TABLE example (id INT);"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
