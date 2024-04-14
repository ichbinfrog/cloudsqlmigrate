package migrate_test

import (
	"errors"
	"testing"

	"github.com/ichbinfrog/cloudsqlmigrate/migrate"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

func TestPreflightVersion(t *testing.T) {
	type testcase struct {
		op       *migrate.Op
		expected error
	}

	for name, spec := range map[string]testcase{
		"empty": {
			op: &migrate.Op{
				Src: &sqladmin.DatabaseInstance{
					DatabaseVersion: "",
				},
				Dst: &sqladmin.DatabaseInstance{
					DatabaseVersion: "POSTGRES_15",
				},
			},
			expected: migrate.ErrPreflightVersionMismatch,
		},
		"equal": {
			op: &migrate.Op{
				Src: &sqladmin.DatabaseInstance{
					DatabaseVersion: "POSTGRES_15",
				},
				Dst: &sqladmin.DatabaseInstance{
					DatabaseVersion: "POSTGRES_15",
				},
			},
			expected: nil,
		},
		"mismatch": {
			op: &migrate.Op{
				Src: &sqladmin.DatabaseInstance{
					DatabaseVersion: "POSTGRES_15",
				},
				Dst: &sqladmin.DatabaseInstance{
					DatabaseVersion: "MYSQL_5_6",
				},
			},
			expected: migrate.ErrPreflightVersionMismatch,
		},
	} {
		spec := spec
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			preflight := migrate.PreflightVersion{}
			err := preflight.Check(spec.op)

			if spec.expected == nil && err != nil {
				t.Fatalf("Preflight check should not have failed, src_database=%s, dst_database=%s\n", spec.op.Src.DatabaseVersion, spec.op.Dst.DatabaseVersion)
			}

			if !errors.Is(err, spec.expected) {
				t.Fatalf("Preflight check should have failed, src_database=%s, dst_database=%s\n", spec.op.Src.DatabaseVersion, spec.op.Dst.DatabaseVersion)
			}
		})
	}
}

func TestPreflightStatus(t *testing.T) {
	type testcase struct {
		op       *migrate.Op
		expected error
	}

	for name, spec := range map[string]testcase{
		"empty": {
			op: &migrate.Op{
				Src: &sqladmin.DatabaseInstance{
					State: "",
				},
				Dst: &sqladmin.DatabaseInstance{
					State: "SQL_INSTANCE_STATE_UNSPECIFIED",
				},
			},
			expected: migrate.ErrPreflightStatus,
		},
		"equal": {
			op: &migrate.Op{
				Src: &sqladmin.DatabaseInstance{
					State: "RUNNABLE",
				},
				Dst: &sqladmin.DatabaseInstance{
					State: "RUNNABLE",
				},
			},
			expected: nil,
		},
		"mismatch": {
			op: &migrate.Op{
				Src: &sqladmin.DatabaseInstance{
					DatabaseVersion: "FAILED",
				},
				Dst: &sqladmin.DatabaseInstance{
					DatabaseVersion: "RUNNABLE",
				},
			},
			expected: migrate.ErrPreflightStatus,
		},
	} {
		spec := spec
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			preflight := migrate.PreflightStatus{}
			err := preflight.Check(spec.op)

			if spec.expected == nil && err != nil {
				t.Fatalf("Preflight check should not have failed, src_database=%s, dst_database=%s\n", spec.op.Src.DatabaseVersion, spec.op.Dst.DatabaseVersion)
			}

			if !errors.Is(err, spec.expected) {
				t.Fatalf("Preflight check should have failed, src_database=%s, dst_database=%s\n", spec.op.Src.DatabaseVersion, spec.op.Dst.DatabaseVersion)
			}
		})
	}
}
