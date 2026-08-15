package storage

import (
	"database/sql"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"storycode/migrations"
)

var migrationName = regexp.MustCompile(`^(\d+)_.+\.sql$`)

// Migrate applies bundled versioned SQL migrations in order.
//
//	err := storage.Migrate(db)
func Migrate(db *sql.DB) error {
	return MigrateFS(db, migrations.SQL)
}

// MigrateFS applies *.sql files from files. Each file runs in one transaction.
//
//	err := storage.MigrateFS(db, fstest.MapFS{"0001_ok.sql": {Data: []byte(sql)}})
func MigrateFS(db *sql.DB, files fs.FS) error {
	if err := ensureMigrationsTable(db); err != nil {
		return err
	}
	pending, err := pendingMigrations(db, files)
	if err != nil {
		return err
	}
	return applyAll(db, pending)
}

func applyAll(db *sql.DB, pending []migrationFile) error {
	for _, item := range pending {
		if err := applyMigration(db, item); err != nil {
			return err
		}
	}
	return nil
}

type migrationFile struct {
	name    string
	version int
	body    string
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY NOT NULL,
    applied_at TEXT NOT NULL
)`)
	if err != nil {
		return fmt.Errorf("cannot create schema_migrations, expected a writable sqlite database: %w", err)
	}
	return nil
}

func pendingMigrations(db *sql.DB, files fs.FS) ([]migrationFile, error) {
	all, err := listMigrations(files)
	if err != nil {
		return nil, err
	}
	applied, err := appliedVersions(db)
	if err != nil {
		return nil, err
	}
	var pending []migrationFile
	for _, item := range all {
		if applied[item.version] {
			continue
		}
		pending = append(pending, item)
	}
	return pending, nil
}

func listMigrations(files fs.FS) ([]migrationFile, error) {
	entries, err := fs.ReadDir(files, ".")
	if err != nil {
		return nil, fmt.Errorf("cannot read migrations directory, expected a filesystem of *.sql files: %w", err)
	}
	return parseMigrationEntries(files, entries)
}

func parseMigrationEntries(files fs.FS, entries []fs.DirEntry) ([]migrationFile, error) {
	var out []migrationFile
	seen := map[int]string{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		item, err := loadMigration(files, entry.Name())
		if err != nil {
			return nil, err
		}
		if err := rememberVersion(seen, item); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].version < out[j].version })
	return out, nil
}

func rememberVersion(seen map[int]string, item migrationFile) error {
	if prev, ok := seen[item.version]; ok {
		return fmt.Errorf("duplicate migration version %d in %q and %q, expected unique numeric prefixes", item.version, prev, item.name)
	}
	seen[item.version] = item.name
	return nil
}

func loadMigration(files fs.FS, name string) (migrationFile, error) {
	version, err := versionFromName(name)
	if err != nil {
		return migrationFile{}, err
	}
	body, err := fs.ReadFile(files, name)
	if err != nil {
		return migrationFile{}, fmt.Errorf("cannot read migration %q, expected a readable SQL file: %w", name, err)
	}
	return migrationFile{name: name, version: version, body: string(body)}, nil
}

func versionFromName(name string) (int, error) {
	match := migrationName.FindStringSubmatch(name)
	if match == nil {
		return 0, fmt.Errorf("invalid migration name %q, expected shape NNNN_description.sql", name)
	}
	version, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, fmt.Errorf("invalid migration version in %q, expected an integer prefix: %w", name, err)
	}
	return version, nil
}

func appliedVersions(db *sql.DB) (map[int]bool, error) {
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("cannot read schema_migrations, expected columns version INTEGER: %w", err)
	}
	defer rows.Close()
	out := map[int]bool{}
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("cannot scan schema_migrations version, expected INTEGER: %w", err)
		}
		out[version] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate schema_migrations: %w", err)
	}
	return out, nil
}

func applyMigration(db *sql.DB, item migrationFile) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("cannot begin migration %q, expected a transactional sqlite connection: %w", item.name, err)
	}
	if err := execStatements(tx, item); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := recordMigration(tx, item); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("cannot commit migration %q, expected a successful transaction: %w", item.name, err)
	}
	return nil
}

func recordMigration(tx *sql.Tx, item migrationFile) error {
	appliedAt := time.Now().UTC().Format(time.RFC3339)
	_, err := tx.Exec(
		`INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`,
		item.version,
		appliedAt,
	)
	if err != nil {
		return fmt.Errorf("cannot record migration %q version %d, expected unique INTEGER version: %w", item.name, item.version, err)
	}
	return nil
}

func execStatements(tx *sql.Tx, item migrationFile) error {
	for _, stmt := range sqlStatements(item.body) {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("cannot apply migration %q statement %q, expected valid SQLite SQL: %w", item.name, stmt, err)
		}
	}
	return nil
}

func sqlStatements(body string) []string {
	var out []string
	for _, part := range strings.Split(body, ";") {
		stmt := strings.TrimSpace(part)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		out = append(out, stmt)
	}
	return out
}
