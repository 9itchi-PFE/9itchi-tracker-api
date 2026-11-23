//go:build integration

package repositories

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5"
)

var (
	adminURL string
	dbSeq    atomic.Int64
)

// TestMain boots one embedded PostgreSQL for the package (no Docker needed). Run: go test -tags integration ./repositories/...
func TestMain(m *testing.M) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()

	dataDir, _ := os.MkdirTemp("", "itchi-pg-data-")
	runDir, _ := os.MkdirTemp("", "itchi-pg-run-")
	pg := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Version(embeddedpostgres.V16).Port(uint32(port)).DataPath(dataDir).RuntimePath(runDir))
	if err := pg.Start(); err != nil {
		fmt.Println("embedded postgres:", err)
		os.Exit(1)
	}
	adminURL = fmt.Sprintf("postgres://postgres:postgres@localhost:%d/postgres?sslmode=disable", port)
	code := m.Run()
	_ = pg.Stop()
	os.RemoveAll(dataDir)
	os.RemoveAll(runDir)
	os.Exit(code)
}

// newTestStore creates an isolated, migrated database for one test.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	ctx := context.Background()
	name := fmt.Sprintf("t_%d", dbSeq.Add(1))
	admin, err := pgx.Connect(ctx, adminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+name); err != nil {
		t.Fatal(err)
	}
	url := fmt.Sprintf("%s/%s?sslmode=disable", adminURL[:len(adminURL)-len("/postgres?sslmode=disable")], name)
	if err := RunMigrations(ctx, url); err != nil {
		t.Fatal(err)
	}
	store, err := NewStore(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(store.Close)
	return store
}
