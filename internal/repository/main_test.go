//go:build integration

// Package repository_test holds the integration tests for the Postgres
// adapters. They are gated behind the `integration` build tag, so the default
// `go test ./...` (which has no database) skips them entirely. Run them with:
//
//	make test-integration
//	# or, against an already-running database:
//	TEST_DATABASE_URL=postgres://... go test -tags=integration ./internal/repository/...
package repository_test

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/Yasin4261/food-delivery/database"
)

// testDB is the shared connection used by every integration test. It is set up
// once in TestMain.
var testDB *sql.DB

func TestMain(m *testing.M) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		// Nothing to test against — treat as a clean skip so that running the
		// tagged suite without a database does not fail the build.
		log.Println("TEST_DATABASE_URL not set; skipping repository integration tests")
		return
	}

	conn, err := database.NewConnection(url)
	if err != nil {
		log.Fatalf("integration: connect to %s: %v", url, err)
	}
	defer conn.Close()
	testDB = conn.DB

	// Apply migrations against the test database (also exercises the migration
	// files and the golang-migrate runner).
	if err := database.RunMigrations(testDB, "../../migrations"); err != nil {
		log.Fatalf("integration: run migrations: %v", err)
	}

	os.Exit(m.Run())
}
