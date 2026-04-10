package test

import (
	"database/sql"
	_ "embed"
	"testing"
)

// TruncateMenuTables truncates tables for products and product categories.
func TruncateMenuTables(t testing.TB, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`TRUNCATE TABLE products, product_categories RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("error truncating menu tables: %v", err)
	}
}

//go:embed testdata/seeds/menu.sql
var menuSeed string

// SeedMenuTables seeds table for products and product categories.
// It automatically truncates tables before inserting new rows.
func SeedMenuTables(t testing.TB, db *sql.DB) {
	t.Helper()

	TruncateMenuTables(t, db)
	if _, err := db.Exec(menuSeed); err != nil {
		t.Fatalf("error seeding menu tables: %v", err)
	}
}

// TruncateOrderTables truncates tables for ordered products and order sessions.
func TruncateOrderTables(t testing.TB, db *sql.DB) {
	t.Helper()

	if _, err := db.Exec(`TRUNCATE TABLE ordered_products, order_sessions RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("error truncating order tables: %v", err)
	}
}

//go:embed testdata/seeds/orders.sql
var orderSeed string

// SeedOrderTables seed the table order sessions and ordered products.
// It automatically truncates tables before inserting new rows.
func SeedOrderTables(t testing.TB, db *sql.DB) {
	t.Helper()

	TruncateOrderTables(t, db)
	if _, err := db.Exec(orderSeed); err != nil {
		t.Fatalf("error seeding order tables: %v", err)
	}
}

// ConnectToDb establishes and checks connection to postgres database.
func ConnectToDb(url string) *sql.DB {
	db, err := sql.Open("postgres", url)
	if err != nil {
		panic(err)
	}
	return db
}
