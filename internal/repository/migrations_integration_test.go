//go:build integration

package repository_test

import (
	"testing"

	"github.com/Yasin4261/food-delivery/database"
)

// TestMigrations_SchemaApplied confirms every expected table exists after the
// migrations run in TestMain.
func TestMigrations_SchemaApplied(t *testing.T) {
	want := []string{"users", "chefs", "menus", "menu_items", "orders", "order_items", "favorites"}
	for _, table := range want {
		var exists bool
		err := testDB.QueryRow(
			`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)`, table,
		).Scan(&exists)
		if err != nil {
			t.Fatalf("query table %q: %v", table, err)
		}
		if !exists {
			t.Errorf("expected table %q to exist after migrations", table)
		}
	}
}

// TestMigrations_Idempotent confirms re-running migrations on an up-to-date
// database is a clean no-op (migrate.ErrNoChange is handled).
func TestMigrations_Idempotent(t *testing.T) {
	if err := database.RunMigrations(testDB, "../../migrations"); err != nil {
		t.Errorf("re-running migrations should be a no-op, got: %v", err)
	}
}
