//go:build e2e

package main_test

import (
	"context"
	"os"
	"testing"

	"github.com/ichbinfrog/cloudsqlmigrate/migrate"
)

func TestMigrationPostgres(t *testing.T) {
	ctx := context.Background()

	srcProject := os.Getenv("MIGRATE_SRC_PROJECT")
	srcInstance := os.Getenv("MIGRATE_SRC_INSTANCE")

	dstProject := os.Getenv("MIGRATE_DST_PROJECT")
	dstInstance := os.Getenv("MIGRATE_DST_INSTANCE")

	op, err := migrate.NewOp(ctx, srcProject, srcInstance, dstProject, dstInstance)
	if err != nil {
		t.Fatal(err)
	}

	if errs := op.Run(ctx,
		[]migrate.Preflight{
			migrate.PreflightVersion{},
			migrate.PreflightStatus{},
		},
		[]migrate.PostFlight{
			&migrate.PostFlightSQLAdmin{},
		},
	); len(errs) > 0 {
		t.Fatalf("checks failed: %v", errs)
	}
}
