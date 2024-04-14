package migrate_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ichbinfrog/cloudsqlmigrate/migrate"
	"google.golang.org/api/option"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

type testcase struct {
	srcProject, srcInstance, dstProject, dstInstance string
	expected                                         error
	srcDBs, dstDBs                                   []*sqladmin.Database
	srcUsers, dstUsers                               []*sqladmin.User
}

func mockSQLAdmin(tc testcase) *httptest.Server {
	mux := http.NewServeMux()
	mux.Handle("GET /sql/v1beta4/projects/{project}/instances/{instance}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(&sqladmin.DatabaseInstance{
			Name:    r.PathValue("instance"),
			Project: r.PathValue("project"),
		})
	}))

	mux.Handle("GET /sql/v1beta4/projects/{project}/instances/{instance}/databases", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.PathValue("instance") {
		case tc.srcInstance:
			_ = json.NewEncoder(w).Encode(&sqladmin.DatabasesListResponse{
				Items: tc.srcDBs,
			})
		case tc.dstInstance:
			_ = json.NewEncoder(w).Encode(&sqladmin.DatabasesListResponse{
				Items: tc.dstDBs,
			})
		}
	}))

	mux.Handle("GET /sql/v1beta4/projects/{project}/instances/{instance}/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.PathValue("instance") {
		case tc.srcInstance:
			_ = json.NewEncoder(w).Encode(&sqladmin.UsersListResponse{
				Items: tc.srcUsers,
			})
		case tc.dstInstance:
			_ = json.NewEncoder(w).Encode(&sqladmin.UsersListResponse{
				Items: tc.dstUsers,
			})
		}
	}))

	return httptest.NewUnstartedServer(mux)
}

func TestPostFlightSQLAdmin(t *testing.T) {
	for name, spec := range map[string]testcase{
		"empty": {
			srcProject:  "prj-src",
			srcInstance: "sql-src",
			dstProject:  "prj-dst",
			dstInstance: "sql-dst",
			expected:    nil,
		},
		"mismatch_db_count_fail": {
			srcProject:  "prj-src",
			srcInstance: "sql-src",
			srcDBs: []*sqladmin.Database{
				{
					Name:      "holla",
					Charset:   "UTF8",
					Collation: "en_US.UTF8",
				},
			},

			dstProject:  "prj-dst",
			dstInstance: "sql-dst",
			dstDBs:      []*sqladmin.Database{},
			expected:    migrate.ErrPostFlightMissingDB,
		},
		"mismatch_db_count_ok": {
			srcProject:  "prj-src",
			srcInstance: "sql-src",
			srcDBs: []*sqladmin.Database{
				{
					Name:      "holla",
					Charset:   "UTF8",
					Collation: "en_US.UTF8",
				},
			},

			dstProject:  "prj-dst",
			dstInstance: "sql-dst",
			dstDBs: []*sqladmin.Database{
				{
					Name:      "holla",
					Charset:   "UTF8",
					Collation: "en_US.UTF8",
				},
				{
					Name:      "quetal",
					Charset:   "UTF8",
					Collation: "en_US.UTF8",
				},
			},
			expected: nil,
		},
		"mismatch_user_count_fail": {
			srcProject:  "prj-src",
			srcInstance: "sql-src",
			srcUsers: []*sqladmin.User{
				{Name: "holla", Type: "BUILT_IN"},
			},

			dstProject:  "prj-dst",
			dstInstance: "sql-dst",
			dstUsers:    []*sqladmin.User{},

			expected: migrate.ErrPostFlightMissingUser,
		},
		"mismatch_user_count_ok": {
			srcProject:  "prj-src",
			srcInstance: "sql-src",
			srcUsers: []*sqladmin.User{
				{Name: "holla", Type: "BUILT_IN"},
			},

			dstProject:  "prj-dst",
			dstInstance: "sql-dst",
			dstUsers: []*sqladmin.User{
				{Name: "como", Type: "BUILT_IN"},
				{Name: "holla", Type: "BUILT_IN"},
			},
			expected: nil,
		},
	} {
		spec := spec
		t.Run(name, func(t *testing.T) {
			srv := mockSQLAdmin(spec)
			defer srv.Close()
			srv.Start()

			ctx := context.Background()
			op, err := migrate.NewOp(ctx, spec.srcProject, spec.srcInstance, spec.dstProject, spec.dstInstance, option.WithoutAuthentication(), option.WithEndpoint(srv.URL))
			if err != nil {
				t.Fatal(err)
			}

			checks := []migrate.PostFlight{
				&migrate.PostFlightSQLAdmin{},
			}
			if err := op.Prepopulate(ctx, checks); len(err) > 0 {
				t.Fatal(err)
			}

			for _, c := range checks {
				err := c.Check(op)
				if spec.expected == nil && err != nil {
					t.Fatalf("Postflight check %s should not have failed, instead got %v", c.Name(), err)
				}

				if !errors.Is(err, spec.expected) {
					t.Fatalf("Postflight check %s should have failed, instead got %v", c.Name(), err)
				}
			}

		})
	}

}
