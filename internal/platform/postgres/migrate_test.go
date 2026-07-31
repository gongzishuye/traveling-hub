package postgres

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveMigrationDirectoryFindsParent(t *testing.T) {
	root := t.TempDir()
	migrations := filepath.Join(root, "migrations")
	if err := os.Mkdir(migrations, 0o755); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "tests", "integration")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(nested); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })
	got, err := resolveMigrationDirectory("migrations")
	if err != nil {
		t.Fatal(err)
	}
	if got != migrations {
		t.Fatalf("resolved migration directory = %q, want %q", got, migrations)
	}
}
