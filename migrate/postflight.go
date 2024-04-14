package migrate

import (
	"errors"
	"fmt"

	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

var (
	ErrPostFlight            = errors.New("postflight_error")
	ErrPostFlightMissingDB   = fmt.Errorf("%w - missing database in target cloudsql instance", ErrPostFlight)
	ErrPostFlightMissingUser = fmt.Errorf("%w - missing user in target cloudsql instance", ErrPostFlight)
)

type PostFlight interface {
	Name() string
	Prepopulate(op *Op) error
	Check(op *Op) error
}

type PostFlightSQLAdmin struct {
	users     []*sqladmin.User
	databases []*sqladmin.Database
}

func (p *PostFlightSQLAdmin) Name() string {
	return "sql_admin_api"
}

func (p *PostFlightSQLAdmin) Prepopulate(op *Op) error {
	dbs, err := op.Svc.Databases.List(op.Src.Project, op.Src.Name).Do()
	if err != nil {
		return err
	}
	p.databases = dbs.Items

	users, err := op.Svc.Users.List(op.Src.Project, op.Src.Name).Do()
	if err != nil {
		return err
	}
	p.users = users.Items

	return nil
}

func (p *PostFlightSQLAdmin) Check(op *Op) error {
	dbs, err := op.Svc.Databases.List(op.Dst.Project, op.Dst.Name).Do()
	if err != nil {
		return err
	}
	for _, ref := range p.databases {
		found := false
		for _, db := range dbs.Items {
			if ref.Name == db.Name && ref.Charset == db.Charset && ref.Collation == db.Collation {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("%w: %q", ErrPostFlightMissingDB, ref.Name)
		}
	}

	users, err := op.Svc.Users.List(op.Dst.Project, op.Dst.Name).Do()
	if err != nil {
		return err
	}
	for _, ref := range p.users {
		found := false
		for _, user := range users.Items {
			if ref.Name == user.Name && ref.Type == user.Type {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("%w: %q", ErrPostFlightMissingUser, ref.Name)
		}
	}

	return nil
}
